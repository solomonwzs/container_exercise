package main

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"syscall"
)

type MainServer struct {
	uniquid uint64
	rwc     io.ReadWriteCloser
}

func SystemCmd(cmd string, arg ...string) (err error) {
	c := exec.Command(cmd, arg...)
	c.Stdin = os.Stdin
	c.Stdout = os.Stdout
	c.Stderr = os.Stderr
	if err = c.Run(); err != nil {
		return
	}
	c.Wait()
	return
}

func NewSocketpair() (f0, f1 *os.File, err error) {
	fds, err := syscall.Socketpair(syscall.AF_UNIX, syscall.SOCK_STREAM, 0)
	if err != nil {
		return nil, nil, os.NewSyscallError("socketpair", err)
	}
	f0 = os.NewFile(uintptr(fds[0]), fmt.Sprintf("socketpair-%d", fds[0]))
	f1 = os.NewFile(uintptr(fds[1]), fmt.Sprintf("socketpair-%d", fds[1]))
	return
}

func NewPipe() (rFs, wFs *os.File, err error) {
	fds := make([]int, 2)
	if err = syscall.Pipe(fds); err != nil {
		return nil, nil, os.NewSyscallError("pipe", err)
	}
	rFs = os.NewFile(uintptr(fds[0]), fmt.Sprintf("pipe-r-%d", fds[0]))
	wFs = os.NewFile(uintptr(fds[1]), fmt.Sprintf("pipe-w-%d", fds[1]))
	return
}
