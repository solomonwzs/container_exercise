package main

import (
	"time"
	"unsafe"
)

const (
	MAX_PROTO_REQ_SIZE = 1024

	PFUNC_WAIT_FOR_REPLY = 1

	SIZEOF_PROTO_HEADER = int(unsafe.Sizeof(ProtoHeader{}))

	OP_INIT_CONTAINER = iota
	OP_FUNCTION_CALL
)

var (
	ProtoReqId = uint64(time.Now().UnixNano())
)

type ProtoFuncCallback func([]byte)

type ProtoHeader struct {
	OpCode uint32
	Unique uint64
	Len    uint32
}
