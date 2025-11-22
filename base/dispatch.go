package base

import (
	"strings"
	"sync"
)

type BufferingLogger struct {
	Logger
	logbuf []LogEntry
	mu     sync.Mutex
}

func NewBufferingLogger() *BufferingLogger {
	return &BufferingLogger{
		Logger: NewLogger(),
	}
}

var bufferinglogger *BufferingLogger

func init() {
	// default logger
	bufferinglogger = NewBufferingLogger()
	go logToBuffer(bufferinglogger)
}

// Log entry handler adding log entries to a buffer.
func logToBuffer(logger *BufferingLogger) {
	for entry := range logger.LogChan {
		logger.mu.Lock()
		logger.logbuf = append(logger.logbuf, entry)
		logger.mu.Unlock()
	}
}

// Read buffered log entries, clear buffer.
// Note that a Go `string` can contain arbitrary bytes, it is essentially
// a read-only slices of bytes. Thus the individual entries – which are
// utf-8 strings – can be joind by an invalid utf-8 byte (0xFF) so that
// they can be split again by the receiver.
func (logger *BufferingLogger) readBuffer() string {
	// Lock as briefly as possible, by copying buffer.
	logger.mu.Lock()
	buf := logger.logbuf
	logger.logbuf = nil
	logger.mu.Unlock()
	entries := []string{}
	for _, entry := range buf {
		lstring := entry.Type.String() + " " + entry.Text
		entries = append(entries, lstring)
	}
	return strings.Join(entries, "\xff")
}

func Dispatch(cmd0 string) string {
	cmdsplit := strings.SplitN(cmd0, " ", 2)
	switch cmd := cmdsplit[0]; cmd {

	case "CONFIG_INIT":
		bufferinglogger.InitConfig()

	case "GET_CONFIG":
		cfg := bufferinglogger.GetConfig(cmdsplit[1])
		bufferinglogger.Result("GET_CONFIG", cfg)

	case "SET_CONFIG":
		bufferinglogger.SetConfig(cmdsplit[1], cmdsplit[2])

	// FET handling
	case "GET_FET":
		bufferinglogger.TestFet()

	default:
		bufferinglogger.Bug("Invalid command: %s", cmd0)

	}

	// Collect the logs as result.
	return bufferinglogger.readBuffer()
}
