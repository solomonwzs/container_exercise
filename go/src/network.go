package main

import (
	"encoding/binary"
	"net"

	"github.com/solomonwzs/goxutil/logger"
)

func BuildNetworks(cPid int, conf *Configuration) (err error) {
	for _, netConf := range conf.Networks {
		switch netConf.Type {
		case "bridge":
			buildNetworkBridge(conf.Name, cPid, &netConf)
		default:
		}
	}
	return
}

func buildNetworkBridge(name string, cPid int, conf *CNetwork) (err error) {
	// pid := strconv.Itoa(cPid)
	// bridge := fmt.Sprintf("%s-bridge-%s", name, pid)
	// vethA := fmt.Sprintf("%s-veth-%d", name, pid)
	// vethB := fmt.Sprintf("%s-veth-%d-0", name, pid)

	// SystemCmd("ip", "link", "add", "name", bridge, "type", "bridge")
	// SystemCmd("ip", "link", "set", bridge, "up")

	mark := net.ParseIP(conf.Mark)
	n := binary.BigEndian.Uint32(mark[12:16])

	ip := make(net.IP, 4)
	binary.BigEndian.PutUint32(ip, n)
	logger.Debug(ip)

	return
}
