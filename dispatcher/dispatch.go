package dispatcher

import (
	"encoding/json"
	"fetrunner/autotimetable"
	"fetrunner/base"
	"fetrunner/fet"
	"fetrunner/makefet"
	"fetrunner/timetable"
	"fetrunner/w365tt"
	"path/filepath"
	"strings"
)

var logger0 *base.Logger

func init() {
	// Set up logger.
	logger0 = base.NewLogger()
	go base.LogToBuffer(logger0)
	// Set up default Dispatcher
	DispatcherMap[""] = &Dispatcher{
		BaseData: &base.BaseData{
			Logger: logger0,
		},
	}
}

type Dispatcher struct {
	BaseData *base.BaseData
	TtSource autotimetable.TtSource
	//AutoTtData *autotimetable.AutoTtData
	Running bool
}

type DispatchOp struct {
	Op   string
	Id   string
	Data []string
}

var DispatcherMap map[string]*Dispatcher = map[string]*Dispatcher{}
var OpHandlerMap map[string]func(*Dispatcher, *DispatchOp) = map[string]func(
	*Dispatcher, *DispatchOp){}

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
	dsp, ok := DispatcherMap[op.Id]
	if !ok {
		//TODO
		panic("No instance with Id = " + op.Id)
	}
	f, ok := OpHandlerMap[op.Op]
	if ok {
		// The valid commands are dependent on the run-state of the timetable
		// generation. Those valid when running have a "_" prefix.
		if dsp.Running {
			if op.Op[0] != '_' {
				dsp.BaseData.Logger.Error("!InvalidOp_RunningOp: %s", op.Op)
				return
			}
		} else if op.Op[0] == '_' {
			dsp.BaseData.Logger.Error("!InvalidOp_NotRunningOp: %s", op.Op)
			return
		}

		f(dsp, op)
	} else {
		dsp.BaseData.Logger.Error("!InvalidOp_Op: %s", op.Op)
	}
}

func CheckArgs(l *base.Logger, op *DispatchOp, n int) bool {
	if len(op.Data) != n {
		l.Error("!InvalidOp_Data: %s", op.Op)
		return false
	}
	return true
}

func init() {
	OpHandlerMap["SET_FILE"] = file_loader
	OpHandlerMap["RUN_TT"] = runtt
	OpHandlerMap["_POLL_TT"] = polltt
	OpHandlerMap["_STOP_TT"] = stoptt
}

// Handle (currently) ".fet" and "_w365.json" input files.
func file_loader(dsp *Dispatcher, op *DispatchOp) {
	bd := dsp.BaseData
	logger := bd.Logger
	if !CheckArgs(logger, op, 1) {
		return
	}
	fpath := op.Data[0]

	if strings.HasSuffix(fpath, ".fet") {
		ttRunDataFet := fet.FetRead(bd, fpath)
		if ttRunDataFet != nil {
			dsp.TtSource = ttRunDataFet
			bd.SourceDir = filepath.Dir(fpath)
			n := filepath.Base(fpath)
			bd.Name = n[:len(n)-4]
			logger.Result(op.Op, fpath)
			logger.Result("DATA_TYPE", "FET")
			bd.Db = nil
			return
		}
	} else if strings.HasSuffix(fpath, "_w365.json") {
		db0 := bd.Db
		bd.Db = base.NewDb()
		if w365tt.LoadJSON(bd, fpath) {
			dsp.TtSource = nil
			bd.SourceDir = filepath.Dir(fpath)
			n := filepath.Base(fpath)
			bd.Name = n[:len(n)-10]
			bd.PrepareDb()
			logger.Result(op.Op, fpath)
			logger.Result("DATA_TYPE", "DB")
			return
		}
		bd.Db = db0
	} else {
		logger.Error("LoadFile_InvalidSuffix: %s", fpath)
		return
	}
	logger.Error("LoadFile_InvalidContent: %s", fpath)
}

func runtt(dsp *Dispatcher, op *DispatchOp) {

	//TODO: Handle parameters, if any. Persumably timeout could be
	// one of them.

	if dsp.TtSource == nil {

		//TODO???
		if dsp.BaseData.Db != nil {
			dsp.TtSource = makefet.FetTree(dsp.BaseData, timetable.BasicSetup(dsp.BaseData))
		} else {
			panic("No source")
		}
	}
	// Set up FET back-end and start processing
	attdata := &autotimetable.AutoTtData{
		Source:            dsp.TtSource,
		NActivities:       dsp.TtSource.GetNActivities(),
		NConstraints:      dsp.TtSource.GetNConstraints(),
		ConstraintTypes:   dsp.TtSource.GetConstraintTypes(),
		HardConstraintMap: dsp.TtSource.GetHardConstraintMap(),
		SoftConstraintMap: dsp.TtSource.GetSoftConstraintMap(),
	}
	attdata.SetParameterDefault()

	fet.SetFetBackend(dsp.BaseData, attdata)

	dsp.BaseData.Logger.Result("OK", "")

	// Need an extra goroutine so that this can return immediately.
	//TODO: timeout
	go func() {
		dsp.Running = true
		attdata.StartGeneration(dsp.BaseData, 10)
		dsp.Running = false
	}()
}

func polltt(dsp *Dispatcher, op *DispatchOp) {
	dsp.BaseData.Logger.Poll()
}

// TODO
func stoptt(dsp *Dispatcher, op *DispatchOp) {
	dsp.BaseData.StopFlag = true
}
