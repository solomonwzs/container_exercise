package main

import (
	"io"
)

type MainServer struct {
	uniquid uint64
	rwc     io.ReadWriteCloser
}
