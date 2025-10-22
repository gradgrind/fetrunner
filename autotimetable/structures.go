package autotimetable

// Structures and global variables used in connection with automation of the
// timetable generation.

type ActivityIndex int
type RoomIndex int

var Ticks int // global time ticker
// The instance tick counter is in `TtInstance` because it may be needed
// by the back-end.
var constraint_data *ConstraintData // the original data

type ConstraintType string
type ConstraintIndex int
type ConstraintData struct {
	InputData         any // Managed by back-end
	NActivities       ActivityIndex
	NConstraints      ConstraintIndex
	ConstraintTypes   []ConstraintType // ordered list of constraint types
	HardConstraintMap map[ConstraintType][]ConstraintIndex
	SoftConstraintMap map[ConstraintType][]ConstraintIndex

	// Read input data:
	Read func(*ConstraintData, string) error
	// Return a string representation of the given constraint:
	ConstraintString func(*ConstraintData, ConstraintIndex) string
	PrepareRun       func(*ConstraintData, []bool, any)
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
	Ticks           int  // run time of this instance
	Stopped         bool // `abort_instance()` has been called on this instance
	ProcessingState int  // -1: queued, 0: running, 1: success, 2: failure,
	// there is also 3: cancelled
	// The following are set by the back-end:
	BackEndData any
	InstanceDir string // working space for this instance
	RunState    int    // set by back-end
	Progress    int    // percent
	Message     string // "" or error message
}

type TtBackend struct {
	New     func(*TtInstance)
	Run     func(*TtInstance, bool)
	Abort   func(*TtInstance)
	Tick    func(*TtInstance)
	Clear   func(*TtInstance)
	Tidy    func(string)
	Results func(*TtInstance) []ActivityPlacement
}

// This structure is used to return the placement results from the
// timetable back-end.
type ActivityPlacement struct {
	Id    ActivityIndex
	Day   int
	Hour  int
	Rooms []RoomIndex
}

var Backend *TtBackend

// `WorkingDir` provides the path to a working directory which can be used
// freely during processing. It may or may not already exist: existing
// contents need not be preserved during processing.
var WorkingDir string
