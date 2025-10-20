package autotimetable

// Structures and global variables used in connection with automation of the
// timetable generation.

var Ticks int // global time ticker
// The instance tick counter is in `TtInstance` because it may be needed
// by the back-end.
var constraint_data *ConstraintData // the original data

type ConstraintType string
type ConstraintIndex int
type ConstraintData struct {
	InputData         any
	Constraints       ConstraintIndex
	ConstraintTypes   []ConstraintType // ordered list of constraint types
	HardConstraintMap map[ConstraintType][]ConstraintIndex
	SoftConstraintMap map[ConstraintType][]ConstraintIndex
	ConstraintString  func(*ConstraintData, ConstraintIndex) string
}

type TtInstance struct {
	Tag         string
	InstanceDir string // working space for this instance
	Timeout     int    // ticks

	// Base instance from which this instance is derived:
	BaseInstance *TtInstance

	// Constraints enabled in this instance:
	ConstraintEnabled []bool // ConstraintIndex -> enabled
	// Constraints to be added in this instance:
	ConstraintType ConstraintType
	Constraints    []ConstraintIndex

	// Run time
	BackEndData     any
	Ticks           int  // run time of this instance
	Stopped         bool // `abort_instance()` has been called on this instance
	ProcessingState int  // -1: queued, 0: running, 1: success, 2: failure,
	// there is also 3: cancelled
}

type ActivityIndex int
type RoomIndex int

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
