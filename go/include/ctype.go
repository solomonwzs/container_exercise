/**
 * @author  Solomon Ng <solomon.wzs@gmail.com>
 * @version 1.0
 * @date    2019-03-19
 * @license MIT
 */

package main

/*
#cgo CFLAGS: -I${SRCDIR}/..
#include "include/cont_proto.h"

#define SIZEOF_CONT_IN_HEADER sizeof(struct cont_in_header)
#define SIZEOF_CONT_INIT_IN sizeof(struct cont_init_in)
#define SIZEOF_CONT_INIT_OUT sizeof(struct cont_init_out)
*/
import "C"

const (
	SIZEOF_CONT_IN_HEADER = C.SIZEOF_CONT_IN_HEADER
	SIZEOF_CONT_INIT_IN   = C.SIZEOF_CONT_INIT_IN
	SIZEOF_CONT_INIT_OUT  = C.SIZEOF_CONT_INIT_OUT

	CONT_INIT = C.CONT_INIT
)

type (
	ContInHeader C.struct_cont_in_header
	ContInitIn   C.struct_cont_init_in
	ContInitOut  C.struct_cont_init_out
)
