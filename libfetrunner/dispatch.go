package main

import (
	"fetrunner/base"
	"os"
	"strings"
)

//TODO: The logger collects entries in a buffer. Should Logger rather be an
// interface?

type ALogger interface {
	Enter()
}

type ThisLogger struct {
	logchan chan []base.LogEntry
	logbuf  []base.LogEntry
}

func NewLogger() ThisLogger {
	return ThisLogger{
		logchan: make(chan []base.LogEntry),
	}
}

// The loggers must be made available somehow. This map makes them accessible
// via a tag.
// TODO: Would an integer key be better?
var loggerMap map[string]ThisLogger

func init() {
	// default logger
	loggerMap = map[string]ThisLogger{"": NewLogger()}
}

func Dispatch(cmd0 string) string {
	var result string
	cmdsplit := strings.Fields(cmd0)
	switch cmd := cmdsplit[0]; cmd {

	case "CONFIG_DIR":
		dir, dirErr := os.UserConfigDir()
		if dirErr == nil {
			result = "> config dir: " + dir
		} else {
			result = "! No config dir"
		}

	case "CONFIG_INIT":
		//TODO: Needs adapting, the call is now
		// logger.InitConfig()
		//was base.InitConfig()

	default:
		result = "! Invalid command: " + cmd0

	}
	return result
}
