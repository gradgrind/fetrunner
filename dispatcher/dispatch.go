package dispatcher

import (
	"encoding/json"
	"fetrunner/base"
)

var OpHandlerMap map[string]func(*base.Logger, *DispatchOp) = map[string]func(
	*base.Logger, *DispatchOp){}

type DispatchOp struct {
	Op   string
	Id   string
	Data []string
}

func Dispatch(logger *base.Logger, cmd0 string) string {
	logger.LogChan <- base.LogEntry{Type: base.STARTOP, Text: cmd0}
	var op DispatchOp
	if err := json.Unmarshal([]byte(cmd0), &op); err != nil {
		logger.Error("!InvalidOp_JSON: %s", err)
	} else {
		dispatchOp(logger, &op)
	}
	// At the end of an operation the log entries must be collected.
	// To ensure that none are missed, the logger channels are used to
	// synchronize the accesses.
	logger.LogChan <- base.LogEntry{Type: base.ENDOP}
	return <-logger.ResultChan
}

func dispatchOp(logger *base.Logger, op *DispatchOp) {
	if op.Id == "" {
		// Some ops are only valid using the null Id.
		switch op.Op {

		case "CONFIG_INIT":
			if CheckArgs(logger, op, 0) {
				logger.InitConfig()
			}
			return

		case "GET_CONFIG":
			if CheckArgs(logger, op, 1) {
				key := op.Data[0]
				logger.Result(key, logger.GetConfig(key))
			}
			return

		case "SET_CONFIG":
			if CheckArgs(logger, op, 2) {
				logger.SetConfig(op.Data[0], op.Data[1])
			}
			return

		// FET handling
		case "GET_FET":
			if CheckArgs(logger, op, 0) {
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

func CheckArgs(l *base.Logger, op *DispatchOp, n int) bool {
	if len(op.Data) != n {
		l.Error("!InvalidOp_Data: %s", op.Op)
		return false
	}
	return true
}
