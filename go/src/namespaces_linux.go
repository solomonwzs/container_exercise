package main

import (
	"path/filepath"
	"syscall"
)

func printNamespacesInfo(ns string) {
	path := filepath.Join(_PATH_PROC_NAMESPACE, ns)
	buf := make([]byte, 64)
	syscall.Readlink(path, buf)
	info := string(buf)
	print(info)
	print(len(info))
}
