package base

import (
	"encoding/json"
	"fmt"
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
	TICKOP
	POLLOP
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
	TICKOP:  ".TICK",
	POLLOP:  ".POLL",
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
	pollwait   int8
	ticked     bool
}

func NewLogger() *Logger {
	return &Logger{
		LogChan:    make(chan LogEntry),
		ResultChan: make(chan string),
	}
}

// Log entry handler adding log entries to a buffer.
// Run it as a goroutine.
func LogToBuffer(logger *Logger) {
	for entry := range logger.LogChan {
		switch entry.Type {

		case TICKOP:
			logger.LogBuf = append(logger.LogBuf, LogEntry{
				RESULT, TICKOP.String() + "=" + entry.Text})
			if logger.pollwait != 2 {
				logger.ticked = true
				continue
			}

		case POLLOP:
			//logger.LogBuf = append(logger.LogBuf, entry)
			logger.pollwait = 1
			continue

		case ENDOP:
			//logger.LogBuf = append(logger.LogBuf, entry)
			if !logger.ticked {
				if logger.pollwait == 1 {
					logger.pollwait = 2
					continue
				}
			}

		default:
			logger.LogBuf = append(logger.LogBuf, entry)
			continue

		}

		bytes, err := json.Marshal(logger.LogBuf)
		logger.LogBuf = nil
		if err != nil {
			panic(err)
		} else {
			logger.ticked = false
			logger.pollwait = 0
			logger.ResultChan <- string(bytes)
		}
	}
}

func (l *Logger) logEnter(ltype LogType, s string, a ...any) {
	lstring := strings.TrimSpace(fmt.Sprintf(s, a...))
	l.LogChan <- LogEntry{ltype, lstring}
	//fmt.Printf("§§§ %s: %+v\n", ltype, lstring)
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

func (l *Logger) Tick(n int) {
	l.logEnter(TICKOP, "%d", n)
}

func (l *Logger) Poll() {
	l.logEnter(POLLOP, "")
}
