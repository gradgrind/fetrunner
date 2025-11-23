package main

import (
	"fetrunner/base"
	"unsafe"
)

// #include <stdlib.h>
import "C"

// The communication is sequential in that FetRunner is never called
// before the last call has returned. Thus it is enough to keep a
// reference to the returned string in the single variable, `cmsg`,
// until the next call to FetRunner.
var cmsg *C.char

//export FetRunner
func FetRunner(cString *C.char) *C.char {
	gString := C.GoString(cString)
	result := base.Dispatch(logger, gString)
	C.free(unsafe.Pointer(cmsg)) // cmsg == `nil` is OK
	cmsg = C.CString(result)
	return cmsg
}

var logger *base.Logger

func init() {
	// Set up logger.
	logger = base.NewLogger()
	go base.LogToBuffer(logger)
}

func main() {}
