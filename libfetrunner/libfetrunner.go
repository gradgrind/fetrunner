package main

import (
	"fetrunner"
	"fetrunner/internal/base"
	"fmt"
	"unsafe"
)

// #include <stdlib.h>
import "C"

// The communication is sequential in that FetRunnerReadLog is never
// called before the last call has returned. Thus it is enough to keep
// a reference to the returned string in the single variable, `cmsg`,
// until the next call to FetRunnerReadLog.
var cmsg *C.char

func init() {
	base.LogToBuffer()
}

//export FetRunnerCommand
func FetRunnerCommand(cString *C.char) C.int {
	gString := C.GoString(cString)

	//TODO--
	fmt.Println("§", gString)
	//return 2

	if fetrunner.Dispatch(gString) {
		//TODO--
		fmt.Println("§§1")

		return 1
	} else {
		//TODO--
		fmt.Println("§§2")

		return 0
	}
}

//export FetRunnerReadLog
func FetRunnerReadLog() *C.char {
	//fmt.Println("FetRunnerReadLog()")
	// Blocks until there is a line to read.
	line := base.ReadLogBufferLine()
	//fmt.Printf("+ %s\n", line)
	C.free(unsafe.Pointer(cmsg)) // cmsg == `nil` is OK
	cmsg = C.CString(line)
	return cmsg
}

func main() {}
