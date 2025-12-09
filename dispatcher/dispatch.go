package dispatcher

import (
	"encoding/json"
	"fetrunner/autotimetable"
	"fetrunner/base"
	"fetrunner/fet"
	"fetrunner/makefet"
	"fetrunner/timetable"
	"fetrunner/w365tt"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var Logger0 *base.Logger

func init() {
	// Set up logger.
	Logger0 = base.NewLogger()
	go base.LogToBuffer(Logger0)
	// Set up default Dispatcher
	DispatcherMap[""] = &Dispatcher{
		TtParameters: autotimetable.DefaultParameters(),
		BaseData: &base.BaseData{
			Logger: Logger0,
		},
	}
}

type Dispatcher struct {
	BaseData     *base.BaseData
	TtSource     autotimetable.TtSource
	AutoTtData   *autotimetable.AutoTtData
	TtParameters *autotimetable.Parameters
	Running      bool
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
	logger := Logger0
	var op DispatchOp
	if err := json.Unmarshal([]byte(cmd0), &op); err != nil {
		logger.Error("!InvalidOp_JSON: %s", err)
	} else {
		logger = dispatchOp(&op)
	}
	// At the end of an operation the log entries must be collected.
	// To ensure that none are missed, the logger channels are used to
	// synchronize the accesses.
	logger.LogChan <- base.LogEntry{Type: base.ENDOP}
	return <-logger.ResultChan
}

func opLog(logger *base.Logger, op *DispatchOp) {
	logger.LogChan <- base.LogEntry{Type: base.STARTOP,
		Text: fmt.Sprintf("%s %+v", op.Op, op.Data)}
}

func dispatchOp(op *DispatchOp) *base.Logger {
	/*TODO--
	if op.Id == "" {
		// Some ops are only valid using the null Id.
		switch op.Op {

			case "CONFIG_INIT":
				opLog(Logger0, op)
				if CheckArgs(Logger0, op, 0) {
					Logger0.InitConfig()
				}
				return Logger0

			case "GET_CONFIG":
				if CheckArgs(Logger0, op, 1) {
					key := op.Data[0]
					Logger0.Result(key, Logger0.GetConfig(key))
				}
				return Logger0

			case "SET_CONFIG":
				if CheckArgs(Logger0, op, 2) {
					Logger0.SetConfig(op.Data[0], op.Data[1])
				}
				return Logger0

		// FET handling
		case "GET_FET":
			if CheckArgs(Logger0, op, 0) {
				Logger0.TestFet()
			}
			return Logger0

		default:

		}
	}

	// Now deal with the other ops, which can (in principle) be used with any Id.
	*/

	dsp, ok := DispatcherMap[op.Id]
	if !ok {
		//TODO
		panic("No instance with Id = " + op.Id)
	}
	logger := dsp.BaseData.Logger
	opLog(logger, op)
	f, ok := OpHandlerMap[op.Op]
	if ok {
		// The valid commands are dependent on the run-state of the timetable
		// generation. Those valid when running have a "_" prefix.
		if dsp.Running {
			if op.Op[0] != '_' {
				logger.Error("!InvalidOp_Running: %s", op.Op)
				return logger
			}
		} else if op.Op[0] == '_' {
			logger.Error("!InvalidOp_NotRunning: %s", op.Op)
			return logger
		}

		f(dsp, op)
	} else {
		logger.Error("!InvalidOp_Op: %s", op.Op)
	}
	return logger
}

func CheckArgs(l *base.Logger, op *DispatchOp, n int) bool {
	if len(op.Data) != n {
		l.Error("!InvalidOp_Data: %s", op.Op)
		return false
	}
	return true
}

func init() {
	OpHandlerMap["GET_FET"] = get_fet
	OpHandlerMap["SET_FILE"] = file_loader
	OpHandlerMap["RUN_TT_SOURCE"] = runtt_source
	OpHandlerMap["RUN_TT"] = runtt
	OpHandlerMap["_POLL_TT"] = polltt
	OpHandlerMap["_STOP_TT"] = stoptt
	OpHandlerMap["RESULT_TT"] = ttresult

	OpHandlerMap["TT_PARAMETER"] = ttparameter
}

// Check path to `fet-cl` and get FET version.
func get_fet(dsp *Dispatcher, op *DispatchOp) {
	logger := dsp.BaseData.Logger
	if CheckArgs(logger, op, 0) {
		fetpath := dsp.TtParameters.FETPATH
		cmd := exec.Command(fetpath, "--version")
		out, err := cmd.CombinedOutput()
		if err != nil {
			//TODO?
			logger.Error("FET_NOT_FOUND %s // %s", fetpath, out)
			return
		}
		version := regexp.MustCompile(`(?m)version +([0-9.]+)`)
		match := version.FindSubmatch(out)
		if match == nil {
			logger.Result("FET_VERSION", "?")
		} else {
			logger.Result("FET_VERSION", string(match[1]))
		}
	}
}

// Handle (currently) ".fet" and "_w365.json" input files.
func file_loader(dsp *Dispatcher, op *DispatchOp) {
	bd := dsp.BaseData
	logger := bd.Logger
	if !CheckArgs(logger, op, 1) {
		return
	}
	fpath := op.Data[0]

	if strings.HasSuffix(strings.ToLower(fpath), ".fet") {
		ttRunDataFet := fet.FetRead(bd, fpath)
		if ttRunDataFet != nil {
			dsp.TtSource = ttRunDataFet
			bd.SourceDir = filepath.Dir(fpath)
			n := filepath.Base(fpath)
			bd.Name = strings.TrimSuffix(n, filepath.Ext(n))
			logger.Result(op.Op, fpath)
			logger.Result("DATA_TYPE", "FET")
			bd.Db = nil
			return
		}
	} else if strings.HasSuffix(strings.ToLower(fpath), "_w365.json") {
		db0 := bd.Db
		bd.Db = base.NewDb()
		if w365tt.LoadJSON(bd, fpath) {
			dsp.TtSource = nil
			bd.SourceDir = filepath.Dir(fpath)
			n := filepath.Base(fpath)
			bd.Name = strings.TrimSuffix(n, filepath.Ext(n))
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

// `runtt_source` must be run before `runtt` to ensure that there is source data.
func runtt_source(dsp *Dispatcher, op *DispatchOp) {
	if dsp.TtSource == nil {
		if dsp.BaseData.Db != nil {
			dsp.TtSource = makefet.FetTree(dsp.BaseData, timetable.BasicSetup(dsp.BaseData))
		} else {
			dsp.BaseData.Logger.Error("No source")
			dsp.BaseData.Logger.Result("OK", "false")
			return
		}
	}
	dsp.BaseData.Logger.Result("OK", "true")
}

func runtt(dsp *Dispatcher, op *DispatchOp) {
	if dsp.Running {
		panic("Attempt to start generation when already running")
	}
	// Set up FET back-end and start processing
	attdata := &autotimetable.AutoTtData{
		Parameters:        dsp.TtParameters,
		Source:            dsp.TtSource,
		NActivities:       dsp.TtSource.GetNActivities(),
		NConstraints:      dsp.TtSource.GetNConstraints(),
		ConstraintTypes:   dsp.TtSource.GetConstraintTypes(),
		HardConstraintMap: dsp.TtSource.GetHardConstraintMap(),
		SoftConstraintMap: dsp.TtSource.GetSoftConstraintMap(),
	}

	dsp.AutoTtData = attdata

	fet.SetFetBackend(dsp.BaseData, attdata)

	dsp.BaseData.Logger.Result("OK", "true")

	// Need an extra goroutine so that this can return immediately.
	dsp.Running = true
	go func() {
		attdata.StartGeneration(dsp.BaseData)
		dsp.Running = false
	}()
}

func polltt(dsp *Dispatcher, op *DispatchOp) {
	dsp.BaseData.Logger.Poll()
}

func stoptt(dsp *Dispatcher, op *DispatchOp) {
	dsp.BaseData.StopFlag = true
}

// Get the result data as a JSON string.
func ttresult(dsp *Dispatcher, op *DispatchOp) {
	result := dsp.AutoTtData.GetLastResult()
	//TODO
	_ = result
}

// Set a parameter for autotimetable.
func ttparameter(dsp *Dispatcher, op *DispatchOp) {
	logger := dsp.BaseData.Logger
	key := op.Data[0]
	val := op.Data[1]
	//TODO
	switch key {

	case "TIMEOUT":
		n, err := strconv.Atoi(val)
		if err != nil {
			logger.Error("BadNumber: %s=%s", key, val)
			return
		} else {
			dsp.TtParameters.TIMEOUT = n
		}

	case "MAXPROCESSES":
		n, err := strconv.Atoi(val)
		if err != nil {
			logger.Error("BadNumber: %s=%s", key, val)
			return
		} else {
			dsp.TtParameters.MAXPROCESSES = autotimetable.MaxProcesses(n)
			val = strconv.Itoa(dsp.TtParameters.MAXPROCESSES)
		}

	case "DEBUG":
		dsp.TtParameters.DEBUG = (val == "true")

	case "TESTING":
		dsp.TtParameters.TESTING = (val == "true")

	case "SKIP_HARD":
		dsp.TtParameters.SKIP_HARD = (val == "true")

	case "FETPATH":
		dsp.TtParameters.FETPATH = val
	}

	dsp.BaseData.Logger.Result(key, val)
}
