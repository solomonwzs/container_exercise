package main

/*
#define _GNU_SOURCE
#include <sched.h>
#include "network.h"
*/
import "C"
import (
	"cnet"
	"csys"
	"flag"
	"fmt"
	"os"
	"strconv"
	"syscall"
	"unsafe"

	"github.com/BurntSushi/toml"
	"github.com/solomonwzs/goxutil/logger"
)

const (
	_BRANCH_CONTAINER = "container"
)

type _LP struct{}

func (lp _LP) L(r *logger.Record) error { fmt.Printf("%s", r); return nil }
func (lp _LP) Close() error             { return nil }

func init() {
	logger.NewLogger(_LP{})

	switch os.Args[0] {
	case _BRANCH_CONTAINER:
		containerRun()
		os.Exit(0)
	default:
	}
}

func UidMap(pid, idInsideNs, idOutsideNs, mapRange int) (err error) {
	f, err := os.OpenFile(csys.PathProcUidMap(pid), os.O_RDWR|os.O_CREATE,
		0755)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%d %d %d",
		idInsideNs, idOutsideNs, mapRange))
	return
}

func GidMap(pid, idInsideNs, idOutsideNs, mapRange int) (err error) {
	f, err := os.OpenFile(csys.PathProcGidMap(pid), os.O_RDWR|os.O_CREATE,
		0755)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%d %d %d",
		idInsideNs, idOutsideNs, mapRange))
	return
}

func main() {
	var conf Configuration
	var filename string
	flag.StringVar(&filename, "f", "", "config filename")
	flag.Parse()
	if _, err := toml.DecodeFile(filename, &conf); err != nil {
		panic(err)
	}
	logger.Debugf("%+v\n", conf)

	f, fCont, err := cnet.NewSocketpair()
	if err != nil {
		panic(err)
	}
	defer f.Close()
	defer fCont.Close()

	process, err := os.StartProcess(
		csys.PATH_PROC_BINARY,
		[]string{_BRANCH_CONTAINER,
			filename,
			strconv.Itoa(int(f.Fd())),
			strconv.Itoa(int(fCont.Fd())),
		},
		&os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
			Sys: &syscall.SysProcAttr{
				Pdeathsig: syscall.SIGTERM,
				Cloneflags: syscall.CLONE_NEWPID |
					// syscall.CLONE_NEWUSER |
					C.CLONE_NEWCGROUP |
					syscall.CLONE_NEWNS |
					syscall.CLONE_NEWUTS |
					syscall.CLONE_NEWIPC |
					syscall.CLONE_NEWNET,
			},
		})
	if err != nil {
		panic(err)
	}

	networkBuilders := cnet.ParserNetworkBuilders(
		process.Pid, conf.Network)
	for _, builder := range networkBuilders {
		if err = builder.BuildNetwork(); err != nil {
			logger.Errorln(err)
		}
		defer builder.ReleaseNetwork()
	}

	UidMap(process.Pid, 0, os.Getuid(), 1)
	GidMap(process.Pid, 0, os.Getgid(), 1)

	buf := make([]byte, MAX_PROTO_REQ_SIZE)
	header := (*ContInHeader)(unsafe.Pointer(&buf[0]))
	header.Len = SIZEOF_CONT_IN_HEADER + SIZEOF_CONT_INIT_IN
	header.Opcode = CONT_INIT
	initIn := (*ContInitIn)(
		unsafe.Pointer(&buf[SIZEOF_CONT_IN_HEADER]))
	initIn.Pid = uint32(process.Pid)
	f.Write(buf[:header.Len])

	defer ReleaseContainer(&conf)

	if _, err = process.Wait(); err != nil {
		panic(err)
	}
}

func ReleaseContainer(conf *Configuration) {
	rootPath := RootPath(conf)
	if err := syscall.Unmount(rootPath, 0); err != nil {
		logger.Errorln(rootPath, err)
	}
}
