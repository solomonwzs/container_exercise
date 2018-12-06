package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"

	"github.com/solomonwzs/goxutil/logger"
)

var devUniqID uint32 = 0

type NetworkBuilder interface {
	BuildNetwork() error
	ReleaseNetwork() error
	SetupNetwork() error
}

func ParserNetworkBuilders(cPid int, conf Configuration) []NetworkBuilder {
	nb := make([]NetworkBuilder, 0)
	for _, netConf := range conf.Network.Interfaces {
		switch netConf.Type {
		case "bridge":
			nb = append(nb, NewCNIBridge(cPid, netConf))
		case "macvlan":
			builder, err := NewCNIMacvlan(cPid, netConf)
			if err != nil {
				logger.Error(err)
			} else {
				nb = append(nb, builder)
			}
		default:
		}
	}
	return nb
}

func Ipv42Dec(ip net.IP) (uint32, error) {
	if tmp := ip.To4(); tmp != nil {
		return binary.BigEndian.Uint32(tmp), nil
	}
	return 0, errors.New("not IPv4 address")
}

func Dec2Ipv4(dec uint32) net.IP {
	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, dec)
	return ip
}

func MaskValidBits(mask net.IP) int {
	if tmp := mask.To4(); tmp != nil {
		dec := binary.LittleEndian.Uint32(tmp)
		n := 0
		for ; dec&1 != 0; dec >>= 1 {
			n += 1
		}
		return n
	}
	return 0
}

func AddNetworkRoutes(routes []CNetworkRoute) (err error) {
	for _, route := range routes {
		mask := net.ParseIP(route.Mask)
		maskValidBits := MaskValidBits(mask)
		dest := fmt.Sprintf("%s/%d", route.Dest, maskValidBits)

		SystemCmd("ip", "route", "add", dest, "via", route.Gateway)
	}
	return
}

type CNIBridge struct {
	BridgeAddr string
	BridgeName string
	Name       string
	Pid        string
	RuleSrc    string
	VethA      string
	VethAddr   string
	VethB      string
}

func NewCNIBridge(cPid int, conf CNetworkInterface) *CNIBridge {
	devid := atomic.AddUint32(&devUniqID, 1)
	pid := strconv.Itoa(cPid)

	bridgeName := fmt.Sprintf("bridge%s-%d", pid, devid)
	vethA := fmt.Sprintf("veth%s-%d", pid, devid)
	vethB := fmt.Sprintf("veth%sx-%d", pid, devid)

	mask := net.ParseIP(conf.Mask)
	maskDec, _ := Ipv42Dec(mask)
	maskValidBits := MaskValidBits(mask)
	ip := net.ParseIP(conf.IP)
	ipDec, _ := Ipv42Dec(ip)

	bridgeIP := Dec2Ipv4(ipDec&maskDec | 1)
	bridgeAddr := fmt.Sprintf("%s/%d", bridgeIP, maskValidBits)
	vethAddr := fmt.Sprintf("%s/%d", ip, maskValidBits)

	ruleSrc := fmt.Sprintf("%s/%d", Dec2Ipv4(ipDec&maskDec), maskValidBits)

	return &CNIBridge{
		BridgeAddr: bridgeAddr,
		BridgeName: bridgeName,
		Name:       conf.Name,
		Pid:        pid,
		RuleSrc:    ruleSrc,
		VethA:      vethA,
		VethAddr:   vethAddr,
		VethB:      vethB,
	}
}

func (conf *CNIBridge) BuildNetwork() (err error) {
	// create bridge
	SystemCmd("ip", "link",
		"add", conf.BridgeName, "type", "bridge")
	SystemCmd("ip", "addr",
		"add", conf.BridgeAddr, "brd", "+", "dev", conf.BridgeName)
	SystemCmd("ip", "link",
		"set", conf.BridgeName, "up")

	// create a pair of veths
	SystemCmd("ip", "link",
		"add", conf.VethA, "type", "veth", "peer", "name", conf.VethB)
	SystemCmd("ip", "link",
		"set", conf.VethB, "netns", conf.Pid)

	// set veth to bridge
	SystemCmd("ip", "link",
		"set", conf.VethA, "master", conf.BridgeName)
	SystemCmd("ip", "link",
		"set", conf.VethA, "up")

	// add iptables rule
	SystemCmd("sysctl", "-w", "net.ipv4.ip_forward=1")
	SystemCmd("iptables",
		"-t", "nat",
		"-A", "POSTROUTING",
		"-s", conf.RuleSrc,
		"-j", "MASQUERADE")

	return
}

func (conf *CNIBridge) ReleaseNetwork() (err error) {
	SystemCmd("iptables",
		"-t", "nat",
		"-D", "POSTROUTING",
		"-s", conf.RuleSrc,
		"-j", "MASQUERADE")

	SystemCmd("ip", "link", "delete", conf.BridgeName, "type", "bridge")

	return
}

func (conf *CNIBridge) SetupNetwork() (err error) {
	SystemCmd("ip", "link", "set", "lo", "up")
	SystemCmd("ip", "link", "set", conf.VethB, "name", conf.Name)
	SystemCmd("ip", "link", "set", conf.Name, "up")
	SystemCmd("ip", "addr", "add", conf.VethAddr, "dev", conf.Name)
	return
}

type CNIMacvlan struct {
	HostInterface string
	Name          string
	VName         string
	Pid           string
	Mode          string
	IP            net.IP
	Mask          net.IP
}

func NewCNIMacvlan(cPid int, conf CNetworkInterface) (*CNIMacvlan, error) {
	devid := atomic.AddUint32(&devUniqID, 1)
	pid := strconv.Itoa(cPid)

	vname := fmt.Sprintf("macv%s-%d", pid, devid)

	var (
		ip   net.IP
		mask net.IP
	)

	if conf.IP != "" {
		mask = net.ParseIP(conf.Mask)
		ip = net.ParseIP(conf.IP)

		if ip == nil || mask == nil {
			return nil, fmt.Errorf("invalid ip (%s) or mask (%s)",
				conf.IP, conf.Mask)
		}
	}

	return &CNIMacvlan{
		HostInterface: conf.HostInterface,
		Mode:          conf.Mode,
		Pid:           pid,
		Name:          conf.Name,
		VName:         vname,
		IP:            ip,
		Mask:          mask,
	}, nil
}

func (conf *CNIMacvlan) getAddr() (addr string, err error) {
	if conf.IP != nil {
		maskValidBits := MaskValidBits(conf.Mask)
		addr = fmt.Sprintf("%s/%d", conf.IP, maskValidBits)
	}
	return
}

func (conf *CNIMacvlan) BuildNetwork() (err error) {
	SystemCmd("ip", "link",
		"add", "link", conf.HostInterface, "name", conf.VName,
		"type", "macvlan", "mode", conf.Mode)
	SystemCmd("ip", "link",
		"set", conf.VName, "netns", conf.Pid)
	return
}

func (conf *CNIMacvlan) ReleaseNetwork() (err error) { return }

func (conf *CNIMacvlan) SetupNetwork() (err error) {
	addr, err := conf.getAddr()
	if err != nil {
		return
	}

	SystemCmd("ip", "link", "set", "lo", "up")
	SystemCmd("ip", "link", "set", conf.VName, "name", conf.Name)
	SystemCmd("ip", "link", "set", conf.Name, "up")
	SystemCmd("ip", "addr", "add", addr, "dev", conf.Name)

	return
}
