package main

import (
	"os"
	"os/exec"
	"sync"
	"syscall"
)

type SysProcess struct {
	*exec.Cmd
}

var (
	registeredProcesses       = make(map[string]func())
	processID           int32 = 0
	addLock                   = &sync.Mutex{}
)

func NewSysProcess(args ...string) *SysProcess {
	addLock.Lock()
	defer addLock.Unlock()

	processID += 1

	return nil
}

var registeredInitFunc = make(map[string]func())

func ExecInitFunc() bool {
	if init, exists := registeredInitFunc[os.Args[0]]; exists {
		init()
		return true
	}
	return false
}

func RegisterInitFunc(name string, init func()) bool {
	if _, exists := registeredInitFunc[name]; exists {
		return false
	}
	registeredInitFunc[name] = init
	return true
}

func InitFuncCommand(args ...string) *exec.Cmd {
	return &exec.Cmd{
		Path: _PATH_PROC_BINARY,
		Args: args,
		SysProcAttr: &syscall.SysProcAttr{
			Pdeathsig: syscall.SIGTERM,
		},
	}
}
