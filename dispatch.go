package fetrunner

import (
	"encoding/json"
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fetrunner/internal/fet"
	"fetrunner/internal/makefet"
	"fetrunner/internal/timetable"
	"fetrunner/internal/w365tt"
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
}

type DispatchOp struct {
	Op   string
	Id   string
	Data []string
}

var DispatcherMap map[string]*Dispatcher = map[string]*Dispatcher{}
var OpHandlerMap map[string]func(*Dispatcher, *DispatchOp) = map[string]func(
	*Dispatcher, *DispatchOp){}

// Read a command from JSON, logs a STARTOP for it, look up the corresponding
// function in `OpHandlerMap` and call it. On return log an ENDOP and return
// the data from the result channel. Each logger has a mutex to avoid calling
// Dispatch on it from more than one thread simultaneously (which should not
// happen normally anyway).
func Dispatch(cmd0 string) string {
	logger := Logger0
	var op DispatchOp
	if err := json.Unmarshal([]byte(cmd0), &op); err != nil {
		logger.Mu.Lock()
		defer logger.Mu.Unlock()
		logger.Error("!InvalidOp_JSON: %s", err)
	} else {
		dsp, ok := DispatcherMap[op.Id]
		if !ok {
			//TODO
			panic("No instance with Id = " + op.Id)
		}
		logger = dsp.BaseData.Logger
		logger.Mu.Lock()
		defer logger.Mu.Unlock()
		opLog(logger, &op)
		f, ok := OpHandlerMap[op.Op]
		if ok {
			// The valid commands are dependent on the run-state of the timetable
			// generation. Those valid when running have a "_" prefix.
			if logger.Running {
				if op.Op[0] != '_' {
					panic("!InvalidOp_Running: " + op.Op)
					//goto opdone
				}
			} else if op.Op[0] == '_' {
				panic("!InvalidOp_NotRunning: " + op.Op)
				//goto opdone
			}

			f(dsp, &op)
		} else {
			logger.Error("!InvalidOp_Op: %s", op.Op)
		}
	}
	//opdone:
	// At the end of an operation the log entries must be collected.
	// To ensure that none are missed, the logger channels are used to
	// synchronize the accesses.
	logger.LogChan <- base.LogEntry{Type: base.OP_END}
	return <-logger.ResultChan
}

func opLog(logger *base.Logger, op *DispatchOp) {
	logger.LogChan <- base.LogEntry{Type: base.OP_START,
		Text: fmt.Sprintf("%s %+v", op.Op, op.Data)}
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
	OpHandlerMap["HARD_CONSTRAINTS"] = hardConstraints
	OpHandlerMap["SOFT_CONSTRAINTS"] = softConstraints
	OpHandlerMap["RUN_TT"] = runtt
	OpHandlerMap["_POLL_TT"] = polltt
	OpHandlerMap["_STOP_TT"] = stoptt
	OpHandlerMap["RESULT_TT"] = ttresult

	OpHandlerMap["TT_PARAMETER"] = ttparameter
	OpHandlerMap["TMP_PATH"] = set_tmp
	OpHandlerMap["N_PROCESSES"] = nprocesses
}

func set_tmp(dsp *Dispatcher, op *DispatchOp) {
	if CheckArgs(dsp.BaseData.Logger, op, 1) {
		base.TEMPORARY_BASEDIR = op.Data[0]
		dsp.BaseData.SetTmpDir()
	}
}

// Check path to `fet-cl` (Windows: `fet-clw.exe`) and get FET version.
func get_fet(dsp *Dispatcher, op *DispatchOp) {
	logger := dsp.BaseData.Logger
	if CheckArgs(logger, op, 1) {
		fetpath := op.Data[0]
		if fetpath == "-" {
			fetpath = fet.FET_CL
		}
		fet.FETPATH = fetpath
		cmd := exec.Command(fetpath, "--version")
		out, err := cmd.CombinedOutput()
		if err != nil {
			logger.Warning("FET_NOT_FOUND: %s", err)
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
	bdata := dsp.BaseData
	logger := bdata.Logger
	ttsource := dsp.TtSource
	if ttsource == nil {
		if dsp.BaseData.Db != nil {
			ttsource = makefet.FetTree(bdata, timetable.BasicSetup(bdata))
		} else {
			logger.Error("No source")
			logger.Result("OK", "false")
			return
		}
	}
	if logger.Running {
		panic("Attempt to start generation when already running")
	}
	// Set up FET back-end and start processing
	attdata := &autotimetable.AutoTtData{
		Parameters:        dsp.TtParameters,
		Source:            ttsource,
		NActivities:       ttsource.GetNActivities(),
		NConstraints:      ttsource.GetNConstraints(),
		ConstraintTypes:   ttsource.GetConstraintTypes(),
		HardConstraintMap: ttsource.GetHardConstraintMap(),
		SoftConstraintMap: ttsource.GetSoftConstraintMap(),
	}
	dsp.AutoTtData = attdata
	fet.InitBackend(bdata, attdata)
	logger.Result("OK", "true")
}

func runtt(dsp *Dispatcher, op *DispatchOp) {
	// Need an extra goroutine so that this can return immediately.
	dsp.BaseData.Logger.StartRun()
	go dsp.AutoTtData.StartGeneration(dsp.BaseData)
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
	//TODO: At present the JSON result is generated automatically as a
	// file. It might be preferable to return the data as a string result
	// instead.
	_ = result
}

// Set a parameter for autotimetable.
func ttparameter(dsp *Dispatcher, op *DispatchOp) {
	logger := dsp.BaseData.Logger
	key := op.Data[0]
	val := op.Data[1]
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

	default:
		logger.Error("UnknownParameter: %s", key)
		return
	}

	dsp.BaseData.Logger.Result(key, val)
}

func nprocesses(dsp *Dispatcher, op *DispatchOp) {
	nmin, np, nopt := autotimetable.MinNpOptProcesses()
	dsp.BaseData.Logger.Result(op.Op, fmt.Sprintf("%d.%d.%d", nmin, np, nopt))
}

func hardConstraints(dsp *Dispatcher, op *DispatchOp) {
	for c, ilist := range dsp.AutoTtData.HardConstraintMap {
		dsp.BaseData.Logger.Result(c, strconv.Itoa(len(ilist)))
	}
}

func softConstraints(dsp *Dispatcher, op *DispatchOp) {
	for c, ilist := range dsp.AutoTtData.SoftConstraintMap {
		dsp.BaseData.Logger.Result(c, strconv.Itoa(len(ilist)))
	}
}
