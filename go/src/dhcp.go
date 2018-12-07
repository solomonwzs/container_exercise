package main

import (
	"errors"
	"net"
	"time"

	"github.com/solomonwzs/goxutil/dhcp"
	"github.com/solomonwzs/goxutil/net/transport"
)

type DHCPReply struct {
	ClientIP   net.IP
	DHPServer  net.IP
	SubnetMask net.IP
	Router     net.IP
	LeaseTime  time.Duration
}

func DHCPApply(interf *net.Interface) (reply *DHCPReply, err error) {
	conn, err := transport.NewUDPBroadcastConn(
		dhcp.CLIENT_PORT, dhcp.SERVER_PORT)
	if err != nil {
		return
	}
	defer conn.Close()
	buf := make([]byte, 1024)

	// discover
	msg := dhcp.NewMessaageForInterface(interf)
	msg.SetMessageType(dhcp.DHCPDISCOVER)
	msg.SetBroadcast()
	if _, err = conn.Write(msg.Marshal()); err != nil {
		return
	}

	// offer
	n, err := conn.Read(buf)
	if err != nil {
		return
	}
	rMsg, err := dhcp.Unmarshal(buf[:n])
	if err != nil {
		return
	} else if t, err := rMsg.MessageType(); err != nil ||
		t != dhcp.DHCPOFFER {
		return nil, errors.New("not expected reply")
	}
	clientIP := rMsg.ClientIP()
	dhcpServerIP, err := rMsg.DHCPServerID()
	if err != nil {
		return nil, errors.New("not expected reply")
	}

	// request
	msg = dhcp.NewMessaageForInterface(interf)
	msg.SetBroadcast()
	msg.SetMessageType(dhcp.DHCPREQUEST)
	msg.SetOptions(dhcp.OPT_ADDR_REQUEST, []byte(clientIP))
	msg.SetOptions(dhcp.OPT_DHCP_SERVER_ID, []byte(dhcpServerIP))
	if _, err = conn.Write(msg.Marshal()); err != nil {
		return
	}

	// ack
	n, err = conn.Read(buf)
	if err != nil {
		return
	}
	rMsg, err = dhcp.Unmarshal(buf[:n])
	if err != nil {
		return
	} else if t, err := rMsg.MessageType(); err != nil ||
		t != dhcp.DHCPACK {
		return nil, errors.New("not expected reply")
	}

	return
}
