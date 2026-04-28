package base

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
)

/*
    The logger uses a buffer to store incoming log lines until they are read using `LogTake()`.

All operations cause an OP_START to be logged before doing anything else,
and an OP_END at the end. Any results or other log entries associated
with an operation will be between these two entries.

A TICK is output directly, not as part of an operation,
so it has no OP_END.
*/

var (
	DataBase *BaseData
	logger   *logBuffer
)

func init() {
	DataBase = &BaseData{}
	logger = &logBuffer{}
	logger.mu.Lock()
	logger.logch = make(chan string, 100)
}

type MsgType int

const (
	none MsgType = iota
	INFO
	WARNING
	ERROR
	BUG

	OP_START = "+++"
	OP_END   = "---"
	OP_QUIT  = "-*-*-"
)

var logType = map[MsgType]string{
	INFO:    "*INFO*",
	WARNING: "*WARNING*",
	ERROR:   "*ERROR*",
	BUG:     "*BUG*",
}

type logBuffer struct {
	mu       sync.Mutex
	logch    chan string
	lines    []string
	index    int
	file     *os.File // set only if logging to file
	running  bool
	stopFlag bool      // used to interrupt long-running processes
	done     chan bool // set only if logging to file
}

func log(line string) {
	logger.logch <- line
}

func LogTake() string {
	return <-logger.logch
}

func LogResult(key string, value any) {
	log(fmt.Sprintf("$ %s=%v", key, value))
}

func LogCommand(slist []string) {
	logger.running = true
	log(fmt.Sprintf("%s %s %+v", OP_START, slist[0], slist[1:]))
}

func LogCommandEnd() {
	log(OP_END)
	logger.running = false
	<-logger.done // wait for the log to catch up
}

func (ltype MsgType) String() string {
	s, ok := logType[ltype]
	if !ok {
		panic(fmt.Sprintf("Invalid LogType: %d", ltype))
	}
	return s
}

func SetStopFlag(on bool) {
	logger.stopFlag = on
}

func LogStop() {
	log(OP_QUIT)
}

func GetStopFlag() bool {
	return logger.stopFlag
}

func LogToFile(logfile *os.File) {
	logger.file = logfile
	logger.done = make(chan bool)
	go logToFile()
}

func logToFile() {
	// Read from log channel until an OP_QUIT is received, writing the log lines
	// to the output file.
	for {
		line := LogTake()
		logger.file.WriteString(strings.ReplaceAll(line, "||", "\n + ") + "\n")
		if line == OP_END {
			logger.done <- true
		} else if line == OP_QUIT {
			break
		}
	}
}

func LogRunning() bool {
	return logger.running
}

func logMessage(ltype MsgType, s string, a ...any) {
	msg := strings.TrimSpace(fmt.Sprintf(ltype.String()+" "+s, a...))
	log(strings.ReplaceAll(msg, "\n", "||"))
}

func LogInfo(s string, a ...any) {
	logMessage(INFO, s, a...)
}

func LogWarning(s string, a ...any) {
	logMessage(WARNING, s, a...)
}

func LogError(s string, a ...any) {
	logMessage(ERROR, s, a...)
}

func LogBug(s string, a ...any) {
	var p string
	_, f, ln, ok := runtime.Caller(1)
	if ok {
		d := filepath.Dir(f)
		fp := filepath.Join(filepath.Base(d), filepath.Base(f))
		p = fmt.Sprintf("%s @ %d: ", fp, ln)
	} else {
		p = "Location?: "
	}
	logMessage(BUG, p+s, a...)
}

func LogTick(n int) {
	LogResult(".TICK", n)
}
