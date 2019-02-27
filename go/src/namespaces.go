package main

import (
	"csys"
	"path/filepath"
	"syscall"
)

func NamespacesID(ns string) string {
	path := filepath.Join(csys.PATH_PROC_NAMESPACE, ns)
	buf := make([]byte, 64)
	syscall.Readlink(path, buf)

	for i, ch := range buf {
		if ch == 0 {
			return string(buf[:i])
		}
	}
	return ""
}

func EnterNamespaces() {
}
