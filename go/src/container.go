package main

import (
	"flag"
	"os"
	"os/exec"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/solomonwzs/goxutil/logger"
)

const _MGR_FD = 4

func containerRun() {
	fd := os.NewFile(uintptr(_MGR_FD), "mgr_socket")
	defer fd.Close()

	buf := make([]byte, SIZEOF_PROTO_IN_HEADER)
	if _, err := fd.Read(buf); err != nil {
		logger.Error(err)
	}
	logger.Debug(buf)

	var conf Configuration
	var filename string
	flag.StringVar(&filename, "f", "", "config filename")
	flag.Parse()
	if _, err := toml.DecodeFile(filename, &conf); err != nil {
		panic(err)
	}

	// mount
	if err := BuildBaseFiles(&conf); err != nil {
		panic(err)
	}

	// set hostname
	if err := syscall.Sethostname([]byte(conf.Hostname)); err != nil {
		panic(err)
	}

	// run
	cmd := exec.Command("/bin/bash")
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Env = conf.Env
	if err := cmd.Run(); err != nil {
		panic(err)
	}
	if err := cmd.Wait(); err != nil {
		panic(err)
	}
}
