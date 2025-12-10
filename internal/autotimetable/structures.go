package autotimetable

import (
	"fetrunner/internal/base"
	"fetrunner/internal/timetable"
)

// Structures and global variables used in connection with automation of the
// timetable generation.

type ConstraintType = string
type ConstraintIndex = int

type Parameters struct {
	//TODO: not all entries are (currently) used by libfetrunner

	// The behaviour of the TESTING flag depends on the back-end. It
	// might, for example, use fixed seeds for random number generators
	// so as to produce reproduceable runs.
	TESTING bool
	// If the SKIP_HARD flag is true, assume the hard constraints are
	// satisfiable – skip the unconstrained instance, basing tests on
	// the hard-only instance.
	SKIP_HARD bool
	// This approach relies on parallel processing. If there are too few
	// real processors it will be inefficient:
	MAXPROCESSES int
	FETPATH      string // path to `fet-cl` executable

	TIMEOUT                  int // the overall timeout, secs
	NEW_BASE_TIMEOUT_FACTOR  int // factor * 10
	CYCLE_TIMEOUT_MIN        int
	NEW_CYCLE_TIMEOUT_FACTOR int // factor * 10

	DEBUG bool

	// Tick count limits for testing whether an instance with no timeout
	// has got stuck. See `(*RunQueue).update_instances()` method.
	LAST_TIME_0 int
	LAST_TIME_1 int
}

// The `AutoTtData` structure is set up once for the handling of a set of
// timetable data (based on a source file, for example).
type AutoTtData struct {
	Parameters *Parameters

	Source           TtSource
	BackendInterface BackendInterface

	NActivities       int
	NConstraints      ConstraintIndex
	ConstraintTypes   []ConstraintType // ordered list of constraint types
	HardConstraintMap map[ConstraintType][]ConstraintIndex
	SoftConstraintMap map[ConstraintType][]ConstraintIndex
	ConstraintErrors  map[ConstraintIndex]string // collect error messages
	BlockConstraint   map[ConstraintIndex]bool   // if true, don't enable the constraint

	Ticks int // "global" time ticker
	// The instance tick counter is in `TtInstance` because it may be needed
	// by the run-time back-end.

	// Local variables
	instanceCounter int
	lastResult      *Result
	cycle_timeout   int
	full_instance   *TtInstance
	null_instance   *TtInstance
	hard_instance   *TtInstance
	phase           int // 0: initial phase, 1: adding hard constraints,
	// 2: adding soft constraints, 3: finished
	//TODO-- cycle int // processing cycle number (in `mainphase`)
	// The (successful) instance on which current trials are based:
	current_instance *TtInstance
	// List of instances adding a constraint type:
	constraint_list []*TtInstance
}

type TtInstance struct {
	Index   int
	Timeout int // ticks

	// Base instance from which this instance is derived:
	BaseInstance *TtInstance

	// Constraints enabled in this instance:
	ConstraintEnabled []bool // ConstraintIndex -> enabled
	// Constraints to be added in this instance:
	ConstraintType ConstraintType
	Constraints    []ConstraintIndex

	// Run time ...
	Backend         TtBackend // interface to generator back-end
	Ticks           int       // run time of this instance
	Stopped         bool      // `abort_instance()` has been called on this instance
	ProcessingState int       // -1: queued, 0: running, 1: success, 2: failure,
	// there is also 3: cancelled

	// The following are set by the back-end:
	RunState int
	Progress int    // percent
	LastTime int    // last (instance) time at which the back-end made progress
	Message  string // "" or error message
}

type TtBackend interface {
	Abort()
	Tick(*base.BaseData, *AutoTtData, *TtInstance)
	Clear()
	Results(*base.BaseData, *AutoTtData, *TtInstance) []TtActivityPlacement
	FinalizeResult(*base.BaseData, *AutoTtData)
}

type TtActivityPlacement = timetable.TtActivityPlacement
