package cnet

/*
#include "network.h"
*/
import "C"
import (
	"csys"
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/solomonwzs/goxutil/logger"
)

type CNetwork struct {
	Interfaces []CNetworkInterface `toml:"interface"`
	Routes     []CNetworkRoute     `toml:"route"`
}

type CNetworkInterface struct {
	HostInterface string `toml:"host_interface"`
	IP            string `toml:"ip"`
	Mask          string `toml:"mask"`
	Mode          string `toml:"mode"`
	Name          string `toml:"name"`
	ID            string `toml:"id"`
	Type          string `toml:"type"`
}

type CNetworkRoute struct {
	Dest    string `toml:"dest"`
	Gateway string `toml:"gateway"`
	Mask    string `toml:"mask"`
}

var devUniqID uint32 = 0

type NetworkBuilder interface {
	BuildNetwork() error
	ReleaseNetwork() error
	SetupNetwork() error
}

func ParserNetworkBuilders(cPid int, conf CNetwork) []NetworkBuilder {
	nb := make([]NetworkBuilder, 0)
	for _, netConf := range conf.Interfaces {
		switch netConf.Type {
		case "bridge":
			nb = append(nb, NewCNIBridge(cPid, netConf))
		case "vlan":
			builder, err := NewCNIVlan(cPid, netConf)
			if err != nil {
				logger.Error(err)
			} else {
				nb = append(nb, builder)
			}
		case "ipvlan":
			builder, err := NewCNIIPvlan(cPid, netConf)
			if err != nil {
				logger.Error(err)
			} else {
				nb = append(nb, builder)
			}
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

		csys.SystemCmd("ip", "route", "add", dest, "via", route.Gateway)
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

func (conf CNIBridge) BuildNetwork() (err error) {
	// create bridge
	csys.SystemCmd("ip", "link",
		"add", conf.BridgeName, "type", "bridge")
	csys.SystemCmd("ip", "addr",
		"add", conf.BridgeAddr, "brd", "+", "dev", conf.BridgeName)
	csys.SystemCmd("ip", "link",
		"set", conf.BridgeName, "up")

	// create a pair of veths
	csys.SystemCmd("ip", "link",
		"add", conf.VethA, "type", "veth", "peer", "name", conf.VethB)
	csys.SystemCmd("ip", "link",
		"set", conf.VethB, "netns", conf.Pid)

	// set veth to bridge
	csys.SystemCmd("ip", "link",
		"set", conf.VethA, "master", conf.BridgeName)
	csys.SystemCmd("ip", "link",
		"set", conf.VethA, "up")

	// add iptables rule
	csys.SystemCmd("sysctl", "-w", "net.ipv4.ip_forward=1")
	csys.SystemCmd("iptables",
		"-t", "nat",
		"-A", "POSTROUTING",
		"-s", conf.RuleSrc,
		"-j", "MASQUERADE")

	return
}

func (conf CNIBridge) ReleaseNetwork() (err error) {
	csys.SystemCmd("iptables",
		"-t", "nat",
		"-D", "POSTROUTING",
		"-s", conf.RuleSrc,
		"-j", "MASQUERADE")

	csys.SystemCmd("ip", "link", "delete", conf.BridgeName, "type", "bridge")

	return
}

func (conf CNIBridge) SetupNetwork() (err error) {
	csys.SystemCmd("ip", "link", "set", "lo", "up")
	csys.SystemCmd("ip", "link", "set", conf.VethB, "name", conf.Name)
	csys.SystemCmd("ip", "link", "set", conf.Name, "up")
	csys.SystemCmd("ip", "addr", "add", conf.VethAddr, "dev", conf.Name)
	return
}

type _CNIVlan struct {
	HostInterface string
	VName         string
	Name          string
	Pid           string
	IP            net.IP
	Mask          net.IP
}

func (conf _CNIVlan) getAddr() (addr string, err error) {
	var (
		interf *net.Interface
		reply  DHCPReply
	)
	if interf, err = net.InterfaceByName(conf.Name); err != nil {
		return
	}

	if conf.IP != nil {
		maskValidBits := MaskValidBits(conf.Mask)
		addr = fmt.Sprintf("%s/%d", conf.IP, maskValidBits)
	} else {
		if reply, err = DHCPApply(interf, 5*time.Second); err != nil {
			return
		}

		maskValidBits := MaskValidBits(reply.SubnetMask)
		addr = fmt.Sprintf("%s/%d", reply.ClientIP, maskValidBits)
	}
	return
}

func (conf _CNIVlan) ReleaseNetwork() (err error) { return }

func (conf _CNIVlan) SetupNetwork() (err error) {
	csys.SystemCmd("ip", "link", "set", "lo", "up")
	// csys.SystemCmd("ip", "link", "set", conf.VName, "name", conf.Name)
	C.net_rename(C.CString(conf.VName), C.CString(conf.Name))
	csys.SystemCmd("ip", "link", "set", conf.Name, "up")

	addr, err := conf.getAddr()
	if err != nil {
		logger.Error(err)
		return
	}
	csys.SystemCmd("ip", "addr", "add", addr, "dev", conf.Name)
	return
}

type CNIVlan struct {
	_CNIVlan
	ID string
}

func NewCNIVlan(cPid int, conf CNetworkInterface) (CNIVlan, error) {
	pid := strconv.Itoa(cPid)
	vname := fmt.Sprintf("vlan-%s", conf.ID)
	mask := net.ParseIP(conf.Mask)
	ip := net.ParseIP(conf.IP)

	if ip == nil || mask == nil {
		return CNIVlan{}, fmt.Errorf("invalid ip (%s) or mask (%s)",
			conf.IP, conf.Mask)
	}

	return CNIVlan{
		_CNIVlan: _CNIVlan{
			HostInterface: conf.HostInterface,
			Pid:           pid,
			VName:         vname,
			Name:          conf.Name,
			IP:            ip,
			Mask:          mask,
		},
		ID: conf.ID,
	}, nil
}

func (conf CNIVlan) BuildNetwork() (err error) {
	csys.SystemCmd("ip", "link",
		"add", "link", conf.HostInterface, "name", conf.VName,
		"type", "vlan", "id", conf.ID)
	csys.SystemCmd("ip", "link",
		"set", conf.VName, "netns", conf.Pid)
	return
}

type CNIIPvlan struct {
	_CNIVlan
	Mode  string
	_Pid  C.unsigned
	_Mode C.uint16_t
}

func NewCNIIPvlan(cPid int, conf CNetworkInterface) (CNIIPvlan, error) {
	var (
		devid uint32 = atomic.AddUint32(&devUniqID, 1)
		pid   string = strconv.Itoa(cPid)
		vname string = fmt.Sprintf("ipv%s-%d", pid, devid)
		ip    net.IP
		mask  net.IP
		mode  C.uint16_t
	)

	if conf.IP != "" {
		mask = net.ParseIP(conf.Mask)
		ip = net.ParseIP(conf.IP)

		if ip == nil || mask == nil {
			return CNIIPvlan{}, fmt.Errorf("invalid ip (%s) or mask (%s)",
				conf.IP, conf.Mask)
		}
	}

	switch conf.Mode {
	case "l2":
		mode = C.IPVLAN_MODE_L2
	case "l3":
		mode = C.IPVLAN_MODE_L3
	default:
		return CNIIPvlan{}, nil
	}

	return CNIIPvlan{
		_CNIVlan: _CNIVlan{
			HostInterface: conf.HostInterface,
			Pid:           pid,
			VName:         vname,
			Name:          conf.Name,
			IP:            ip,
			Mask:          mask,
		},
		_Pid:  C.unsigned(cPid),
		_Mode: mode,
		Mode:  conf.Mode,
	}, nil
}

func (conf CNIIPvlan) BuildNetwork() (err error) {
	// csys.SystemCmd("ip", "link",
	// 	"add", "link", conf.HostInterface, "name", conf.VName,
	// 	"type", "ipvlan", "mode", conf.Mode)
	// csys.SystemCmd("ip", "link",
	// 	"set", conf.VName, "netns", conf.Pid)
	if C.net_create_ipvlan(C.CString(conf.HostInterface),
		C.CString(conf.VName), conf._Mode, conf._Pid) != 0 {
		return errors.New("create ipvlan failed")
	}
	return
}

type CNIMacvlan struct {
	_CNIVlan
	Mode string
}

func NewCNIMacvlan(cPid int, conf CNetworkInterface) (CNIMacvlan, error) {
	var (
		devid uint32 = atomic.AddUint32(&devUniqID, 1)
		pid   string = strconv.Itoa(cPid)
		vname string = fmt.Sprintf("macv%s-%d", pid, devid)
		ip    net.IP
		mask  net.IP
	)

	if conf.IP != "" {
		mask = net.ParseIP(conf.Mask)
		ip = net.ParseIP(conf.IP)

		if ip == nil || mask == nil {
			return CNIMacvlan{}, fmt.Errorf("invalid ip (%s) or mask (%s)",
				conf.IP, conf.Mask)
		}
	}

	if conf.Mode != "bridge" && conf.Mode != "vepa" &&
		conf.Mode != "private" {
		return CNIMacvlan{}, nil
	}

	return CNIMacvlan{
		_CNIVlan: _CNIVlan{
			HostInterface: conf.HostInterface,
			Pid:           pid,
			VName:         vname,
			Name:          conf.Name,
			IP:            ip,
			Mask:          mask,
		},
		Mode: conf.Mode,
	}, nil
}

func (conf CNIMacvlan) BuildNetwork() (err error) {
	csys.SystemCmd("ip", "link",
		"add", "link", conf.HostInterface, "name", conf.VName,
		"type", "macvlan", "mode", conf.Mode)
	csys.SystemCmd("ip", "link",
		"set", conf.VName, "netns", conf.Pid)
	return
}
