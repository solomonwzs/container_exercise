package main

import "fmt"

const (
	_PATH_PROC_NAMESPACE = "/proc/self/ns"
	_PATH_PROC_BINARY    = "/proc/self/exe"
	_PATH_PROC_UID_MAP   = "/proc/self/uid_map"
	_PATH_PROC_GID_MAP   = "/proc/self/uid_map"
)

func PathProcUidMap(pid int) string {
	return fmt.Sprintf("/proc/%d/uid_map", pid)
}

func PathProcGidMap(pid int) string {
	return fmt.Sprintf("/proc/%d/gid_map", pid)
}

func PathProcNsIPC(pid int) string {
	return fmt.Sprintf("/proc/%d/ns/ipc", pid)
}

func PathProcNsMount(pid int) string {
	return fmt.Sprintf("/proc/%d/ns/mnt", pid)
}

func PathProcNsNet(pid int) string {
	return fmt.Sprintf("/proc/%d/ns/net", pid)
}

func PathProcNsPid(pid int) string {
	return fmt.Sprintf("/proc/%d/ns/pid", pid)
}

func PathProcNsUTS(pid int) string {
	return fmt.Sprintf("/proc/%d/ns/uts", pid)
}

func PathProcNsUser(pid int) string {
	return fmt.Sprintf("/proc/%d/ns/user", pid)
}
