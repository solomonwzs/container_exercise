package main

import (
	"flag"
	"fmt"
	"os"
	"os/exec"
	"syscall"

	"github.com/BurntSushi/toml"
)

func init() {
	RegisterInitFunc("foo", foo)
	if ExecInitFunc() {
		os.Exit(0)
	}
}

func foo() {
	fmt.Println(os.Getpid(), "foo", os.Args)
	fmt.Println("hello")

	if err := syscall.Mount("proc", "/proc", "proc", 0, ""); err != nil {
		panic(err)
	}
	defer func() {
		err := syscall.Unmount("/proc", 0)
		fmt.Println("unmount", err)
	}()

	cmd := exec.Command("/bin/bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err := cmd.Run(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}

func main() {
	var filename string
	var conf Configuration

	flag.StringVar(&filename, "f", "", "config filename")
	flag.Parse()
	fmt.Println(os.Getpid(), "main", os.Args)

	if _, err := toml.DecodeFile(filename, &conf); err != nil {
		panic(err)
	}

	cmd := InitFuncCommand("foo")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.SysProcAttr.Cloneflags = syscall.CLONE_NEWPID | syscall.CLONE_NEWNS

	if err := cmd.Start(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}
