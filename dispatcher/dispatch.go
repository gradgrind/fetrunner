package dispatcher

import (
	"encoding/json"
	"fetrunner/autotimetable"
	"fetrunner/base"
	"fetrunner/db"
	"fetrunner/fet"
	"fetrunner/w365tt"
	"path/filepath"
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

	Name       string // stem of source file name (i.e. without ending)
	WorkingDir string // directory of source file

	Db        *db.DbTopLevel
	BasicData *autotimetable.BasicData
}

var frInstanceMap map[string]*FrInstance = map[string]*FrInstance{}
var OpHandlerMap map[string]func(*FrInstance, *DispatchOp) = map[string]func(
	*FrInstance, *DispatchOp){}

func Dispatch(cmd0 string) string {
	var op DispatchOp
	if err := json.Unmarshal([]byte(cmd0), &op); err != nil {
		logger0.Error("!InvalidOp_JSON: %s", err)
	} else {
		dispatchOp(&op)
	}
	// At the end of an operation the log entries must be collected.
	// To ensure that none are missed, the logger channels are used to
	// synchronize the accesses.
	logger0.LogChan <- base.LogEntry{Type: base.ENDOP}
	return <-logger0.ResultChan
}

func dispatchOp(op *DispatchOp) {
	if op.Id == "" {
		// Some ops are only valid using the null Id.
		startOp(logger0, op)
		switch op.Op {

		case "CONFIG_INIT":
			if CheckArgs(logger0, op, 0) {
				logger0.InitConfig()
			}
			return

		case "GET_CONFIG":
			if CheckArgs(logger0, op, 1) {
				key := op.Data[0]
				logger0.Result(key, logger0.GetConfig(key))
			}
			return

		case "SET_CONFIG":
			if CheckArgs(logger0, op, 2) {
				logger0.SetConfig(op.Data[0], op.Data[1])
			}
			return

		// FET handling
		case "GET_FET":
			if CheckArgs(logger0, op, 0) {
				logger0.TestFet()
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

var logger0 *base.Logger

func init() {
	// Set up logger.
	logger0 = base.NewLogger()
	go base.LogToBuffer(logger0)
	// Set up default FrInstance
	frInstanceMap[""] = &FrInstance{
		Logger: logger0,
	}
}

// Handle (currently) ".fet" and "_w365.json" input files.
func file_loader(fr *FrInstance, op *DispatchOp) {
	logger := fr.Logger
	if !CheckArgs(logger, op, 1) {
		return
	}
	fpath := op.Data[0]

	if strings.HasSuffix(fpath, ".fet") {
		bdata := &autotimetable.BasicData{}
		bdata.SetParameterDefault()
		bdata.Logger = logger
		if fet.FetRead(bdata, fpath) {
			fr.WorkingDir = filepath.Dir(fpath)
			n := filepath.Base(fpath)
			fr.Name = n[:len(n)-4]
			logger.Result(op.Op, fpath)
			logger.Result("DATA_TYPE", "FET")
			fr.BasicData = bdata
			fr.Db = nil
			return
		}
	} else if strings.HasSuffix(fpath, "_w365.json") {
		db0 := db.NewDb(logger)
		if w365tt.LoadJSON(db0, fpath) {
			fr.WorkingDir = filepath.Dir(fpath)
			n := filepath.Base(fpath)
			fr.Name = n[:len(n)-10]
			db0.PrepareDb()
			logger.Result(op.Op, fpath)
			logger.Result("DATA_TYPE", "DB")
			fr.Db = db0
			fr.BasicData = nil
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
