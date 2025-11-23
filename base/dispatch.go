package base

import (
	"encoding/json"
)

var defaultlogger *Logger

// At the end of an operation the log entries must be collected.
// To ensure that none are missed, the logger channels are used to
// synchronize the accesses.
func (l *Logger) OpDone() string {
	l.logchan <- LogEntry{Type: ENDOP}
	return <-l.resultchan
}

func init() {
	// default logger
	defaultlogger = NewLogger()
	go logToBuffer(defaultlogger)
}

// Log entry handler adding log entries to a buffer.
func logToBuffer(logger *Logger) {
	for entry := range logger.logchan {
		if entry.Type == ENDOP {
			bytes, err := json.Marshal(logger.logbuf)
			logger.logbuf = nil
			if err != nil {
				panic(err)
			} else {
				logger.resultchan <- string(bytes)
			}
		} else {
			logger.logbuf = append(logger.logbuf, entry)
		}
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

var LoggerMap map[string]*Logger

type DispatchOp struct {
	Op   string
	Id   string
	Data []string
}

func Dispatch(cmd0 string) string {

	var (
		op     DispatchOp
		logger *Logger
		ok     bool
	)
	logger = defaultlogger
	if err := json.Unmarshal([]byte(cmd0), &op); err != nil {
		logger.Error("!InvalidOp_JSON: %s // %s", err, cmd0)
		goto done
	}

	if op.Id == "" {
		// Some ops are only valid using the default logger:
		switch op.Op {

		case "CONFIG_INIT":
			if logger.checkargs(&op, 0) {
				logger.InitConfig()
			}
			goto done

		case "GET_CONFIG":
			if logger.checkargs(&op, 1) {
				defaultlogger.Result(
					"GET_CONFIG", defaultlogger.GetConfig(op.Data[0]))
			}
			goto done

		case "SET_CONFIG":
			if logger.checkargs(&op, 2) {
				defaultlogger.SetConfig(op.Data[0], op.Data[1])
			}
			goto done

		// FET handling
		case "GET_FET":
			if logger.checkargs(&op, 0) {
				defaultlogger.TestFet()
			}
			goto done

		default:

		}
	} else {
		logger, ok = LoggerMap[op.Id]
		if !ok {
			logger.Error("!InvalidOp_Logger: %s", cmd0)
			goto done
		}
	}

	switch op.Op {

	default:
		logger.Error("!InvalidOp_Op: %s", cmd0)

	}

done:
	return logger.OpDone() // collect the logs as result
}

func (l *Logger) checkargs(op *DispatchOp, n int) bool {
	if len(op.Data) != n {
		l.Error("!InvalidOp_Data: %s", op.Op)
		return false
	}
	return true
}
