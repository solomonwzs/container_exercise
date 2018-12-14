package cnet

import (
	"errors"
	"net"
	"time"

	"github.com/solomonwzs/goxutil/dhcp"
	"github.com/solomonwzs/goxutil/net/transport"
)

var ERR_NOT_EXPECTED_REPLY = errors.New("not expected reply")

type DHCPReply struct {
	ClientIP   net.IP
	DHPServer  net.IP
	SubnetMask net.IP
	Router     net.IP
	LeaseTime  time.Duration
}

func DHCPApply(interf *net.Interface) (reply DHCPReply, err error) {
	conn, err := transport.NewUDPBroadcastRawConn(interf,
		dhcp.CLIENT_PORT, dhcp.SERVER_PORT)
	if err != nil {
		return
	}
	defer conn.Close()
	buf := make([]byte, 1024)

	// discover
	msg := dhcp.NewMessageForInterface(interf)
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
		return DHCPReply{}, ERR_NOT_EXPECTED_REPLY
	}
	clientIP := rMsg.ClientIP()
	dhcpServerIP, err := rMsg.DHCPServerID()
	if err != nil {
		return DHCPReply{}, ERR_NOT_EXPECTED_REPLY
	}

	// request
	msg = dhcp.NewMessageForInterface(interf)
	msg.SetBroadcast()
	msg.SetMessageType(dhcp.DHCPREQUEST)
	msg.SetOptions(dhcp.OPT_ADDR_REQUEST, []byte(clientIP.To4()))
	msg.SetOptions(dhcp.OPT_DHCP_SERVER_ID, []byte(dhcpServerIP.To4()))
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
		return DHCPReply{}, ERR_NOT_EXPECTED_REPLY
	}
	dhcpServer, err := rMsg.DHCPServerID()
	if err != nil {
		return DHCPReply{}, ERR_NOT_EXPECTED_REPLY
	}
	subnetMask, err := rMsg.SubnetMask()
	if err != nil {
		return DHCPReply{}, ERR_NOT_EXPECTED_REPLY
	}
	router, err := rMsg.Router()
	if err != nil {
		return DHCPReply{}, ERR_NOT_EXPECTED_REPLY
	}
	leaseTime, err := rMsg.AddressLeaseTime()
	if err != nil {
		return DHCPReply{}, ERR_NOT_EXPECTED_REPLY
	}

	return DHCPReply{
		ClientIP:   clientIP,
		DHPServer:  dhcpServer,
		SubnetMask: subnetMask,
		Router:     router,
		LeaseTime:  time.Duration(leaseTime),
	}, nil
}
