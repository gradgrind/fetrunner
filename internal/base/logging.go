package base

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

/*
    The logger passes lines to stdout.

All operations cause a OP_START to be logged before doing anything else,
and an OP_END at the end. Any results or other log entries associated
with an operation will be between these two entries.

A TICK is output directly, not as part of an operation,
so it has no OP_END.
*/

var (
	DataBase *BaseData
	logger   *loggerBase
)

func init() {
	DataBase = &BaseData{}
}

type MsgType int

const (
	none MsgType = iota
	INFO
	WARNING
	ERROR
	BUG

	OP_START   = "+++"
	OP_END     = "---"
	OP_LONGRUN = "***"
	OP_QUIT    = "-*-*-"
)

var logType = map[MsgType]string{
	INFO:    "*INFO*",
	WARNING: "*WARNING*",
	ERROR:   "*ERROR*",
	BUG:     "*BUG*",
}

func (ltype MsgType) String() string {
	s, ok := logType[ltype]
	if !ok {
		panic(fmt.Sprintf("Invalid LogType: %d", ltype))
	}
	return s
}

type loggerBase struct {
	ch       chan string
	running  bool
	file     *os.File // set only if logging to file
	ticker   chan string
	stopFlag bool // used to interrupt long-running processes
}

func log(s string) {
	logger.ch <- s
}

func LogTake() string {
	return <-logger.ch
}

func LogWaitTicker() string {
	return <-logger.ticker
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
	logger = &loggerBase{
		// The channel buffer should be large enough for the writer not to be held up.
		ch:     make(chan string, 100),
		ticker: make(chan string),
		file:   logfile,
	}
	go logToFile()
}

func logToFile() {
	// Read from log channel until an OP_QUIT is received, writing the log lines
	// to the output file.
	for {
		line := LogTake()
		logger.file.WriteString(strings.ReplaceAll(line, "||", "\n + ") + "\n")
		if strings.HasPrefix(line, "$ .TICK=") {
			_, t, _ := strings.Cut(line, "=")
			logger.ticker <- t
		}
		if line == OP_QUIT {
			close(logger.ticker)
			break
		}
	}
}

func LogCommand(slist []string) {
	logger.running = true
	log(fmt.Sprintf("%s %s %+v", OP_START, slist[0], slist[1:]))
}

func LogCommandEnd(real_end bool) {
	if real_end {
		logger.running = false
		log(OP_END)
	} else {
		log(OP_LONGRUN)
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

func LogResult(key string, value any) {
	log(fmt.Sprintf("$ %s=%v", key, value))
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
