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
	Op  string
	Arg string
	OK  bool // can be set by called function
}

var OpHandlerMap map[string]func(*DispatchOp) = map[string]func(*DispatchOp){}

// A command is supplied as a string. If it constains a space character, the
// command name is the part before this space. The rest of the string is taken
// as argument to the command. The command name, which is looked up in `OpHandlerMap`,
// which provides the actual function to call. Normally, the beginning and end of a
// command will be logged and a command can only be started when no other is running.
// However, there are also special commands starting with "_", whose beginning and
// end are not logged and which may be run while another is still running. This
// is necessary for the command "_STOP_TT", which interrupts a long running process.
// It may be desirable for long running commands to return quickly – in particular
// for starting from a GUI, so as not to block the GUI. These commands should begin
// with a "!", which will cause the process to be run in a separate goroutine.
func Dispatch(cmd0 string) bool {
	cmd, arg, _ := strings.Cut(cmd0, " ")
	op := DispatchOp{Op: cmd, Arg: arg}
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
			base.LogCommand(cmd0)
			if op.Op[0] == '!' {
				go func() {
					f(&op)
					base.LogCommandEnd()
				}()
			} else {
				f(&op)
				base.LogCommandEnd()
			}
		}
	} else {
		panic("!InvalidOp: " + op.Op)
	}
	return op.OK
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
	OpHandlerMap["!RUN_TT"] = runtt
	OpHandlerMap["_STOP_TT"] = stoptt
	OpHandlerMap["RESULT_TT"] = ttresult

	OpHandlerMap["TT_PARAMETER"] = ttparameter
	OpHandlerMap["TMP_PATH"] = set_tmp
	OpHandlerMap["N_PROCESSES"] = nprocesses
	OpHandlerMap["TT_ConstraintsCheck"] = constraintsCheck
}

func fetrunner_version(op *DispatchOp) {
	base.LogResult("FETRUNNER_VERSION", VERSION)
}

func set_tmp(op *DispatchOp) {
	base.TEMPORARY_BASEDIR = op.Arg
	op.OK = base.SetTmpDir()
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
	fetpath := op.Arg
	if filepath.Dir(fetpath) == "." {
		// First try the directory containing the running executable.
		p, err := os.Executable()
		if err == nil {
			// Sanitize path
			p, err = filepath.EvalSymlinks(p)
			if err == nil {
				fetpath0 := filepath.Join(filepath.Dir(p), fetpath)
				if check_fet(fetpath0) {
					fet.FETPATH = fetpath0
					op.OK = true
					return
				}
			}
		}
		// Otherwise, try the bare command in case it is in the PATH.
	} else if !filepath.IsAbs(fetpath) {
		base.LogError("--FETPATH_NOT_ABSOLUTE %s", fetpath)
	}
	if check_fet(fetpath) {
		fet.FETPATH = fetpath
		op.OK = true
	}
}

// Handle (currently) ".fet" and "_w365.json" input files.
func file_loader(op *DispatchOp) {
	bd := base.DataBase
	fpath := op.Arg
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
	bdata := base.DataBase
	if bdata.Source == nil {
		base.LogError("--NO_SOURCE")
		//op.OK = false
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
	op.OK = true
}

func runtt(op *DispatchOp) {
	switch autotimetable.TtParameters.BACKEND {
	case "", "FET":
		fet.InitBackend(autotimetable.AutoTt)
	default:
		panic("Unsupported timetable-generation back-end: " + autotimetable.TtParameters.BACKEND)
	}
	autotimetable.AutoTt.StartGeneration()
}

func stoptt(op *DispatchOp) {
	base.SetStopFlag(true)
}

// Get the result data as a JSON string.
func ttresult(op *DispatchOp) {
	result := autotimetable.AutoTt.GetLastResultJSON()
	//TODO: At present the JSON result is generated automatically as a
	// file. It might be preferable to return the data as a string result
	// instead.
	_ = result
}

// Set a parameter for autotimetable.
func ttparameter(op *DispatchOp) {
	key, val, ok := strings.Cut(op.Arg, "=")
	if !ok {
		panic("Invalid parameter setting: " + op.Arg)
	}
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
			base.LogResult("MAXPROCESSES", autotimetable.TtParameters.MAXPROCESSES)
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
}

func nprocesses(op *DispatchOp) {
	nmin, np, nopt := autotimetable.MinNpOptProcesses()
	base.LogResult(op.Op, fmt.Sprintf("%d.%d.%d", nmin, np, nopt))
}

// Return the high priority processes handled in autotimetable's phase 0
func priortityConstraints(op *DispatchOp) {
	ctlist := []string{}
	for _, ct := range autotimetable.AutoTt.Source.GetPhase0ConstraintTypes() {
		ctlist = append(ctlist, strings.TrimPrefix(ct, "Constraint"))
	}
	base.LogResult(op.Op, strings.Join(ctlist, ":"))
}

// Return the hard constraints sorted according to priority.
func hardConstraints(op *DispatchOp) {
	for _, c := range autotimetable.AutoTt.Constraint_Types {
		ilist, ok := autotimetable.AutoTt.HardConstraintMap[c]
		if ok {
			base.LogResult(
				op.Op,
				fmt.Sprintf("%s*%d", strings.TrimPrefix(c, "Constraint"), len(ilist)))
		}
	}
}

// Return the soft constraints sort according to weight.
func softConstraints(op *DispatchOp) {
	clist := slices.SortedFunc(maps.Keys(autotimetable.AutoTt.SoftConstraintMap),
		func(a, b string) int { return strings.Compare(b, a) })
	for _, c := range clist {
		base.LogResult(
			op.Op,
			fmt.Sprintf("%s*%d", strings.Replace(c, ":Constraint", ":", 1),
				len(autotimetable.AutoTt.SoftConstraintMap[c])))
	}
}

func constraintsCheck(op *DispatchOp) {
	base.LogResult(
		op.Op,
		fmt.Sprintf("%d:%d",
			len(autotimetable.AutoTt.HardConstraintMap),
			len(autotimetable.AutoTt.SoftConstraintMap)))
}

func nActivities(op *DispatchOp) {
	n := autotimetable.AutoTt.NActivities
	if n == 0 {
		base.LogError("--NO_ACTIVITIES")
	}
	base.LogResult(op.Op, strconv.Itoa(n))
}
