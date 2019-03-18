package main

import (
	"cnet"
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
	conf     Configuration
	rwc      io.ReadWriteCloser
	callback map[uint64]ProtoFuncCallback
}

func NewContainerServer(rwc io.ReadWriteCloser, conf Configuration) (
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
	buf := make([]byte, MAX_PROTO_REQ_SIZE)
	for {
		_, err := s.rwc.Read(buf)
		if err != nil {
			logger.Errorln(err)
			return
		}
	}
}

func getMessageSock() *os.File {
	f0, _ := strconv.Atoi(os.Args[2])
	f1, _ := strconv.Atoi(os.Args[3])

	fm := os.NewFile(uintptr(f0), "sock-main")
	fm.Close()

	syscall.CloseOnExec(f1)
	f := os.NewFile(uintptr(f1), "mgrs")
	return f
}

func getConfiguration() (Configuration, error) {
	filename := os.Args[1]
	var conf Configuration
	if _, err := toml.DecodeFile(filename, &conf); err != nil {
		return conf, err
	}
	return conf, nil
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

	f := getMessageSock()
	mgrs := NewMSock(f)
	defer f.Close()

	conf, err := getConfiguration()
	if err != nil {
		panic(err)
	}

	p, _ := mgrs.ReadUint32()
	pid := int(p)
	logger.Debugln(pid)

	// set network
	networkBuilders := cnet.ParserNetworkBuilders(pid, conf.Network)
	for _, builder := range networkBuilders {
		builder.SetupNetwork()
	}
	cnet.AddNetworkRoutes(conf.Network.Routes)

	// mount
	if err := BuildBaseFiles(&conf); err != nil {
		panic(err)
	}
	defer UmountContainFileSystems()

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
		logger.Errorln(err)
	}
}
