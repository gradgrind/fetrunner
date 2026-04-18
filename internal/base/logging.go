package base

import (
	"fmt"
	"path/filepath"
	"runtime"
)

/*
	The logger passes lines to stdout.

All operations cause a OP_START to be logged before doing anything else,
and an OP_END at the end. Any results or other log entries associated
with an operation will be between these two entries.

A TICK is output directly, not as part of an operation,
so it has no OP_END.
*/

type LogType int

const (
	none LogType = iota
	INFO
	WARNING
	ERROR
	BUG

	OP_START
	OP_END
)

var logType = map[LogType]string{
	INFO:    "*INFO*",
	WARNING: "*WARNING*",
	ERROR:   "*ERROR*",
	BUG:     "*BUG*",

	OP_START: "+++",
	OP_END:   "---",
}

func (ltype LogType) String() string {
	s, ok := logType[ltype]
	if !ok {
		panic(fmt.Sprintf("Invalid LogType: %d", ltype))
	}
	return s
}

type Logger struct {
}

func NewLogger() *Logger {
	return &Logger{}
}

func (l *Logger) logMessage(ltype LogType, s string, a ...any) {
	fmt.Printf(ltype.String()+" "+s, a...)
	//TODO: Do I need to trim? lstring := strings.TrimSpace(fmt.Sprintf(s, a...))
}

func (l *Logger) Info(s string, a ...any) {
	l.logMessage(INFO, s, a...)
}

func (l *Logger) Result(key string, value any) {
	fmt.Printf("$ %s=%v\n", key, value)
}

func (l *Logger) Warning(s string, a ...any) {
	l.logMessage(WARNING, s, a...)
}

func (l *Logger) Error(s string, a ...any) {
	l.logMessage(ERROR, s, a...)
}

func (l *Logger) Bug(s string, a ...any) {
	var p string
	_, f, ln, ok := runtime.Caller(1)
	if ok {
		d := filepath.Dir(f)
		fp := filepath.Join(filepath.Base(d), filepath.Base(f))
		p = fmt.Sprintf("%s @ %d: ", fp, ln)
	} else {
		p = "Location?: "
	}
	l.logMessage(BUG, p+s, a...)
}

func (l *Logger) Tick(n int) {
	l.Result(".TICK", n)
}
