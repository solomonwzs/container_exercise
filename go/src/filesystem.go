package main

import (
	"fmt"
	"io"
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
	mountArgs{"proc", "/proc", "proc", 0, ""},
	mountArgs{"sysfs", "/sys", "sysfs", 0, ""},
	mountArgs{"none", "/tmp", "tmpfs", 0, ""},
	mountArgs{"udev", "/dev", "devtmpfs", 0, ""},
	mountArgs{"devpts", "/dev/pts", "devpts", 0, ""},
	mountArgs{"shm", "/dev/shm", "tmpfs", 0, ""},
	mountArgs{"tmpfs", "/run", "tmpfs", 0, ""},
}

var copyFiles = [][2]string{
	[2]string{"/etc/resolv.conf", "/etc/resolv.conf"},
	[2]string{"/etc/hosts", "/etc/hosts"},
}

func RootPaht(conf *Configuration) string {
	return filepath.Join(conf.BaseSys.Workspace, conf.Name, "meger")
}

func CopyFile(src, dst string) (err error) {
	stat0, err := os.Stat(src)
	if err != nil {
		return
	}
	if !stat0.Mode().IsRegular() {
		return fmt.Errorf("cp: non-regular src file %s", stat0.Name())
	}

	stat1, err := os.Stat(dst)
	if err != nil {
		if !os.IsNotExist(err) {
			return
		}
	} else {
		if !stat1.Mode().IsRegular() {
			return fmt.Errorf("cp: non-regular dest file %s", stat1.Name())
		}
		if os.SameFile(stat0, stat1) {
			return
		}
	}

	f0, err := os.OpenFile(src, os.O_RDONLY, stat0.Mode().Perm())
	if err != nil {
		return
	}
	defer f0.Close()

	f1, err := os.OpenFile(dst, os.O_RDWR|os.O_CREATE, stat0.Mode().Perm())
	if err != nil {
		return
	}
	defer f1.Close()

	if _, err = io.Copy(f1, f0); err != nil {
		return
	}

	return f1.Sync()
}

func MountCustomFiles(path string, customMountList []CMount) error {
	for _, m := range customMountList {
		target := filepath.Join(path, m.Target)
		stat, err := os.Stat(target)
		if err != nil {
			if os.IsNotExist(err) {
				os.MkdirAll(target, 0755)
			} else {
				logger.Errorln(err)
				continue
			}
		} else if !stat.Mode().IsDir() {
			logger.Errorln(fmt.Errorf("non-dir: %s", stat.Name()))
			continue
		}

		if err = syscall.Mount(m.Source, target,
			"none", syscall.MS_BIND, ""); err != nil {
			logger.Errorln(err)
		}
	}
	return nil
}

func UmountCustomFiles(path string, customMountList []CMount) error {
	for i := len(customMountList) - 1; i >= 0; i-- {
		m := customMountList[i]
		target := filepath.Join(path, m.Target)
		if err := syscall.Unmount(target, 0); err != nil {
			logger.Errorln(err)
		}
	}
	return nil
}

func BuildBaseFiles(conf *Configuration) (err error) {
	lowerPath := filepath.Join(conf.BaseSys.Dir, conf.BaseSys.System)
	mergePath := RootPaht(conf)
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
			logger.Errorln(args, err0)
		}
	}

	for _, fs := range copyFiles {
		dest := filepath.Join(mergePath, fs[1])
		if err0 := CopyFile(fs[0], dest); err0 != nil {
			logger.Errorln(err0)
		}
	}

	MountCustomFiles(mergePath, conf.Mount)

	if err0 := syscall.Chdir(mergePath); err0 != nil {
		logger.Errorln(err0)
	}
	if err0 := syscall.Chroot("./"); err0 != nil {
		logger.Errorln(err0)
	}

	return nil
}

func ReleaseBaseFiles(conf *Configuration) {
	mergePath := RootPaht(conf)

	UmountCustomFiles(mergePath, conf.Mount)

	for i := len(mountList) - 1; i >= 0; i-- {
		arg := mountList[i]
		target := filepath.Join(mergePath, arg.target)
		if err := syscall.Unmount(target, 0); err != nil {
			logger.Errorln(err)
		}
	}

	if err := syscall.Unmount(mergePath, 0); err != nil {
		logger.Errorln(err)
	}
}
