package main

import "unsafe"

const (
	_MAX_PROTO_REQ_SIZE = 1024

	PFUNC_WAIT_FOR_REPLY = 1

	OP_NOTIFY = iota
	OP_FUNCTION_CALL
)

var (
	ProtoReqId = uint64(0)

	SIZEOF_PROTO_HEADER = int(unsafe.Sizeof(ProtoHeader{}))
)

type ProtoFuncCallback func([]byte)

type ProtoHeader struct {
	OpCode uint32
	Unique uint64
	Len    uint32
}

type ProtoFuncCall struct {
	FuncCode uint32
	RawLen   uint32
	Flags    uint32
	padding  uint32
}

type ProtoFuncReply struct {
	Unique  uint64
	RawLen  uint32
	padding uint32
}
