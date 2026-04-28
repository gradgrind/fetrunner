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
	buffer   *LogBuffer
	running  bool
	file     *os.File // set only if logging to file
	ticker   chan string
	stopFlag bool // used to interrupt long-running processes
}

func LogFromBuffer(buf *LogBuffer) {
	logger.buffer = buf
}

func log(s string) {
	logger.ch <- s
}

func LogTake() string {
	// Read from a LogBuffer if there is one, until OP_END, then remove it.
	if logger.buffer != nil {
		line := logger.buffer.Take()
		if line == OP_END {
			// buffer finished, remove it
			logger.buffer = nil
		}
		return line
	}
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
	<-logger.ticker // only LogToFile uses this channel
}

func GetStopFlag() bool {
	return logger.stopFlag
}

// The basic logging function, the entries must be read externally using `LogTake()`.
func LogToBuffer() {
	logger = &loggerBase{
		// The channel buffer should be large enough for the writer not to be held up.
		ch: make(chan string, 100),
	}
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
		} else if line == OP_QUIT {
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
	log(formatLogResult(key, value))
}

func formatLogResult(key string, value any) string {
	return fmt.Sprintf("$ %s=%v", key, value)
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

//+++ Could this make the channel buffer unnecessary, or is a circular buffer better there?

// Note that there is no space reclamation here, so don't use this for very
// large data lists.
type LogBuffer struct {
	mu    sync.Mutex
	lines []string
	index int
}

func GetLogBuffer() *LogBuffer {
	buf := &LogBuffer{}
	buf.mu.Lock()
	return buf
}

func (buf *LogBuffer) Add(line string) {
	buf.mu.TryLock() // ensure locked
	buf.lines = append(buf.lines, line)
	buf.mu.Unlock()
}

func (buf *LogBuffer) AddResult(key string, value any) {
	buf.mu.TryLock() // ensure locked
	buf.lines = append(buf.lines, formatLogResult(key, value))
	buf.mu.Unlock()
}

func (buf *LogBuffer) End() {
	buf.Add(OP_END)
	logger.running = false
}

func (buf *LogBuffer) Take() string {
	buf.mu.Lock()
	l := buf.lines[buf.index]
	buf.index++
	if buf.index < len(buf.lines) {
		buf.mu.Unlock() // more items are available
	}
	return l
}

//---
