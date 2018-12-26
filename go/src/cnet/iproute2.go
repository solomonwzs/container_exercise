package cnet

/*
#include "network.h"
#include "uapi/linux/if.h"
*/
import "C"

const (
	DEV_MASK_STATUS       = C.IFF_UP
	DEV_MASK_MULTICAST    = C.IFF_MULTICAST
	DEV_MASK_ALLMULTICAST = C.IFF_ALLMULTI
	DEV_MASK_PROMISC      = C.IFF_PROMISC
	DEV_MASK_TRAILERS     = C.IFF_NOTRAILERS
	DEV_MASK_ARP          = C.IFF_NOARP
	DEV_MASK_DYNAMIC      = C.IFF_DYNAMIC
)

type NetDevFlags struct {
	mask  int32
	flags int32
}

func (f *NetDevFlags) SetStatus(up bool) {
	f.mask |= DEV_MASK_STATUS
	if up {
		f.flags |= DEV_MASK_STATUS
	} else {
		f.flags &= ^DEV_MASK_STATUS
	}
}
