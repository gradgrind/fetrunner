package fet

// Constraint "priority", used for making an ordered list of the constraint
// types in use.
// Note that the absolute values are not important, it is the relative
// values which determine the order. Constraint types not listed here
// have priority 0.
var ConstraintPriority = map[constraintType]int{
	"ConstraintRoomNotAvailableTimes":         100,
	"ConstraintStudentsSetNotAvailableTimes":  99,
	"ConstraintTeacherNotAvailableTimes":      98,
	"ConstraintActivityPreferredStartingTime": 90,

	"ConstraintStudentsMinHoursDaily":         5,
	"ConstraintStudentsSetMinHoursDaily":      4,
	"ConstraintStudentsMinHoursPerMorning":    -4,
	"ConstraintStudentsSetMinHoursPerMorning": -5,

	"ConstraintTeachersMaxGapsPerDay":     -93,
	"ConstraintTeachersMaxGapsPerWeek":    -94,
	"ConstraintTeacherMaxGapsPerDay":      -95,
	"ConstraintTeacherMaxGapsPerWeek":     -96,
	"ConstraintStudentsMaxGapsPerDay":     -97,
	"ConstraintStudentsMaxGapsPerWeek":    -98,
	"ConstraintStudentsSetMaxGapsPerDay":  -99,
	"ConstraintStudentsSetMaxGapsPerWeek": -100,
}
