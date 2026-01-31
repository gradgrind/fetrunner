package autotimetable

import "fetrunner/internal/base"

type BackendInterface interface {
	RunBackend(*base.BaseData, *TtInstance) TtBackend
	Tidy(*base.BaseData)
	ConstraintName(*TtInstance) string
}

type TtSource interface {
	Prepare(real_soft bool) // set the soft-constraint weights

	//TODO?
	//GetConstraintTypeSets() map[string][]int // ctype -> []constraint-index

	GetDays() []IdPair
	GetHours() []IdPair
	GetTeachers() []IdPair

	//TODO???
	GetClasses() []IdPair
	//GetStudentGroups() []IdPair

	GetRooms() []IdPair
	GetActivities() []IdPair

	GetConstraints() []Constraint

	GetNActivities() int
	GetNConstraints() ConstraintIndex
	GetConstraint_Types() []ConstraintType // ordered list of constraint types
	GetHardConstraintMap() map[ConstraintType][]ConstraintIndex
	GetSoftConstraintMap() map[ConstraintType][]ConstraintIndex

	// Prepare the "source" for a run with a set of enabled constraints:
	PrepareRun([]bool, any)
}

type IdPair struct {
	Source  string // source reference
	Backend string // generator back-end id
}

type Constraint struct {
	IdPair
	Ctype      string
	Parameters []int
	Weight     int
}
