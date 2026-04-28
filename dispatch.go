package fetrunner

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fetrunner/internal/fet"
	"fetrunner/internal/timetable"
	"fetrunner/internal/w365tt"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
)

//TODO: Note that at present the commands actually used have a very simple structure.
// They have 0, 1 or two arguments, which consist of fairly short strings. No complex
// data is passed in. So a simple text line should be quite enough. I can split this
// to build a DispatchOp structure.
// If this is now such that Dispatch returns only when the command is completed, it
// should probably return a boolean indicating whether the command completed
// successfully, whatever than means. Perhaps with no errors logged? That seems to
// be the case with the old version, returning `true` if no errors.

type DispatchOp struct {
	Op   string
	Data []string
}

var OpHandlerMap map[string]func(*DispatchOp) = map[string]func(*DispatchOp){}

// A command is supplied as a string with separator "|". The first item is the
// command name, which is looked up in `OpHandlerMap`. This provides the actual
// function to call. Normally a command will only be accepted when no other is
// running, but to handle long-running commands there is also the possibility of
// using a command beginning with "_", which will be accepted when another
// command is running.
func Dispatch(cmd0 string) {
	slist := strings.Split(cmd0, "|")
	op := DispatchOp{Op: slist[0], Data: slist[1:]}
	f, ok := OpHandlerMap[op.Op]
	if ok {
		if op.Op[0] == '_' {
			// Don't log this command.
			f(&op)
		} else {
			if base.LogRunning() {
				// Don't log this command.
				panic("!InvalidOp_Running: " + op.Op)
			}
			base.LogCommand(slist)
			f(&op)
			base.LogCommandEnd()
		}
	} else {
		panic("!InvalidOp: " + op.Op)
	}
}

func opLog(op *DispatchOp) {
	fmt.Printf("%s %s %+v", base.OP_START, op.Op, op.Data)
}

func CheckArgs(op *DispatchOp, n int) bool {
	if len(op.Data) != n {
		base.LogError("--INVALID_OP_OP %s", op.Op)
		return false
	}
	return true
}

func init() {
	OpHandlerMap["VERSION"] = fetrunner_version
	OpHandlerMap["GET_FET"] = get_fet
	OpHandlerMap["SET_FILE"] = file_loader
	OpHandlerMap["RUN_TT_SOURCE"] = runtt_source
	OpHandlerMap["TT_PRIORITY_CONSTRAINT_TYPES"] = priortityConstraints
	OpHandlerMap["TT_HARD_CONSTRAINTS"] = hardConstraints
	OpHandlerMap["TT_SOFT_CONSTRAINTS"] = softConstraints
	OpHandlerMap["TT_NACTIVITIES"] = nActivities
	OpHandlerMap["RUN_TT"] = runtt
	OpHandlerMap["_STOP_TT"] = stoptt
	OpHandlerMap["RESULT_TT"] = ttresult

	OpHandlerMap["TT_PARAMETER"] = ttparameter
	OpHandlerMap["TMP_PATH"] = set_tmp
	OpHandlerMap["N_PROCESSES"] = nprocesses
}

func fetrunner_version(op *DispatchOp) {
	if CheckArgs(op, 0) {
		base.LogResult("FETRUNNER_VERSION", VERSION)
	}
}

func set_tmp(op *DispatchOp) {
	if CheckArgs(op, 1) {
		base.TEMPORARY_BASEDIR = op.Data[0]
		base.DataBase.SetTmpDir()
	}
}

func check_fet(fetpath string) bool {
	cmd := exec.Command(fetpath, "--version")
	out, err := cmd.CombinedOutput()
	if err != nil {
		base.LogWarning("--FET_NOT_FOUND %s // %s", err, string(out))
		return false
	}
	base.LogResult("FET_PATH", fetpath)
	version := regexp.MustCompile(`(?m)version +([0-9.]+)`)
	match := version.FindSubmatch(out)
	if match == nil {
		base.LogResult("FET_VERSION", "?")
	} else {
		base.LogResult("FET_VERSION", string(match[1]))
	}
	return true
}

