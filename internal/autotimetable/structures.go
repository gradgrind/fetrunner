package autotimetable

import (
	"fetrunner/internal/base"
)

// Structures and global variables used in connection with automation of the
// timetable generation.

type ConstraintType = string
type ConstraintIndex = int

const MIN_TIMEOUT int = 3 // minimum "timeout" value for all non-special instances

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
	// If the WITHOUT_ROOM_CONSTRAINTS flag is true, no rooms will be allocated.
	// TODO: Is this flag useful? It is implemented, but currently not set anywhere.
	WITHOUT_ROOM_CONSTRAINTS bool

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

	Ticks int // "global" time ticker
	// The instance tick counter is in `TtInstance` because it may be needed
	// by the run-time back-end.

	// Local variables
	instanceCounter   int
	lastResult        *Result
	cycle_timeout     int
	full_instance     *TtInstance
	null_instance     *TtInstance
	hard_instance     *TtInstance
	priority_instance *TtInstance

	phase int // 0: initial phase, 1: adding hard constraints,
	// 2: adding soft constraints, 3: finished
	// The (successful) instance on which current trials are based:
	current_instance *TtInstance

	run_queue           []*TtInstance // pending constraint-type instances
	run_queue_next      int           // index of next instance in queue
	active_instances    []*TtInstance // list of running constraint-type instances
	timed_out_instances []*TtInstance // single-constraint instances stopped for being too slow
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
	InstanceBackend TtInstanceBackend // interface to generator back-end for this instance
	Ticks           int               // run time of this instance

	// `RunState` is used in the tick-loop, but the "Done" flag is set
	// using the back-end `DoTick` method (though still in the thread of the
	// tick-loop).
	RunState  int  // for possible values see constants below
	Processed bool // set to `true` when certain run-states have been handled
	// The following are set by the back-end:
	Finished bool   // set to `true` by `DoTick` when run completes
	Progress int    // percent
	LastTime int    // last (instance) time at which the back-end made progress
	Message  string // "" or error message
}

// The values of an instance's `RunState`
const (
	// unstarted = 0
	INSTANCE_RUNNING    = 1  // the instance has been started and is running
	ABORT_NEW_CYCLE     = 2  // the instance is no longer required
	ABORT_TIMED_OUT     = 3  // the instance was progressing too slowly
	INSTANCE_SUCCESSFUL = -1 // the instance completed successfully
	INSTANCE_FAILED     = -2 // the instance failed due to a data error
	INSTANCE_ABANDONED  = -3 // the instance is no longer relevant
	INSTANCE_ACCEPTED   = -4 // the (successful) instance has been used as a base
)

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

type TtActivity struct {
	Id string // FET: "Activity:" + ActivityId
	// DB: NodeRef of the source activity from which this is derived.
	Tag string // optionally usable by the back-end

	Duration           int
	Subject            string
	Groups             []base.ElementBase // a `Class` is represented by its ClassGroup
	AtomicGroupIndexes []AtomicIndex
	Teachers           []TeacherIndex
}

type TtConstraint struct {
	Id string // FET: "[index]", or "[index:fet-weight]" if weight < 100
	// DB: NodeRef of the source constraint from which this is derived.
	// A single NodeRef may be referenced by multiple TtConstraints.

	CType  string // name of constraint type.
	Weight int    // 0 – 100
	Data   any    // content dependent on constraint type
}

func (c *TtConstraint) IsHard() bool {
	return c.Weight == base.MAXWEIGHT
}
