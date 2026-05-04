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
    The logger uses a buffer to store incoming log lines until they are read using `logTake()`.

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
	logger = &logBuffer{
		logch: make(chan string, 100),
		done:  make(chan bool),
	}
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
	logch        chan string
	file         *os.File   // set only if logging to file
	buffer       []string   // used only if logging to buffer
	bufReadIndex int        // used only if logging to buffer
	bufmu        sync.Mutex // used only if logging to buffer
	bufmuread    sync.Mutex // used only if logging to buffer
	bufreadready int        // used only if logging to buffer
	running      bool
	stopFlag     bool      // used to interrupt long-running processes
	done         chan bool // set only if logging to file
}

func log(line string) {
	logger.logch <- line
}

func logTake() string {
	return <-logger.logch
}

func LogResult(key string, value any) {
	log(fmt.Sprintf("$ %s=%v", key, value))
}

func LogCommand(cmd string) {
	logger.running = true
	log(OP_START + " " + cmd)
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
	close(logger.logch) // any subsequent `log` calls will panic
}

func GetStopFlag() bool {
	return logger.stopFlag
}

func LogToFile(logfile *os.File) {
	logger.file = logfile
	go logToFile()
}

func logToFile() {
	// Read from log channel until an OP_QUIT is received, writing the log lines
	// to the output file.
	for {
		line := logTake()
		logger.file.WriteString(strings.ReplaceAll(line, "||", "\n + ") + "\n")
		if line == OP_END {
			logger.done <- true
		} else if line == OP_QUIT {
			break
		}
	}
}

func LogToBuffer() {
	logger.bufmuread.Lock()
	logger.bufreadready = 0
	go logToBuffer()
}

func logToBuffer() {
	// Read from log channel until an OP_QUIT is received, writing the log lines
	// to the buffer.
	logger.buffer = nil //TODO?
	logger.bufReadIndex = 0
	for {
		line := logTake()
		logger.bufmu.Lock()
		logger.buffer = append(logger.buffer, line)
		logger.bufreadready++
		if logger.bufreadready == 1 {
			logger.bufmuread.Unlock()
		}
		logger.bufmu.Unlock()

		//*
		if line == OP_END {
			logger.done <- true
		} else if line == OP_QUIT {
			break
		}
		//*/
	}
}

//TODO: It might well be a good idea to reset the buffer sometimes ...

func ReadLogBufferLine() string {
	logger.bufmuread.Lock()

	//TODO: What if logToBuffer happens here?
	// Surely it can only unlock the mutex if bufreadready == 0, which should be
	// impossible, because then bufmuread wouldn't have been unlocked.

	logger.bufmu.Lock()
	line := logger.buffer[logger.bufReadIndex]
	logger.bufReadIndex++
	logger.bufreadready--
	if logger.bufreadready != 0 {
		logger.bufmuread.Unlock()
	}
	logger.bufmu.Unlock()

	/* Placing this here seemed more logical (?), but it blocks somehow
	if line == OP_END {
		logger.done <- true
	} else if line == OP_QUIT {

	}
	*/

	return line
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
