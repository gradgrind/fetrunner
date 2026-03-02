package autotimetable

import "fetrunner/internal/base"

type TtBackend interface {
	RunBackend(*base.BaseData, *TtInstance) TtInstanceBackend
	Tidy(*base.BaseData)
	ConstraintName(*TtInstance) string
}

//TODO: Make autotimetable use only indexes (instead of TtSourceItem structs)
// for the elements, all 0-based (including the activities!). If some other
// sort of reference is needed (distinguish node-ref and tag) it should be
// available in the source and/or back-end interfaces.

type TtSource interface {
	//TODO?
	Prepare(real_soft bool) // set the soft-constraint weights

	//TODO?
	//GetConstraintTypeSets() map[string][]int // ctype -> []constraint-index

	GetDays() []TtSourceItem
	GetHours() []TtSourceItem
	GetTeachers() []TtSourceItem

	//TODO???
	GetClasses() []TtSourceItem
	//GetStudentGroups() []TtSourceItem

	GetRooms() []TtSourceItem
	GetActivities() []TtSourceItem

	GetConstraints() []Constraint

	GetNActivities() int
	GetNConstraints() ConstraintIndex
	GetConstraint_Types() []ConstraintType // ordered list of constraint types
	GetHardConstraintMap() map[ConstraintType][]ConstraintIndex
	GetSoftConstraintMap() map[ConstraintType][]ConstraintIndex

	//TODO: probably not here ...
	// Prepare the "source" for a run with a set of enabled constraints:
	PrepareRun([]bool, any)
}

type TtSourceItem struct {
	Index int    // source reference as index (0-based)
	Tag   string // short text identifier
}

// TODO?
//type IdPair struct {
//	Source  string // source reference
//	Backend string // generator back-end id
//}

type Constraint struct {
	TtSourceItem
	Ctype      string
	Parameters []int
	Weight     int
}
