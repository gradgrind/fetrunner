package base

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

type LogType int

const (
	none LogType = iota
	INFO
	WARNING
	ERROR
	BUG

	COMMAND
	RESULT

	STARTOP
	ENDOP
)

var logType = map[LogType]string{
	INFO:    "*INFO*",
	WARNING: "*WARNING*",
	ERROR:   "*ERROR*",
	BUG:     "*BUG*",
	COMMAND: "#",
	RESULT:  "$",
	STARTOP: "+++",
	ENDOP:   "---",
}

func (ltype LogType) String() string {
	s, ok := logType[ltype]
	if !ok {
		panic(fmt.Sprintf("Invalid LogType: %d", ltype))
	}
	return s
}

type LogEntry struct {
	Type LogType
	Text string
}

func (e LogEntry) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type string
		Text string
	}{
		Type: e.Type.String(),
		Text: e.Text,
	})
}

type Logger struct {
	LogChan    chan LogEntry
	LogBuf     []LogEntry
	ResultChan chan string
}

func NewLogger() *Logger {
	return &Logger{
		LogChan:    make(chan LogEntry),
		ResultChan: make(chan string),
	}
}

// LogToFile allows the log entries to be saved to a file, as they are
// generated.
// Run it as a goroutine.
func LogToFile(logger *Logger, logpath string) {
	file, err := os.OpenFile(logpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	for entry := range logger.LogChan {
		lstring := entry.Type.String() + " " + entry.Text
		file.WriteString(lstring + "\n")
	}
}

func (l *Logger) logEnter(ltype LogType, s string, a ...any) {
	lstring := strings.TrimSpace(fmt.Sprintf(s, a...))
	l.LogChan <- LogEntry{ltype, lstring}
}

func (l *Logger) Info(s string, a ...any) {
	l.logEnter(INFO, s, a...)
}

func (l *Logger) Result(key string, value string) {
	l.logEnter(RESULT, "%s=%s", key, value)
}

func (l *Logger) Warning(s string, a ...any) {
	l.logEnter(WARNING, s, a...)
}

func (l *Logger) Error(s string, a ...any) {
	l.logEnter(ERROR, s, a...)
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
	l.logEnter(BUG, p+s, a)
}

// TODO?
var CONSOLE bool

func Report(msg string) {
	if CONSOLE {
		fmt.Print(msg)
	}
}
