package csys

/*
#include <string.h>
#include <errno.h>
*/
import "C"
import (
	"errors"
	"unsafe"
)

const (
	_ERRNO_DESC_LEN = 64
)

func ErrnoDesc(errnum int) error {
	if errnum == 0 {
		return nil
	}

	buf := make([]byte, _ERRNO_DESC_LEN)
	C.strerror_r(C.int(errnum), (*C.char)(unsafe.Pointer(&buf[0])),
		_ERRNO_DESC_LEN)
	return errors.New(string(buf))
}