// Check path to FET command-line executable and get FET version.
func get_fet(op *DispatchOp) {
	if CheckArgs(op, 2) {
		fetpath := op.Data[0]
		if fetpath == "" {
			// Get the bare command without path.
			if op.Data[1] == "" {
				fetpath = fet.FET_CL // command-line version
			} else {
				// For fetrunner GUI version: in Windows a special version of fet-cl
				// without console pop-up is required.
				fetpath = fet.FET_CLW
			}
			// First try the directory containing the running executable.
			p, err := os.Executable()
			if err == nil {
				// Sanitize path
				p, err = filepath.EvalSymlinks(p)
				if err == nil {
					fetpath0 := filepath.Join(filepath.Dir(p), fetpath)
					if check_fet(fetpath0) {
						fet.FETPATH = fetpath0
						return
					}
				}
			}
			// Otherwise, try the bare command in case it is in the PATH.
		}
		if check_fet(fetpath) {
			fet.FETPATH = fetpath
		}
	}
}

// Handle (currently) ".fet" and "_w365.json" input files.
func file_loader(op *DispatchOp) {
	bd := base.DataBase
	if !CheckArgs(op, 1) {
		return
	}
	fpath := op.Data[0]

	if strings.HasSuffix(strings.ToLower(fpath), ".fet") {
		source_fet := fet.FetRead(bd, fpath)
		if source_fet != nil {
			bd.Source = source_fet
			bd.SourceDir = filepath.Dir(fpath)
			n := filepath.Base(fpath)
			bd.Name = strings.TrimSuffix(n, filepath.Ext(n))
			base.LogResult(op.Op, fpath)
			base.LogResult("DATA_TYPE", "FET")
			bd.Db = nil
			return
		}
	} else if strings.HasSuffix(strings.ToLower(fpath), "_w365.json") {
		db0 := bd.Db // save old Db in case loading of new data fails
		bd.Db = base.NewDb()
		if w365tt.LoadJSON(fpath) {
			bd.Source = &base.SourceDB{}
			bd.SourceDir = filepath.Dir(fpath)
			n := filepath.Base(fpath)
			bd.Name = strings.TrimSuffix(n, filepath.Ext(n))
			base.PrepareDb()
			base.LogResult(op.Op, fpath)
			base.LogResult("DATA_TYPE", "DB")
			return
		}
		bd.Db = db0
	} else {
		base.LogError("--LOAD_FILE_INVALID_SUFFIX %s", fpath)
		return
	}
	base.LogError("--LOAD_FILE_INVALID_CONTENT %s", fpath)
}

// `runtt_source` must be run before `runtt` to ensure that there is source data.
func runtt_source(op *DispatchOp) {
	if CheckArgs(op, 0) {
		bdata := base.DataBase
		//if logger.Running {
		//  panic("Attempt to start generation when already running")
		//}
		if bdata.Source == nil {
			base.LogError("--NO_SOURCE")
			base.LogResult("OK", "false")
			return
		}
		var ttsource autotimetable.TtSource
		switch stype := bdata.Source.SourceType(); stype {
		case "DB":
			ttsource = timetable.MakeTimetableData()
		case "FET":
			ttsource = bdata.Source.(*fet.TtSourceFet)
		default:
			panic("Unknown source type: " + stype)
		}
		// Set up FET back-end and start processing
		hcmap, scmap := ttsource.GetConstraintMaps()
		attdata := &autotimetable.AutoTtData{
			Source:            ttsource,
			NActivities:       len(ttsource.GetActivities()),
			NConstraints:      len(ttsource.GetConstraints()),
			Constraint_Types:  ttsource.GetConstraintTypes(),
			HardConstraintMap: hcmap,
			SoftConstraintMap: scmap,
		}
		autotimetable.AutoTt = attdata
		base.LogResult("OK", "true")
	}
}

func runtt(op *DispatchOp) {
	if CheckArgs(op, 0) {
		switch autotimetable.TtParameters.BACKEND {
		case "", "FET":
			fet.InitBackend(autotimetable.AutoTt)
		default:
			panic("Unsupported timetable-generation back-end: " + autotimetable.TtParameters.BACKEND)
		}

		autotimetable.AutoTt.StartGeneration()

		// Need an extra goroutine so that this can return immediately.
		//go autotimetable.AutoTt.StartGeneration()
		//base.LogCommandEnd(false)
		//return false
	}
	//return true
}

