package cnet

/*
#include "network.h"
#include "uapi/linux/if.h"
#include <errno.h>
#include <stdlib.h>

#define SIZEOF_PTR sizeof(void *)
*/
import "C"
import (
	"errors"
	"unsafe"
)

const (
	MACVLAN_MODE_BRIDGE  = C.MACVLAN_MODE_BRIDGE
	MACVLAN_MODE_VEPA    = C.MACVLAN_MODE_VEPA
	MACVLAN_MODE_PRIVATE = C.MACVLAN_MODE_PRIVATE

	IPVLAN_MODE_L2  = C.IPVLAN_MODE_L2
	IPVLAN_MODE_L3  = C.IPVLAN_MODE_L3
	IPVLAN_MODE_L3S = C.IPVLAN_MODE_L3S

	IFF_UP         = C.IFF_UP
	IFF_MULTICAST  = C.IFF_MULTICAST
	IFF_ALLMULTI   = C.IFF_ALLMULTI
	IFF_PROMISC    = C.IFF_PROMISC
	IFF_NOTRAILERS = C.IFF_NOTRAILERS
	IFF_NOARP      = C.IFF_NOARP
	IFF_DYNAMIC    = C.IFF_DYNAMIC

	IFNAMSIZ = C.IFNAMSIZ
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
		return errors.New("change flags failed")
	}
	return nil
}

func AddRoute(argv []string) (err error) {
	if len(argv) > 1024 {
		return errors.New("too many arguments")
	}

	arr := C.malloc(C.SIZEOF_PTR * C.size_t(len(argv)))
	defer C.free(unsafe.Pointer(arr))

	goArr := (*[1024]*C.char)(arr)
	for i, str := range argv {
		goArr[i] = C.CString(str)
	}

	if C.iproute_add(C.int(len(argv)), (**C.char)(arr)) != 0 {
		return errors.New("add route failed")
	}
	return nil
}

func AddAddr(argv []string) (err error) {
	if len(argv) > 1024 {
		return errors.New("too many arguments")
	}

	arr := C.malloc(C.SIZEOF_PTR * C.size_t(len(argv)))
	defer C.free(unsafe.Pointer(arr))

	goArr := (*[1024]*C.char)(arr)
	for i, str := range argv {
		goArr[i] = C.CString(str)
	}

	if C.ipaddr_add(C.int(len(argv)), (**C.char)(arr)) != 0 {
		return errors.New("ad addr failed")
	}
	return nil
}

func CreateBridge(name string) (err error) {
	if C.iplink_create_bridge(C.CString(name)) != 0 {
		return errors.New("create bridge failed")
	}
	return nil
}

func CreateVeths(nameA, nameB string, pid int) (err error) {
	if C.iplink_create_veth(C.CString(nameA), C.CString(nameB),
		C.unsigned(pid)) != 0 {
		return errors.New("create veths failed")
	}
	return nil
}

func SetDevMaster(name string, master string) (err error) {
	if C.iplink_set_master(C.CString(name), C.CString(master)) != 0 {
		return errors.New("set master failed")
	}
	return nil
}

func DeleteDev(name string) (err error) {
	if C.iplink_delete_dev(C.CString(name)) != 0 {
		return errors.New("delete device failed")
	}
	return nil
}

func RenameDev(name, newName string) (err error) {
	if C.iplink_rename(C.CString(name), C.CString(newName)) != 0 {
		return errors.New("rename device failed")
	}
	return nil
}

func CreateVlan(host string, name string, pid int, id uint16) (err error) {
	if C.iplink_create_vlan(C.CString(host), C.CString(name),
		C.unsigned(pid), C.uint16_t(id)) != 0 {
		return errors.New("create vlan failed")
	}
	return nil
}

func CreateIPVlan(host string, name string, pid int, mode uint16) (
	err error) {
	if C.iplink_create_ipvlan(C.CString(host), C.CString(name),
		C.unsigned(pid), C.uint16_t(mode)) != 0 {
		return errors.New("create ipvlan failed")
	}
	return nil
}

func CreateMACVlan(host string, name string, pid int, mode uint32) (
	err error) {
	if C.iplink_create_macvlan(C.CString(host), C.CString(name),
		C.unsigned(pid), C.uint32_t(mode)) != 0 {
		return errors.New("create macvlan failed")
	}
	return nil
}
