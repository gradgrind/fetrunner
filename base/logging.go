package base

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
)

//TODO: The logger runs in its own goroutine, communicating via channels.

type LogType int

const (
	none LogType = iota
	INFO
	WARNING
	ERROR
	BUG

	COMMAND
	RESULT
)

var logType = map[LogType]string{
	INFO:    "*INFO*",
	WARNING: "*WARNING*",
	ERROR:   "*ERROR*",
	BUG:     "*BUG*",
	COMMAND: "<",
	RESULT:  ">",
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

type LogInstance struct {
	logbuf     []LogEntry
	resultchan chan []LogEntry
}

//TODO: For the library version (at least) I need to have the log available
// as JSON, perhaps in addition to the file version?
// It would be cleared after each reading, otherwise it could be pretty much
// like the log file. However, the lines would be saved as a list (stripping
// newline characters).

type LogCmd int

const (
	noop LogCmd = iota
	NEW_ENTRY
	GET_LOGS
)

type logcmd struct {
	logger *LogInstance
	cmd    LogCmd
	data   any
}

var logcmdchan chan logcmd

func init() {
	logcmdchan = make(chan logcmd)
	go logreceive()
}

func NewLog() *LogInstance {
	return &LogInstance{
		resultchan: make(chan []LogEntry),
	}
}

func logreceive() {
	var waiting []*LogInstance
	var waiting1 []*LogInstance
	for ld := range logcmdchan {
		l := ld.logger
		switch ld.cmd {

		case NEW_ENTRY:
			entry := ld.data.(LogEntry)
			l.logbuf = append(l.logbuf, entry)

		case GET_LOGS:
			waiting = append(waiting, l)

		}

		waiting1 = nil
		// Read new log entries
		for _, w := range waiting {
			if len(w.logbuf) != 0 {
				w.resultchan <- w.logbuf
				w.logbuf = nil
			} else {
				waiting1 = append(waiting1, w)
			}
		}
		waiting = waiting1

	}
}

// LogToFile allows the log entries to be saved to a file, as they are
// generated.
// Run it as a goroutine.
func LogToFile(logger *LogInstance, logpath string) {
	file, err := os.OpenFile(logpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		panic(err)
	}
	defer file.Close()
	cmd := logcmd{logger, GET_LOGS, nil}
	for {
		logcmdchan <- cmd
		logs, ok := <-logger.resultchan // waits for data to become available
		if !ok {
			// channel closed
			break
		}
		for _, entry := range logs {
			lstring := entry.Type.String() + " " + entry.Text
			file.WriteString(lstring + "\n")
		}
	}
}

func (l *LogInstance) logEnter(ltype LogType, s string, a ...any) {
	lstring := strings.TrimSpace(fmt.Sprintf(s, a...))
	logcmdchan <- logcmd{l, NEW_ENTRY, LogEntry{ltype, lstring}}
}

func (l *LogInstance) Info(s string, a ...any) {
	l.logEnter(INFO, s, a...)
}

func (l *LogInstance) Result(key string, value string) {
	l.logEnter(RESULT, "%s=%s", key, value)
}

func (l *LogInstance) Warning(s string, a ...any) {
	l.logEnter(WARNING, s, a...)
}

func (l *LogInstance) Error(s string, a ...any) {
	l.logEnter(ERROR, s, a...)
}

func (l *LogInstance) Bug(s string, a ...any) {
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

//TODO--
//--------------------------------------------------
/*
// TODO: These are globals, perhaps they should not be, but separate instances
// for each input file?
var (
	CONSOLE bool
	Message *log.Logger
	Warning *log.Logger
	Error   *log.Logger
	Bug     *log.Logger
	result  *log.Logger

	logbuf []string
)

func Result(key string, val any) {
	result.Println(key, "=", val)
}

func OpenLog(logpath string) {
	var file *os.File
	if logpath == "" {
		file = os.Stdout
	} else {
		var err error
		file, err = os.OpenFile(logpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			log.Fatal(err)
		}
	}

	Message = log.New(file, "*INFO* ", 0)
	Warning = log.New(file, "*WARNING* ", 0)
	Error = log.New(file, "*ERROR* ", 0)
	Bug = log.New(file, "*BUG* ", log.Lshortfile)
	result = log.New(file, "+ ", 0)
}
*/
