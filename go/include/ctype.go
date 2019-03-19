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
*/
import "C"

type (
	ContInHeader C.struct_cont_in_header
)
