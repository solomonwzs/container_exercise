package main

import (
	"fmt"
	"os"
	"path/filepath"
	"syscall"

	"github.com/solomonwzs/goxutil/logger"
)

type mountArgs struct {
	src    string
	target string
	ftype  string
	flags  uintptr
	data   string
}

var mountList = []mountArgs{
	mountArgs{"proc", "proc", "proc", 0, ""},
	mountArgs{"sysfs", "sys", "sysfs", 0, ""},
	mountArgs{"none", "tmp", "tmpfs", 0, ""},
	mountArgs{"udev", "dev", "devtmpfs", 0, ""},
	mountArgs{"devpts", "dev/pts", "devpts", 0, ""},
	mountArgs{"shm", "dev/shm", "tmpfs", 0, ""},
	mountArgs{"tmpfs", "run", "tmpfs", 0, ""},
}

func BuildBaseFiles(conf *Configuration) (err error) {
	lowerPath := filepath.Join(conf.BaseSys.Dir, conf.BaseSys.System)
	mergePath := filepath.Join(conf.BaseSys.Workspace, conf.Name, "meger")
	upperPath := filepath.Join(conf.BaseSys.Workspace, conf.Name, "upper")
	workPath := filepath.Join(conf.BaseSys.Workspace, conf.Name, "work")

	os.MkdirAll(mergePath, 0755)
	os.MkdirAll(upperPath, 0755)
	os.MkdirAll(workPath, 0755)

	data := fmt.Sprintf("lowerdir=%s,upperdir=%s,workdir=%s",
		lowerPath, upperPath, workPath)
	if err = syscall.Mount("overlay", mergePath, "overlay",
		0, data); err != nil {
		return os.NewSyscallError("mount", err)
	}

	for _, args := range mountList {
		target := filepath.Join(mergePath, args.target)
		if err0 := syscall.Mount(args.src, target, args.ftype,
			args.flags, args.data); err0 != nil {
			logger.Error(err0)
		}
	}

	if err0 := syscall.Chdir(mergePath); err0 != nil {
		logger.Error(err0)
	}

	if err0 := syscall.Chroot("./"); err0 != nil {
		logger.Error(err0)
	}

	return nil
}

func ReleaseBaseFiles(conf *Configuration) {
	mergePath := filepath.Join(conf.BaseSys.Workspace, conf.Name, "meger")

	for i := len(mountList) - 1; i >= 0; i-- {
		arg := mountList[i]
		target := filepath.Join(mergePath, arg.target)
		if err := syscall.Unmount(target, 0); err != nil {
			logger.Error(err)
		}
	}

	if err := syscall.Unmount(mergePath, 0); err != nil {
		logger.Error(err)
	}
}
