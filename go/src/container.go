package main

import (
	"io"
	"os"
	"os/exec"
	"strconv"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/solomonwzs/goxutil/closer"
)

type ContainerServer struct {
	closer.Closer
	uniquid  uint64
	conf     *Configuration
	rwc      io.ReadWriteCloser
	callback map[uint64]ProtoFuncCallback
}

func NewContainerServer(rwc io.ReadWriteCloser, conf *Configuration) (
	s *ContainerServer) {
	s = &ContainerServer{
		uniquid: 0,
		conf:    conf,
		rwc:     rwc,
	}
	s.Closer = closer.NewCloser(func() error {
		return s.rwc.Close()
	})
	return
}

func (s *ContainerServer) Serv() {
	buf := make([]byte, _MAX_PROTO_REQ_SIZE)
	for {
		_, err := s.rwc.Read(buf)
		if err != nil {
			panic(err)
		}
	}
}

func containerRun() {
	filename := os.Args[1]
	fd, _ := strconv.Atoi(os.Args[2])
	var conf Configuration
	if _, err := toml.DecodeFile(filename, &conf); err != nil {
		panic(err)
	}

	mgrs := os.NewFile(uintptr(fd), "mgrs")
	defer mgrs.Close()

	buf := make([]byte, 1)
	mgrs.Read(buf)

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
