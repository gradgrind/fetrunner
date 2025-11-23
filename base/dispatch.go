package base

import (
	"encoding/json"
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

// Read buffered log entries to a JSON array, clear buffer.
func (logger *BufferingLogger) readBuffer() string {
	// Lock as briefly as possible, by copying buffer.
	logger.mu.Lock()
	buf := logger.logbuf
	logger.logbuf = nil
	logger.mu.Unlock()
	bytes, err := json.Marshal(buf)
	if err != nil {
		panic(err)
	} else {
		return string(bytes)
	}
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

var LoggerMap map[string]*BufferingLogger

type DispatchOp struct {
	Op   string
	Id   string
	Data []string
}

func Dispatch(cmd0 string) string {

	var (
		op     DispatchOp
		logger *BufferingLogger
		ok     bool
	)
	if err := json.Unmarshal([]byte(cmd0), &op); err != nil {
		bufferinglogger.Error("!InvalidOp_JSON: %s // %s", err, cmd0)
		goto done
	}
	if op.Op == "" {
		bufferinglogger.Error("!InvalidOp_NoOp: %s", cmd0)
		goto done
	}
	if op.Id == "" {
		// Some ops are only valid using the base logger:
		switch op.Op {

		case "CONFIG_INIT":
			if checkargs(bufferinglogger, &op, 0) {
				bufferinglogger.InitConfig()
			}
			goto done

		case "GET_CONFIG":
			if checkargs(bufferinglogger, &op, 1) {
				bufferinglogger.Result(
					"GET_CONFIG", bufferinglogger.GetConfig(op.Data[0]))
			}
			goto done

		case "SET_CONFIG":
			if checkargs(bufferinglogger, &op, 2) {
				bufferinglogger.SetConfig(op.Data[0], op.Data[1])
			}
			goto done

		// FET handling
		case "GET_FET":
			if checkargs(bufferinglogger, &op, 0) {
				bufferinglogger.TestFet()
			}
			goto done

		default:
			logger = bufferinglogger

		}
	} else {
		logger, ok = LoggerMap[op.Id]
		if !ok {
			bufferinglogger.Error("!InvalidOp_Logger: %s", cmd0)
			goto done
		}
	}

	switch op.Op {

	default:
		logger.Error("!InvalidOp_Op: %s", cmd0)

	}

done:
	// Collect the logs as result.
	return bufferinglogger.readBuffer()
}

func checkargs(logger *BufferingLogger, op *DispatchOp, n int) bool {
	if len(op.Data) != n {
		logger.Error("!InvalidOp_Data: %s", op.Op)
		return false
	}
	return true
}
