package main

import (
	"os"
	"syscall"
)

type MainManager struct {
	uniquid uint64
	fd      *os.File
}

type ContainerManager struct {
	uniquid uint64
	fd      *os.File
}

func NewSocketpair() (f0, f1 *os.File, err error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, nil, os.NewSyscallError("socketpair", err)
	}
	f0 = os.NewFile(uintptr(fds[0]), "socketpair-0")
	f1 = os.NewFile(uintptr(fds[1]), "socketpair-1")
	return
}
