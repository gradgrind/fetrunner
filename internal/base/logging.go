package base

import (
	"encoding/json"
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
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

	OP_START
	OP_END
	TICK
	POLLOP
)

var logType = map[LogType]string{
	INFO:    "*INFO*",
	WARNING: "*WARNING*",
	ERROR:   "*ERROR*",
	BUG:     "*BUG*",
	COMMAND: "#",
	RESULT:  "$",

	POLLOP: "_POLL",

	OP_START: "+++",
	OP_END:   "---",
	TICK:     ".TICK",
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
	Running    bool
	endrun     bool
	Mu         sync.Mutex
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

/*
	Log entry handler adding log entries to a buffer.

Run it as a goroutine.

All operations cause a OP_START to be logged before doing anything else,
and an OP_END at the end. Any results or other log entries associated
with this operation will normally be between these two entries.
The OP_END will cause the buffer contents to be output to the result
channel.
However, for long-running operations it may be desirable to read the log
entries before the operation is completed. This can be managed by letting
the long-running operation start a goroutine for the lengthy part. The
starter part would return, sending its OP_END before the lengthy part is
completed. Then POLLOP operations are initiated periodically to monitor
the progress.
As the polling will be done in a loop, a mechanism is needed to ensure that
the polling is not too rapid. This could be done by a timer in the polling
loop, but as the long-running backend already has a timer, issuing ticks
every second, this is used instead, delaying the return from a POLLOP
operation until a tick has been logged.

A TICK is entered into the buffer directly, not as part of an operation,
so it has no OP_END. Normally it just sets the flag `logger.ticked` to true,
but if a polling operation is awaiting a tick (logger.pollwait = 2) it will
cause the buffer to be read as result, allowing the POLLOP operation to
finish.

A POLLOP operation sets the flag `logger.pollwait` to 1, indicating to the
OP_END handler that the buffer should not be passed to the result channel if
no tick has been registered. If there has not been a tick, the OP_END sets
`logger.pollwait` to 2, so that the next tick will cause it to complete.
*/
func LogToBuffer(logger *Logger) {
	for entry := range logger.LogChan {
		switch entry.Type {

		case TICK:
			logger.LogBuf = append(logger.LogBuf, LogEntry{
				RESULT, TICK.String() + "=" + entry.Text})
			if entry.Text == "-1" {
				logger.endrun = true
			}
			if logger.pollwait != 2 {
				logger.ticked = true
				continue
			}

		case POLLOP:
			//logger.LogBuf = append(logger.LogBuf, entry)
			logger.pollwait = 1
			continue

		case OP_END:
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
			if logger.endrun {
				logger.Running = false
			}
			logger.ticked = false
			logger.pollwait = 0
			logger.ResultChan <- string(bytes)
		}
	}
}

func (l *Logger) StartRun() {
	l.Running = true
	l.endrun = false
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
	l.logEnter(TICK, "%d", n)
}

func (l *Logger) Poll() {
	l.logEnter(POLLOP, "")
}
