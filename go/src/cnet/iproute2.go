package cnet

/*
#include "network.h"
#include "uapi/linux/if.h"
#include <errno.h>
*/
import "C"
import "errors"

const (
	IFF_UP         = C.IFF_UP
	IFF_MULTICAST  = C.IFF_MULTICAST
	IFF_ALLMULTI   = C.IFF_ALLMULTI
	IFF_PROMISC    = C.IFF_PROMISC
	IFF_NOTRAILERS = C.IFF_NOTRAILERS
	IFF_NOARP      = C.IFF_NOARP
	IFF_DYNAMIC    = C.IFF_DYNAMIC
)

type NetDevFlags struct {
	name  string
	mask  int32
	flags int32
}

func NewNetDevFlags(name string) *NetDevFlags {
	return &NetDevFlags{
		name:  name,
		mask:  0,
		flags: 0,
	}
}

func (f *NetDevFlags) SetUp(up bool) *NetDevFlags {
	f.mask |= IFF_UP
	if up {
		f.flags |= IFF_UP
	} else {
		f.flags &= ^IFF_UP
	}
	return f
}

func (f *NetDevFlags) SetMulticast(on bool) *NetDevFlags {
	f.mask |= IFF_MULTICAST
	if on {
		f.flags |= IFF_MULTICAST
	} else {
		f.flags &= ^IFF_MULTICAST
	}
	return f
}

func (f *NetDevFlags) SetAllMulticast(on bool) *NetDevFlags {
	f.mask |= IFF_ALLMULTI
	if on {
		f.flags |= IFF_ALLMULTI
	} else {
		f.flags &= ^IFF_ALLMULTI
	}
	return f
}

func (f *NetDevFlags) SetPromisc(on bool) *NetDevFlags {
	f.mask |= IFF_PROMISC
	if on {
		f.flags |= IFF_PROMISC
	} else {
		f.flags &= ^IFF_PROMISC
	}
	return f
}

func (f *NetDevFlags) SetTrailers(on bool) *NetDevFlags {
	f.mask |= C.IFF_NOTRAILERS
	if on {
		f.flags &= ^IFF_NOTRAILERS
	} else {
		f.flags |= IFF_NOTRAILERS
	}
	return f
}

func (f *NetDevFlags) SetArp(on bool) *NetDevFlags {
	f.mask |= IFF_NOARP
	if on {
		f.flags &= ^IFF_NOARP
	} else {
		f.flags |= IFF_NOARP
	}
	return f
}

func (f *NetDevFlags) SetDynamic(on bool) *NetDevFlags {
	f.mask |= IFF_DYNAMIC
	if on {
		f.flags |= IFF_DYNAMIC
	} else {
		f.flags &= ^IFF_DYNAMIC
	}
	return f
}

func (f *NetDevFlags) Commit() error {
	if C.iplink_chflags(C.CString(f.name), C.uint32_t(f.flags),
		C.uint32_t(f.mask)) != 0 {
		return errors.New("change flags fail")
	}
	return nil
}
