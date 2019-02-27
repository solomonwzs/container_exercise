/**
 * @author  Solomon Ng <solomon.wzs@gmail.com>
 * @version 1.0
 * @date    2019-02-27
 * @license MIT
 */

package csys

import (
	"bufio"
	"errors"
	"io"
	"os"
	"strconv"
	"strings"
)

const (
	MOFT_SHARED    = 0x00
	MOFT_MASTER    = 0x01
	MOFT_PROPAGATE = 0X02
)

type OptionField struct {
	Type  int
	Group int
}

type Mountinfo struct {
	MountID      int
	ParentID     int
	Major        int
	Minor        int
	Root         string
	MountPoint   string
	MountOptions []string
	OptionFields []OptionField
	Type         string
	MountSrc     string
	SuperOptions []string
}

func ParseMountinfo(line string) (minfo Mountinfo, err error) {
	info := strings.Split(line, " ")
	length := len(info)

	i := 0
	if i == length {
		err = errors.New("invalid mountinfo: mount id")
		return
	}
	str := info[i]
	if minfo.MountID, err = strconv.Atoi(str); err != nil {
		err = errors.New("invalid mountinfo: mount id")
		return
	}

	i += 1
	if i == length {
		err = errors.New("invalid mountinfo: parent id")
		return
	}
	str = info[i]
	if minfo.ParentID, err = strconv.Atoi(str); err != nil {
		err = errors.New("invalid mountinfo: parent id")
		return
	}

	i += 1
	if i == length {
		err = errors.New("invalid mountinfo: major/minor")
		return
	}
	str = info[i]
	arr := strings.Split(str, ":")
	if len(arr) != 2 {
		err = errors.New("invalid mountinfo: major/minor")
		return
	}
	if minfo.Major, err = strconv.Atoi(arr[0]); err != nil {
		err = errors.New("invalid mountinfo: major/minor")
		return
	}
	if minfo.Minor, err = strconv.Atoi(arr[1]); err != nil {
		err = errors.New("invalid mountinfo: major/minor")
		return
	}

	i += 1
	if i == length {
		err = errors.New("invalid mountinfo: root")
		return
	}
	minfo.Root = info[i]

	i += 1
	if i == length {
		err = errors.New("invalid mountinfo: mount point")
		return
	}
	minfo.MountPoint = info[i]

	i += 1
	if i == length {
		err = errors.New("invalid mountinfo: mount options")
		return
	}
	minfo.MountOptions = strings.Split(info[i], ",")

	i += 1
	if i == length {
		err = errors.New("invalid mountinfo: option fields")
		return
	}
	minfo.OptionFields = make([]OptionField, 0)
	for ; i < length && info[i] != "-"; i++ {
	}

	return
}

func ParsePidMountinfos(pid int) (infoList []Mountinfo, err error) {
	f, err := os.Open(PathProcMountInfo(pid))
	if err != nil {
		return
	}
	defer f.Close()

	reader := bufio.NewReader(f)
	infoList = make([]Mountinfo, 0)
	var minfo Mountinfo
	var line string
	for {
		line, err = reader.ReadString('\n')
		if err == io.EOF {
			err = nil
			break
		} else if err != nil {
			return
		}

		minfo, err = ParseMountinfo(line)
		if err == nil {
			infoList = append(infoList, minfo)
		}
	}

	return
}
