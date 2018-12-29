package cnet

/*
#cgo CFLAGS:	-I${SRCDIR}/../../include
#cgo CFLAGS:	-I${SRCDIR}/../../deps/iproute2/include
#cgo LDFLAGS:	${SRCDIR}/../../build/libnetwork.a
#cgo LDFLAGS:	${SRCDIR}/../../deps/iproute2/lib/libnetlink.a
#cgo LDFLAGS:	${SRCDIR}/../../deps/iproute2/lib/libutil.a
#cgo LDFLAGS:	-lmnl -lcap
*/
import "C"
