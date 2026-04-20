package main

import (
	"fetrunner"
	"fetrunner/internal/base"
	"unsafe"
)

// #include <stdlib.h>
import "C"

// The communication is sequential in that FetRunnerReadLog is never
// called before the last call has returned. Thus it is enough to keep
// a reference to the returned string in the single variable, `cmsg`,
// until the next call to FetRunnerReadLog.
var cmsg *C.char

//export FetRunnerCommand
func FetRunnerCommand(cString *C.char) {
	gString := C.GoString(cString)
	fetrunner.Dispatch(gString)
}

//export FetRunnerReadLog
func FetRunnerReadLog() *C.char {
	// Blocks until there is a line to read.
	line := base.LogTake()
	C.free(unsafe.Pointer(cmsg)) // cmsg == `nil` is OK
	cmsg = C.CString(line)
	return cmsg
}

func main() {}
