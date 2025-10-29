package autotimetable

// Structures and global variables used in connection with automation of the
// timetable generation.

type ConstraintType string
type ConstraintIndex int

// The `BasicData` structure is set up once for the handling of a set of
// timetable data (based on a source file, for example).
type BasicData struct {
	Parameters struct {
		// The behaviour of the TESTING flag depends on the back-end. It
		// might, for example, use fixed seeds for random number generators
		// so as to produce reproduceable runs.
		TESTING bool
		// This approach relies on parallel processing. If there are too few
		// real processors it will be inefficient:
		MAXPROCESSES int

		NEW_BASE_TIMEOUT_FACTOR  int // factor * 10
		CYCLE_TIMEOUT_MIN        int
		NEW_CYCLE_TIMEOUT_FACTOR int // factor * 10

		DEBUG bool

		// Tick count limits for testing whether an instance with no timeout
		// has got stuck. See `(*RunQueue).update_instances()` method.
		LAST_TIME_0 int
		LAST_TIME_1 int
	}

	Source     TtSource
	RunBackend func(*BasicData, *TtInstance) TtBackend

	NActivities       int
	NConstraints      ConstraintIndex
	ConstraintTypes   []ConstraintType // ordered list of constraint types
	HardConstraintMap map[ConstraintType][]ConstraintIndex
	SoftConstraintMap map[ConstraintType][]ConstraintIndex
	ConstraintErrors  map[ConstraintIndex]string // collect error messages

	// `WorkingDir` provides the path to a working directory which can be used
	// freely during processing. It is set up before entering `StartGeneration`.
	WorkingDir string
	Ticks      int // "global" time ticker
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
	cycle int // processing cycle number (in `mainphase`)
	// The (successful) instance on which current trials are based:
	current_instance *TtInstance
	// List of instances adding a constraint type:
	constraint_list []*TtInstance
}

type TtSource interface {
	GetActivityIds() []ActivityId
	GetRooms() []TtItem
	GetDayTags() []string
	GetHourTags() []string
	// Return a string representation of the given constraint:
	GetConstraintItems() []TtItem
	// Prepare the "source" for a run with a set of enabled constraints:
	PrepareRun([]bool, any)
}

type ActivityId struct {
	Id  int    // (generator) back-end activity index
	Ref string // (input) source reference/identifier for activity, if
	// distinct from `Id`
}

type TtInstance struct {
	Tag     string
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
	TimedOut bool // marked as timed out, will (probably) lead to termination

	// The following are set by the back-end:
	RunState int
	Progress int    // percent
	LastTime int    // last (instance) time at which `Progress` changed
	Message  string // "" or error message
}

type TtBackend interface {
	Abort()
	Tick(*BasicData, *TtInstance)
	Clear()
	Tidy(string)
	Results(*BasicData, *TtInstance) []ActivityPlacement
	FinalizeResult(*BasicData)
}

// This structure is used to return the placement results from the
// timetable back-end.
type ActivityPlacement struct {
	// All indexes are zero-based.
	Activity int   // activity index
	Day      int   // day index
	Hour     int   // hour index
	Rooms    []int // room indexes
}

//************** Resources **************

// TODO ...

/*
type Resource interface {
	GetIndex() int  //TODO: within the type?
	GetTag() string // short identifier (unique)
	//Name // TODO: possibly longer identifier (probably, but not
	// necessarily unique?)
	//Item any // TODO: back-end dependent data
}

// +++ Student group resources
type StudentResource interface {
	Resource
	GetClass() *TtClass
	IsAtomicGroup() bool
}

type TtResource struct {
	Index int
	Tag   string
	Data  any
}

func (r *TtResource) GetIndex() int {
	return r.Index
}

func (r *TtResource) GetTag() string {
	return r.Tag
}

type TtClass struct {
	TtResource
	Groups []*TtGroup
}

type TtGroup struct {
	TtResource
	Class     *TtClass
	Subgroups []*TtSubgroup
}

type TtSubgroup struct {
	TtResource
	Class *TtClass
	//Groups []*TtGroup
}

func (r *TtClass) GetClass() *TtClass {
	return r
}

func (r *TtGroup) GetClass() *TtClass {
	return r.Class
}

func (r *TtSubgroup) GetClass() *TtClass {
	return r.Class
}

func (r *TtClass) IsAtomic() bool {
	return len(r.Groups) == 0
}

func (r *TtGroup) IsAtomic() bool {
	return len(r.Subgroups) == 0
}

func (r *TtSubgroup) IsAtomic() bool {
	return true
}

// --- Student group resources

// +++ Teacher resource
type TtTeacher struct {
	TtResource
}

// --- Teacher resource

// +++ Room resource
type TtRoom struct {
	TtResource
}

// --- Room resource
*/
