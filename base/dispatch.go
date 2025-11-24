package base

import (
	"encoding/json"
)

var OpHandlerMap map[string]func(*Logger, *DispatchOp) = map[string]func(*Logger, *DispatchOp){}

// Log entry handler adding log entries to a buffer.
func LogToBuffer(logger *Logger) {
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

type DispatchOp struct {
	Op   string
	Id   string
	Data []string
}

func Dispatch(logger *Logger, cmd0 string) string {
	logger.logchan <- LogEntry{Type: STARTOP, Text: cmd0}
	var op DispatchOp
	if err := json.Unmarshal([]byte(cmd0), &op); err != nil {
		logger.Error("!InvalidOp_JSON: %s", err)
	} else {
		dispatchOp(logger, &op)
	}
	// At the end of an operation the log entries must be collected.
	// To ensure that none are missed, the logger channels are used to
	// synchronize the accesses.
	logger.logchan <- LogEntry{Type: ENDOP}
	return <-logger.resultchan
}

func dispatchOp(logger *Logger, op *DispatchOp) {
	if op.Id == "" {
		// Some ops are only valid using the null Id.
		switch op.Op {

		case "CONFIG_INIT":
			if logger.CheckArgs(op, 0) {
				logger.InitConfig()
			}
			return

		case "GET_CONFIG":
			if logger.CheckArgs(op, 1) {
				key := op.Data[0]
				logger.Result(key, logger.GetConfig(key))
			}
			return

		case "SET_CONFIG":
			if logger.CheckArgs(op, 2) {
				logger.SetConfig(op.Data[0], op.Data[1])
			}
			return

		// FET handling
		case "GET_FET":
			if logger.CheckArgs(op, 0) {
				logger.TestFet()
			}
			return

		default:

		}
	}

	// Now deal with the other ops, which can (in principle) be used with any Id.
	f, ok := OpHandlerMap[op.Op]
	if ok {
		f(logger, op)
	} else {
		logger.Error("!InvalidOp_Op: %s", op.Op)
	}
}

func (l *Logger) CheckArgs(op *DispatchOp, n int) bool {
	if len(op.Data) != n {
		l.Error("!InvalidOp_Data: %s", op.Op)
		return false
	}
	return true
}
