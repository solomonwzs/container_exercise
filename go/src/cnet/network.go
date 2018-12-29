package cnet

/*
#include "uapi/linux/if.h"
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

const _IP_CMD = "/home/solomon/workspace/c/iproute2/ip/ip"

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

func CheckIfName(name string) bool {
	l := len(name)
	if l == 0 || l > C.IFNAMSIZ {
		return false
	}
	for i := 0; i < l; i++ {
		if name[i] == '/' || name[i] == ' ' {
			return false
		}
	}
	return true
}

func ParserNetworkBuilders(cPid int, conf CNetwork) []NetworkBuilder {
	nb := make([]NetworkBuilder, 0)
	var builder NetworkBuilder
	var err error
	for _, netConf := range conf.Interfaces {
		if !CheckIfName(netConf.Name) {
			logger.Errorf("invalid name: %s\n", netConf.Name)
			continue
		}

		switch netConf.Type {
		case "bridge":
			builder, err = NewCNIBridge(cPid, netConf)
		case "vlan":
			builder, err = NewCNIVlan(cPid, netConf)
		case "ipvlan":
			builder, err = NewCNIIPvlan(cPid, netConf)
		case "macvlan":
			builder, err = NewCNIMacvlan(cPid, netConf)
		default:
			err = fmt.Errorf("invalid type: %s", netConf.Type)
		}

		if err != nil {
			logger.Error(err)
		} else {
			nb = append(nb, builder)
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

		csys.SystemCmd(_IP_CMD, "route", "add", dest, "via", route.Gateway)
		// AddRoute([]string{dest, "via", route.Gateway})
	}
	return
}

type CNIBridge struct {
	BridgeAddr string
	BridgeName string
	Name       string
	Pid        int
	RuleSrc    string
	VethA      string
	VethAddr   string
	VethB      string
}

func NewCNIBridge(cPid int, conf CNetworkInterface) (*CNIBridge, error) {
	devid := atomic.AddUint32(&devUniqID, 1)

	bridgeName := fmt.Sprintf("bridge%d-%d", cPid, devid)
	vethA := fmt.Sprintf("veth%d-%d", cPid, devid)
	vethB := fmt.Sprintf("veth%dx-%d", cPid, devid)

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
		Pid:        cPid,
		RuleSrc:    ruleSrc,
		VethA:      vethA,
		VethAddr:   vethAddr,
		VethB:      vethB,
	}, nil
}

func (conf CNIBridge) BuildNetwork() (err error) {
	// create bridge
	C.iplink_create_bridge(C.CString(conf.BridgeName))
	csys.SystemCmd(_IP_CMD, "addr",
		"add", conf.BridgeAddr, "brd", "+", "dev", conf.BridgeName)
	NewNetDevFlags(conf.BridgeName).SetUp(true).Commit()

	// create a pair of veths
	C.iplink_create_veth(C.CString(conf.VethA), C.CString(conf.VethB),
		C.unsigned(conf.Pid))

	// set veth to bridge
	C.iplink_set_master(C.CString(conf.VethA), C.CString(conf.BridgeName))
	NewNetDevFlags(conf.VethA).SetUp(true).Commit()

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

	C.iplink_delete_dev(C.CString(conf.BridgeName))

	return
}

func (conf CNIBridge) SetupNetwork() (err error) {
	C.iplink_rename(C.CString(conf.VethB), C.CString(conf.Name))
	NewNetDevFlags("lo").SetUp(true).Commit()
	NewNetDevFlags(conf.Name).SetUp(true).Commit()
	csys.SystemCmd(_IP_CMD, "addr", "add", conf.VethAddr, "dev", conf.Name)
	return
}

type _CNIVlan struct {
	HostInterface string
	VName         string
	Name          string
	Pid           int
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
	C.iplink_rename(C.CString(conf.VName), C.CString(conf.Name))
	NewNetDevFlags("lo").SetUp(true).Commit()
	NewNetDevFlags(conf.Name).SetUp(true).Commit()

	addr, err := conf.getAddr()
	if err != nil {
		logger.Error(err)
		return
	}
	csys.SystemCmd(_IP_CMD, "addr", "add", addr, "dev", conf.Name)
	return
}

type CNIVlan struct {
	_CNIVlan
	ID uint16
}

func NewCNIVlan(cPid int, conf CNetworkInterface) (CNIVlan, error) {
	vname := fmt.Sprintf("vlan-%s", conf.ID)
	mask := net.ParseIP(conf.Mask)
	ip := net.ParseIP(conf.IP)

	if ip == nil || mask == nil {
		return CNIVlan{}, fmt.Errorf("invalid ip (%s) or mask (%s)",
			conf.IP, conf.Mask)
	}

	id, err := strconv.ParseUint(conf.ID, 10, 16)
	if err != nil {
		return CNIVlan{}, fmt.Errorf("invalid vlan id: %s", conf.ID)
	}

	return CNIVlan{
		_CNIVlan: _CNIVlan{
			HostInterface: conf.HostInterface,
			Pid:           cPid,
			VName:         vname,
			Name:          conf.Name,
			IP:            ip,
			Mask:          mask,
		},
		ID: uint16(id),
	}, nil
}

func (conf CNIVlan) BuildNetwork() (err error) {
	if C.iplink_create_vlan(C.CString(conf.HostInterface),
		C.CString(conf.VName), C.unsigned(conf.Pid),
		C.uint16_t(conf.ID)) != 0 {
		return errors.New("create vlan failed")
	}
	return
}

type CNIIPvlan struct {
	_CNIVlan
	Mode uint16
}

func NewCNIIPvlan(cPid int, conf CNetworkInterface) (CNIIPvlan, error) {
	var (
		devid uint32 = atomic.AddUint32(&devUniqID, 1)
		pid   string = strconv.Itoa(cPid)
		vname string = fmt.Sprintf("ipv%s-%d", pid, devid)
		ip    net.IP
		mask  net.IP
		mode  uint16
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
	case "l3s":
		mode = C.IPVLAN_MODE_L3S
	default:
		return CNIIPvlan{}, fmt.Errorf("invalid mode: %s", conf.Mode)
	}

	return CNIIPvlan{
		_CNIVlan: _CNIVlan{
			HostInterface: conf.HostInterface,
			Pid:           cPid,
			VName:         vname,
			Name:          conf.Name,
			IP:            ip,
			Mask:          mask,
		},
		Mode: mode,
	}, nil
}

func (conf CNIIPvlan) BuildNetwork() (err error) {
	if C.iplink_create_ipvlan(C.CString(conf.HostInterface),
		C.CString(conf.VName), C.unsigned(conf.Pid),
		C.uint16_t(conf.Mode)) != 0 {
		return errors.New("create ipvlan failed")
	}
	return
}

type CNIMacvlan struct {
	_CNIVlan
	Mode uint32
}

func NewCNIMacvlan(cPid int, conf CNetworkInterface) (CNIMacvlan, error) {
	var (
		devid uint32 = atomic.AddUint32(&devUniqID, 1)
		pid   string = strconv.Itoa(cPid)
		vname string = fmt.Sprintf("macv%s-%d", pid, devid)
		ip    net.IP
		mask  net.IP
		mode  uint32
	)

	if conf.IP != "" {
		mask = net.ParseIP(conf.Mask)
		ip = net.ParseIP(conf.IP)

		if ip == nil || mask == nil {
			return CNIMacvlan{}, fmt.Errorf("invalid ip (%s) or mask (%s)",
				conf.IP, conf.Mask)
		}
	}

	switch conf.Mode {
	case "bridge":
		mode = C.MACVLAN_MODE_BRIDGE
	case "vepa":
		mode = C.MACVLAN_MODE_VEPA
	case "private":
		mode = C.MACVLAN_MODE_PRIVATE
	default:
		return CNIMacvlan{}, fmt.Errorf("invalid mode: %s", conf.Mode)
	}

	return CNIMacvlan{
		_CNIVlan: _CNIVlan{
			HostInterface: conf.HostInterface,
			Pid:           cPid,
			VName:         vname,
			Name:          conf.Name,
			IP:            ip,
			Mask:          mask,
		},
		Mode: mode,
	}, nil
}

func (conf CNIMacvlan) BuildNetwork() (err error) {
	if C.iplink_create_macvlan(C.CString(conf.HostInterface),
		C.CString(conf.VName), C.unsigned(conf.Pid),
		C.uint32_t(conf.Mode)) != 0 {
		return errors.New("create macvlan failed")
	}
	return
}
