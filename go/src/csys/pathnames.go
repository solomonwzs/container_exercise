package csys

import "fmt"

const (
	PATH_PROC_NAMESPACE = "/proc/self/ns"
	PATH_PROC_BINARY    = "/proc/self/exe"
	PATH_PROC_UID_MAP   = "/proc/self/uid_map"
	PATH_PROC_GID_MAP   = "/proc/self/uid_map"
	PATH_PROC_MOUNTINFO = "/proc/self/mountinfo"
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

func PathProcMountInfo(pid int) string {
	return fmt.Sprintf("/proc/%d/mountinfo", pid)
}
