package main

import "unsafe"

const (
	OP_NOTIFY = iota
	OP_C_END
)

var (
	ProtoReqId = uint64(0)

	SIZEOF_PROTO_IN_HEADER = int(unsafe.Sizeof(ProtoInHeader{}))
)

type ProtoInHeader struct {
	OpCode uint32
	Unique uint64
	Len    uint32
}
