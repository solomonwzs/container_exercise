package main

import (
	"flag"
	"fmt"
	"os"
	"sync/atomic"
	"syscall"
	"unsafe"

	"github.com/BurntSushi/toml"
	"github.com/solomonwzs/goxutil/logger"
)

func init() {
	logger.NewLogger(func(r *logger.Record) {
		fmt.Printf("%s", r)
	})

	RegisterBranchCommand("container", containerRun)
	if ExecBranch() {
		os.Exit(0)
	}
}

func main() {
	f0, f1, err := NewSocketpair()
	if err != nil {
		panic(err)
	}
	defer f0.Close()
	defer f1.Close()

	var conf Configuration
	var filename string
	flag.StringVar(&filename, "f", "", "config filename")
	flag.Parse()
	if _, err := toml.DecodeFile(filename, &conf); err != nil {
		panic(err)
	}

	var args = os.Args
	args[0] = "container"

	cmd := BranchCommand(args...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	// cmd.ExtraFiles = []*os.File{f1}
	cmd.SysProcAttr.Cloneflags = syscall.CLONE_NEWPID |
		syscall.CLONE_NEWNS |
		syscall.CLONE_NEWUTS |
		syscall.CLONE_NEWIPC |
		syscall.CLONE_NEWNET

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	logger.Debug(cmd.Process.Pid)

	buf := make([]byte, SIZEOF_PROTO_IN_HEADER)
	inHeader := (*ProtoInHeader)(unsafe.Pointer(&buf[0]))
	inHeader.Len = 0
	inHeader.OpCode = OP_NOTIFY
	inHeader.Unique = atomic.AddUint64(&ProtoReqId, 1)
	logger.Debug(f0.Write(buf))

	defer func() {
		ReleaseBaseFiles(&conf)
	}()

	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}