func stoptt(op *DispatchOp) {
	if CheckArgs(op, 0) {
		base.SetStopFlag(true)
	}
}

// Get the result data as a JSON string.
func ttresult(op *DispatchOp) {
	if CheckArgs(op, 0) {
		result := autotimetable.AutoTt.GetLastResultJSON()
		//TODO: At present the JSON result is generated automatically as a
		// file. It might be preferable to return the data as a string result
		// instead.
		_ = result
	}
}

// Set a parameter for autotimetable.
func ttparameter(op *DispatchOp) {
	key := op.Data[0]
	val := op.Data[1]
	switch key {

	case "TIMEOUT":
		n, err := strconv.Atoi(val)
		if err != nil {
			base.LogError("--BAD_NUMBER %s=%s", key, val)
			return
		} else {
			autotimetable.TtParameters.TIMEOUT = n
		}

	case "MAXPROCESSES":
		n, err := strconv.Atoi(val)
		if err != nil {
			base.LogError("--BAD_NUMBER %s=%s", key, val)
			return
		} else {
			autotimetable.TtParameters.MAXPROCESSES = autotimetable.MaxProcesses(n)
			val = strconv.Itoa(autotimetable.TtParameters.MAXPROCESSES)
		}

	case "WRITE_FET_FILE":
		autotimetable.TtParameters.WRITE_FET_FILE = (val == "true")

	case "DEBUG":
		autotimetable.TtParameters.DEBUG = (val == "true")

	case "TESTING":
		autotimetable.TtParameters.TESTING = (val == "true")

	case "SKIP_HARD":
		autotimetable.TtParameters.SKIP_HARD = (val == "true")

	case "REAL_SOFT":
		autotimetable.TtParameters.REAL_SOFT = (val == "true")

	default:
		base.LogError("--UNKNOWN_PARAMETER %s", key)
		return
	}

	base.LogResult(key, val)
}

func nprocesses(op *DispatchOp) {
	nmin, np, nopt := autotimetable.MinNpOptProcesses()
	base.LogResult(op.Op, fmt.Sprintf("%d.%d.%d", nmin, np, nopt))
}

// Return the high priority processes handled in autotimetable's phase 0
func priortityConstraints(op *DispatchOp) {
	if CheckArgs(op, 0) {
		ctlist := []string{}
		for _, ct := range autotimetable.AutoTt.Source.GetPhase0ConstraintTypes() {
			ctlist = append(ctlist, strings.TrimPrefix(ct, "Constraint"))
		}
		base.LogResult("PRIORITY_CONSTRAINTS", strings.Join(ctlist, ":"))
	}
}

// Return the hard constraints sorted according to priority.
func hardConstraints(op *DispatchOp) {
	if CheckArgs(op, 0) {
		for _, c := range autotimetable.AutoTt.Constraint_Types {
			ilist, ok := autotimetable.AutoTt.HardConstraintMap[c]
			if ok {
				base.LogResult(
					strings.TrimPrefix(c, "Constraint"),
					strconv.Itoa(len(ilist)))
			}
		}
	}
}

// Return the soft constraints sort according to weight.
func softConstraints(op *DispatchOp) {
	if CheckArgs(op, 0) {
		clist := slices.SortedFunc(maps.Keys(autotimetable.AutoTt.SoftConstraintMap),
			func(a, b string) int { return strings.Compare(b, a) })
		for _, c := range clist {
			base.LogResult(
				strings.Replace(c, ":Constraint", ":", 1),
				strconv.Itoa(len(autotimetable.AutoTt.SoftConstraintMap[c])))
		}
	}
}

func nActivities(op *DispatchOp) {
	if CheckArgs(op, 0) {
		n := autotimetable.AutoTt.NActivities
		if n == 0 {
			base.LogError("--NO_ACTIVITIES")
		}
		base.LogResult("N_ACTIVITIES", strconv.Itoa(n))
	}
}
