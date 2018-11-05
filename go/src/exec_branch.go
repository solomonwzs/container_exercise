package main

import (
	"os"
	"os/exec"
	"syscall"
)

var registeredBranch = make(map[string]func())

func ExecBranch() bool {
	if init, exists := registeredBranch[os.Args[0]]; exists {
		init()
		return true
	}
	return false
}

func RegisterBranchCommand(name string, init func()) bool {
	if _, exists := registeredBranch[name]; exists {
		return false
	}
	registeredBranch[name] = init
	return true
}

func BranchCommand(args ...string) *exec.Cmd {
	return &exec.Cmd{
		Path: _PATH_PROC_BINARY,
		Args: args,
		SysProcAttr: &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGTERM,
		},
	}
}
