package main

import (
	"encoding/binary"
	"flag"
	"fmt"
	"os"
	"strconv"
	"syscall"

	"github.com/BurntSushi/toml"
	"github.com/solomonwzs/goxutil/logger"
)

const (
	_BRANCH_CONTAINER = "container"
)

func init() {
	logger.NewLogger(func(r *logger.Record) {
		fmt.Printf("%s", r)
	})

	switch os.Args[0] {
	case _BRANCH_CONTAINER:
		containerRun()
		os.Exit(0)
	default:
	}
}

func UidMap(pid, idInsideNs, idOutsideNs, mapRange int) (err error) {
	f, err := os.OpenFile(PathProcUidMap(pid), os.O_RDWR|os.O_CREATE, 0755)
	if err != nil {
		return
	}
	defer f.Close()

	_, err = f.WriteString(fmt.Sprintf("%d %d %d",
		idInsideNs, idOutsideNs, mapRange))
	return
}

func GidMap(pid, idInsideNs, idOutsideNs, mapRange int) (err error) {
	f, err := os.OpenFile(PathProcGidMap(pid), os.O_RDWR|os.O_CREATE, 0755)
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

	f0, f1, err := NewSocketpair()
	if err != nil {
		panic(err)
	}
	defer f0.Close()
	defer f1.Close()

	process, err := os.StartProcess(
		_PATH_PROC_BINARY,
		[]string{_BRANCH_CONTAINER, filename, strconv.Itoa(int(f1.Fd()))},
		&os.ProcAttr{
			Files: []*os.File{os.Stdin, os.Stdout, os.Stderr},
			Sys: &syscall.SysProcAttr{
				Pdeathsig: syscall.SIGTERM,
				Cloneflags: syscall.CLONE_NEWPID |
					syscall.CLONE_NEWNS |
					syscall.CLONE_NEWUTS |
					syscall.CLONE_NEWIPC |
					syscall.CLONE_NEWNET,
			},
		})
	if err != nil {
		panic(err)
	}

	networkBuilders := ParserNetworkBuilders(process.Pid, conf)
	for _, builder := range networkBuilders {
		builder.BuildNetwork()
		defer builder.ReleaseNetwork()
	}

	buf := make([]byte, 4)
	binary.BigEndian.PutUint32(buf, uint32(process.Pid))
	f0.Write(buf)

	defer ReleaseBaseFiles(&conf)

	if _, err = process.Wait(); err != nil {
		panic(err)
	}
}

func releaseContainerResource(conf *Configuration) {
	ReleaseBaseFiles(conf)
}
