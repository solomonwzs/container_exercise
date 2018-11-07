package main

import (
	"encoding/binary"
	"errors"
	"fmt"
	"net"
	"strconv"
	"sync/atomic"
)

var devUniqID uint32 = 0

type NetworkBuilder interface {
	BuildNetwork() error
	ReleaseNetwork() error
	SetupNetwork() error
}

func ParserNetworkBuilders(cPid int, conf Configuration) []NetworkBuilder {
	nb := make([]NetworkBuilder, len(conf.Networks))
	for i, netConf := range conf.Networks {
		switch netConf.Type {
		case "bridge":
			nb[i] = NewCNetworkBridge(cPid, netConf)
		default:
			nb[i] = nil
		}
	}
	return nb
}

type CNetworkBridge struct {
	BridgeAddr string
	BridgeName string
	DevName    string
	Pid        string
	RouteIP    string
	RuleSrc    string
	VethA      string
	VethAddr   string
	VethB      string
}

func NewCNetworkBridge(cPid int, conf CNetwork) CNetworkBridge {
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

	routeIP := Dec2Ipv4(ipDec&maskDec | 1)
	bridgeAddr := fmt.Sprintf("%s/%d", routeIP, maskValidBits)
	vethAddr := fmt.Sprintf("%s/%d", ip, maskValidBits)

	ruleSrc := fmt.Sprintf("%s/%d", Dec2Ipv4(ipDec&maskDec), maskValidBits)

	return CNetworkBridge{
		BridgeAddr: bridgeAddr,
		BridgeName: bridgeName,
		DevName:    conf.Name,
		Pid:        pid,
		RouteIP:    fmt.Sprintf("%s", routeIP),
		RuleSrc:    ruleSrc,
		VethA:      vethA,
		VethAddr:   vethAddr,
		VethB:      vethB,
	}
}

func (conf CNetworkBridge) BuildNetwork() (err error) {
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

func (conf CNetworkBridge) ReleaseNetwork() (err error) {
	SystemCmd("iptables",
		"-t", "nat",
		"-D", "POSTROUTING",
		"-s", conf.RuleSrc,
		"-j", "MASQUERADE")

	SystemCmd("ip", "link", "delete", conf.BridgeName, "type", "bridge")

	return
}

func (conf CNetworkBridge) SetupNetwork() (err error) {
	SystemCmd("ip", "link", "set", "lo", "up")
	SystemCmd("ip", "link", "set", conf.VethB, "name", conf.DevName)
	SystemCmd("ip", "link", "set", conf.DevName, "up")
	SystemCmd("ip", "addr", "add", conf.VethAddr, "dev", conf.DevName)
	SystemCmd("ip", "route", "add", "default", "via", conf.RouteIP)
	return
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
