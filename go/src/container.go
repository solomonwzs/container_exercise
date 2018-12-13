package main

import (
	"cnet"
	"csys"
	"encoding/binary"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/solomonwzs/goxutil/closer"
	"github.com/solomonwzs/goxutil/logger"
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
	ch := make(chan os.Signal)
	signal.Notify(ch, os.Interrupt, os.Kill, syscall.SIGINT,
		syscall.SIGTERM, syscall.SIGCHLD)
	go func() {
		for {
			select {
			case sig := <-ch:
				if sig == syscall.SIGCHLD {
					syscall.Wait4(-1, nil, 0, nil)
				}
			}
		}
	}()

	filename := os.Args[1]
	f0, _ := strconv.Atoi(os.Args[2])
	f1, _ := strconv.Atoi(os.Args[3])
	var conf Configuration
	if _, err := toml.DecodeFile(filename, &conf); err != nil {
		panic(err)
	}

	fm := os.NewFile(uintptr(f0), "sock-main")
	fm.Close()

	syscall.CloseOnExec(f1)
	mgrs := os.NewFile(uintptr(f1), "mgrs")
	defer mgrs.Close()

	csys.SystemCmd("id")

	buf := make([]byte, 4)
	mgrs.Read(buf)
	pid := int(binary.BigEndian.Uint32(buf))
	logger.Debug(pid)

	// set network
	networkBuilders := cnet.ParserNetworkBuilders(pid, conf.Network)
	for _, builder := range networkBuilders {
		builder.SetupNetwork()
	}
	cnet.AddNetworkRoutes(conf.Network.Routes)

	// mount
	if err := BuildBaseFiles(&conf); err != nil {
		logger.Error(err)
		// panic(err)
	}

	// set hostname
	if err := syscall.Sethostname([]byte(conf.Hostname)); err != nil {
		logger.Error(err)
		// panic(err)
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
