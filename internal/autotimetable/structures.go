package autotimetable

import (
	"fetrunner/internal/base"
)

// Structures and global variables used in connection with automation of the
// timetable generation.

type ConstraintType = string
type ConstraintIndex = int

const MIN_TIMEOUT int = 3

type Parameters struct {
	// The behaviour of the TESTING flag depends on the back-end. It
	// might, for example, use fixed seeds for random number generators
	// so as to produce reproduceable runs.
	TESTING bool
	// If the SKIP_HARD flag is true, assume the hard constraints are
	// satisfiable – skip the unconstrained instance, basing tests on
	// the hard-only instance.
	SKIP_HARD bool
	// If the REAL_SOFT flag is true, retain the weight of the soft
	// constraints, i.e. treat them as hard constraints. Otherwise (the
	// default) they will be made "hard" by setting the weight to 100.
	// For instance construction and running order, they will still be
	// treated as soft constraints.
	REAL_SOFT bool
	// `fetrunner` relies on parallel processing. If there are too few
	// real processors it will be inefficient. MAXPROCESSES sets the
	// maximum number of timetable generations (`FET` processes) which
	// may run at the same time.
	MAXPROCESSES int

	// A string to select an alternative timetable-generation back-end.
	// Currently only FET is available, and is the default.
	BACKEND string

	TIMEOUT                  int // the overall timeout, secs
	NEW_BASE_TIMEOUT_FACTOR  int // factor * 10
	NEW_PHASE_TIMEOUT_FACTOR int // factor * 10

	DEBUG bool

	// If true, write fully constrained FET-file with "_" prefix to source
	// directory:
	WRITE_FET_FILE bool

	// Tick count limits for testing whether an instance with no timeout
	// has got stuck. See `(*RunQueue).update_instances()` method.
	LAST_TIME_0 int
	LAST_TIME_1 int
}

// The `AutoTtData` structure is set up once for the handling of a set of
// timetable data (based on a source file, for example).
type AutoTtData struct {
	Parameters *Parameters
	BaseData   *base.BaseData

	Source  TtSource
	Backend TtBackend

	NActivities       int
	NConstraints      ConstraintIndex
	Constraint_Types  []ConstraintType // ordered list of constraint types
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
	na_instance     *TtInstance

	//TODO
	phase int // 0: initial phase, 1: adding hard constraints,
	// 2: adding soft constraints, 3: finished
	// The (successful) instance on which current trials are based:
	current_instance *TtInstance
	// List of instances adding a constraint type:
	constraint_instance_list []*TtInstance
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
	Weight         string // "internal" weight ("00"-"99", or "" if hard)
	Constraints    []ConstraintIndex

	// Run time ...
	Backend TtInstanceBackend // interface to generator back-end
	Ticks   int               // run time of this instance
	Stopped bool              // `abort_instance()` has been called on this instance

	// `RunState` is used in the tick-loop, but the "finished" states are set
	// using the back-end `DoTick` method (though still in the thread of the
	// tick-loop).
	RunState int // 0: not started, <0: running (not finished),
	// 1: finished (success, 100%), 2: finished (unsuccessful),
	// 3: obsolete / don't start (waiting for deletion)
	// The following are set by the back-end:
	Progress int    // percent
	LastTime int    // last (instance) time at which the back-end made progress
	Message  string // "" or error message
}

type TtInstanceBackend interface {
	Abort()
	DoTick(*base.BaseData, *AutoTtData, *TtInstance)
	Clear()
	Results(*base.BaseData, *AutoTtData, *TtInstance) []TtActivityPlacement
	FinalizeResult(*base.BaseData, *AutoTtData)
}

type ActivityIndex = int
type TeacherIndex = int
type RoomIndex = int
type ClassIndex = int
type AtomicIndex = int

type TtClass struct {
	Id            base.NodeRef
	Tag           string // the (short) name of the class
	AtomicIndexes []AtomicIndex
	Groups        []*TtGroup
}

type TtGroup struct {
	Id            base.NodeRef
	Tag           string // the (short) name of the group (without class)
	ClassIndex    int
	AtomicIndexes []AtomicIndex
}

// This structure is used to return the placement results from the
// timetable back-end. It differs from `base.ActivityPlacement` in that it
// uses indexes rather than NodeRefs.
type TtActivityPlacement struct {
	Activity ActivityIndex
	Day      int
	Hour     int
	Rooms    []RoomIndex
}

type AttConstraint struct {
	TtSourceTag string
	Ctype       string
	Parameters  []int
	Weight      int
}

/*TODO: How to handle source and back-end constraints?

There needs to be a converter which takes both source and back-end into
account, so it can't be associated with just one end. In general, I suppose
it is an m*n situation. Maybe it can be primarily source-based, allowing a
choice of back-end on the basis of the source. The other way round would
also be possible, in principle, but perhaps less "practical".

As far as the autotimetable function is concerned, only a limited set of
parameters/methods needs to be exposed via the back-end interface, though
the source constraints would be the visible ones. In AutoTtData:

...

    NActivities       int

    NConstraints      ConstraintIndex
    Constraint_Types  []ConstraintType // ordered list of constraint types
    HardConstraintMap map[ConstraintType][]ConstraintIndex
    SoftConstraintMap map[ConstraintType][]ConstraintIndex

...

*/
