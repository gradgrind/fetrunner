package main

import (
	"fetrunner/autotimetable"
	"fetrunner/base"
	"fetrunner/db"
	"fetrunner/fet"
	"fetrunner/w365tt"
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

// Handle (currently) ".fet" and "_w365.json" input files.
func file_loader(logger *base.Logger, op *base.DispatchOp) {
	if !logger.CheckArgs(op, 1) {
		return
	}
	fpath := op.Data[0]

	//TODO: what to do with the data structures produced here?
	// Should they be attached to the logger? That might require an
	// "any" field, not least because the Logger struct is defined in
	// package "base".
	if strings.HasSuffix(fpath, ".fet") {
		bdata := &autotimetable.BasicData{}
		bdata.SetParameterDefault()
		bdata.Logger = logger
		if fet.FetRead(bdata, fpath) {
			logger.Result(op.Op, fpath)
			logger.Result("DATA_TYPE", "FET")
			return
		}
	} else if strings.HasSuffix(fpath, "_w365.json") {
		db0 := db.NewDb(logger)
		if w365tt.LoadJSON(db0, fpath) {
			db0.PrepareDb()
			logger.Result(op.Op, fpath)
			logger.Result("DATA_TYPE", "DB")
			return
		}
	} else {
		logger.Error("LoadFile_InvalidSuffix: %s", fpath)
		return
	}
	logger.Error("LoadFile_InvalidContent: %s", fpath)
}

func init() {
	base.OpHandlerMap["SET_FILE"] = file_loader
}
