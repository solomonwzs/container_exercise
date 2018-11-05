package main

import "fmt"

const (
	_PATH_PROC_NAMESPACE = "/proc/self/ns"
	_PATH_PROC_BINARY    = "/proc/self/exe"
	_PATH_PROC_UID_MAP   = "/proc/self/uid_map"
	_PATH_PROC_GID_MAP   = "/proc/self/uid_map"
)

func procUidMap(pid int) string {
	return fmt.Sprintf("/proc/%d/uid_map", pid)
}

func procGidMap(pid int) string {
	return fmt.Sprintf("/proc/%d/gid_map", pid)
}
