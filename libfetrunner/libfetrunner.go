package main

import (
	"os"
	"strings"
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
	var result string
	gString := C.GoString(cString)
	cmd := strings.Fields(gString)
	switch cmd0 := cmd[0]; cmd0 {

	case "CONFIG_DIR":
		dir, dirErr := os.UserConfigDir()
		if dirErr == nil {
			result = "> config dir: " + dir
		} else {
			result = "! No config dir"
		}

	default:
		result = "! Invalid command: " + gString

	}
	C.free(unsafe.Pointer(cmsg)) // cmsg == `nil` is OK
	cmsg = C.CString(result)
	return cmsg
}

func main() {}
