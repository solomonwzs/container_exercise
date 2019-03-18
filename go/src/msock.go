/**
 * @author  Solomon Ng <solomon.wzs@gmail.com>
 * @version 1.0
 * @date    2019-03-18
 * @license MIT
 */

package main

import (
	"encoding/binary"
	"io"
	"sync"
)

type MSock struct {
	io.ReadWriter
	buffer []byte
	lock   *sync.RWMutex
}

func NewMSock(rw io.ReadWriter) *MSock {
	return &MSock{
		ReadWriter: rw,
		buffer:     make([]byte, 1024),
		lock:       &sync.RWMutex{},
	}
}

func (ms *MSock) ReadUint32() (i uint32, err error) {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	_, err = ms.Read(ms.buffer)
	if err != nil {
		return
	}
	i = binary.BigEndian.Uint32(ms.buffer)
	return
}

func (ms *MSock) WriteUint32(i uint32) error {
	ms.lock.Lock()
	defer ms.lock.Unlock()

	binary.BigEndian.PutUint32(ms.buffer, i)
	_, err := ms.Write(ms.buffer[:4])
	return err
}
