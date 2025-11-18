package main

import "unsafe"

// #include <stdlib.h>
import "C"

var cmsg *C.char

//export Test
func Test(cString *C.char) *C.char {
	gString := ">> " + C.GoString(cString)
	C.free(unsafe.Pointer(cmsg)) // cmsg == `nil` is OK
	cmsg = C.CString(gString)
	return cmsg
}

func main() {}
