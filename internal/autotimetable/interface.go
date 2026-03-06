package autotimetable

import "fetrunner/internal/base"

type TtBackend interface {
	RunBackend(*base.BaseData, *TtInstance) TtInstanceBackend
	Tidy(*base.BaseData)
	ConstraintName(*TtInstance) string

	//TODO:
	EnableConstraint(index int)
}

//TODO: In autotimetable primarily indexes should be used to refer to elements
// of all sorts, including constraints, all 0-based (including
// the activities!). If some other sort of reference is needed (distinguish
// node-ref and tag) it should be available in the source and/or back-end
// interfaces.

type TtSource interface {
	//TODO?
	Prepare(real_soft bool) // set the soft-constraint weights

	//TODO?
	//GetConstraintTypeSets() map[string][]int // ctype -> []constraint-index

	GetDays() []base.ElementBase
	GetHours() []base.ElementBase
	GetTeachers() []base.ElementBase

	//TODO???
	GetClasses() []base.ElementBase
	//GetStudentGroups() []base.ElementBase

	GetRooms() []base.ElementBase
	GetActivities() []base.ElementBase

	GetConstraints() []AttConstraint

	//TODO:
	//ConstraintRef(index int) string // get source reference for indexed constraint

	GetNActivities() int
	GetNConstraints() ConstraintIndex
	GetConstraint_Types() []ConstraintType // ordered list of constraint types
	GetHardConstraintMap() map[ConstraintType][]ConstraintIndex
	GetSoftConstraintMap() map[ConstraintType][]ConstraintIndex

	//TODO: probably not here ...
	// Prepare the "source" for a run with a set of enabled constraints:
	PrepareRun([]bool, any)
}
