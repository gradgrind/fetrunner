package dispatcher

import (
	"encoding/json"
	"fetrunner/autotimetable"
	"fetrunner/base"
	"fetrunner/db"
	"fetrunner/fet"
	"fetrunner/timetable"
	"fetrunner/w365tt"
	"strings"
)

type DispatchOp struct {
	Op   string
	Id   string
	Data []string
}

type FrInstance struct {
	Id     string
	Logger *base.Logger

	//TODO: Which of these are really necessary? Something else?
	Db        *db.DbTopLevel
	TtData    *timetable.TtData
	BasicData *autotimetable.BasicData
}

var frInstanceMap map[string]*FrInstance = map[string]*FrInstance{}
var OpHandlerMap map[string]func(*FrInstance, *DispatchOp) = map[string]func(
	*FrInstance, *DispatchOp){}

func Dispatch(cmd0 string) string {
	var op DispatchOp
	if err := json.Unmarshal([]byte(cmd0), &op); err != nil {
		logger.Error("!InvalidOp_JSON: %s", err)
	} else {
		dispatchOp(&op)
	}
	// At the end of an operation the log entries must be collected.
	// To ensure that none are missed, the logger channels are used to
	// synchronize the accesses.
	logger.LogChan <- base.LogEntry{Type: base.ENDOP}
	return <-logger.ResultChan
}

func dispatchOp(op *DispatchOp) {
	if op.Id == "" {
		// Some ops are only valid using the null Id.
		startOp(logger, op)
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
	fr, ok := frInstanceMap[op.Id]
	if !ok {
		//TODO
		panic("No instance with Id = " + op.Id)
	}
	f, ok := OpHandlerMap[op.Op]
	if ok {
		if fr.Id != "" {
			startOp(fr.Logger, op)
		}
		f(fr, op)
	} else {
		fr.Logger.Error("!InvalidOp_Op: %s", op.Op)
	}
}

func CheckArgs(l *base.Logger, op *DispatchOp, n int) bool {
	if len(op.Data) != n {
		l.Error("!InvalidOp_Data: %s", op.Op)
		return false
	}
	return true
}

func startOp(logger *base.Logger, op *DispatchOp) {
	var text string
	if op.Id != "" {
		text = "[" + op.Id + "] "
	}
	text += op.Op
	if len(op.Data) != 0 {
		text += " (" + strings.Join(op.Data, ", ") + ")"
	}
	logger.LogChan <- base.LogEntry{
		Type: base.STARTOP, Text: text}
}

var logger *base.Logger

func init() {
	// Set up logger.
	logger = base.NewLogger()
	go base.LogToBuffer(logger)
	// Set up default FrInstance
	frInstanceMap[""] = &FrInstance{
		Logger: logger,
	}
}

// Handle (currently) ".fet" and "_w365.json" input files.
func file_loader(fr *FrInstance, op *DispatchOp) {
	logger := fr.Logger
	if !CheckArgs(logger, op, 1) {
		return
	}
	fpath := op.Data[0]

	//TODO: what to do with the data structures produced here?
	// Should they be attached to the logger? That might require an
	// "any" field, not least because the Logger struct is defined in
	// package "base".
	if strings.HasSuffix(fpath, ".fet") {
		bdata := &autotimetable.BasicData{}
		bdata.SetParameterDefault()
		bdata.Logger = logger
		if fet.FetRead(bdata, fpath) {
			logger.Result(op.Op, fpath)
			logger.Result("DATA_TYPE", "FET")
			return
		}
	} else if strings.HasSuffix(fpath, "_w365.json") {
		db0 := db.NewDb(logger)
		if w365tt.LoadJSON(db0, fpath) {
			db0.PrepareDb()
			logger.Result(op.Op, fpath)
			logger.Result("DATA_TYPE", "DB")
			return
		}
	} else {
		logger.Error("LoadFile_InvalidSuffix: %s", fpath)
		return
	}
	logger.Error("LoadFile_InvalidContent: %s", fpath)
}

func init() {
	OpHandlerMap["SET_FILE"] = file_loader
}
