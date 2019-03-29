package main

import (
	"cnet"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/BurntSushi/toml"
	"github.com/solomonwzs/goxutil/logger"
)

func getMessageSock() *os.File {
	f0, _ := strconv.Atoi(os.Args[2])
	f1, _ := strconv.Atoi(os.Args[3])

	fm := os.NewFile(uintptr(f0), "sock-main")
	fm.Close()

	syscall.CloseOnExec(f1)
	fc := os.NewFile(uintptr(f1), "sock-cont")
	return fc
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
	end := make(chan struct{})

	f := getMessageSock()
	defer f.Close()

	conf, err := getConfiguration()
	if err != nil {
		panic(err)
	}

	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGCHLD)
	go func() {
		for {
			select {
			case <-ch:
				syscall.Wait4(-1, nil, 0, nil)
			}
		}
	}()

	go func() {
		buf := make([]byte, MAX_PROTO_REQ_SIZE)
		for {
			n, err := f.Read(buf)
			if err != nil {
				logger.Errorln(err)
				return
			}

			header := (*ContInHeader)(unsafe.Pointer(&buf[0]))
			if uint32(n) != header.Len {
				continue
			}

			switch header.Opcode {
			case CONT_INIT:
				initIn := (*ContInitIn)(
					unsafe.Pointer(&buf[SIZEOF_CONT_IN_HEADER]))
				contHandlerInit(&conf, header, initIn, end)
			}
		}
	}()
	<-end
}

func contHandlerInit(conf *Configuration, header *ContInHeader,
	initIn *ContInitIn, end chan struct{}) {
	logger.Debugln(initIn.Pid)
	// set network
	networkBuilders := cnet.ParserNetworkBuilders(int(initIn.Pid),
		conf.Network)
	for _, builder := range networkBuilders {
		builder.SetupNetwork()
	}
	cnet.AddNetworkRoutes(conf.Network.Routes)

	// mount
	if err := BuildBaseFiles(conf); err != nil {
		panic(err)
	}

	// set hostname
	if err := syscall.Sethostname([]byte(conf.Hostname)); err != nil {
		panic(err)
	}

	go func() {
		// run
		cmd := exec.Command("/bin/bash")
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		cmd.Env = conf.Env
		if err := cmd.Run(); err != nil {
			logger.Errorln(err)
		}
		UmountContainFileSystems()
		close(end)
	}()
}
