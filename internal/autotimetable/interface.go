package autotimetable

import "fetrunner/internal/base"

type TtBackend interface {
	RunBackend(*AutoTtData, *TtInstance) TtInstanceBackend
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
	SourceType() string

	//TODO?
	//SetSoftConstraintWeights(real_soft bool)

	//TODO?
	//GetConstraintTypeSets() map[string][]int // ctype -> []constraint-index

	GetDays() []base.ElementBase
	GetHours() []base.ElementBase
	GetTeachers() []base.ElementBase

	//TODO???
	GetClasses() []*TtClass
	GetAtomicGroups() []string
	//GetStudentGroups() []base.ElementBase

	GetSubjects() []base.ElementBase
	GetRooms() []base.ElementBase
	GetActivities() []*TtActivity

	//TODO: temporary name? -> GetConstraints?
	GetSourceConstraints() []*TtConstraint
	//TODO--???
	//GetConstraints() []AttConstraint

	//TODO:
	//ConstraintRef(index int) string // get source reference for indexed constraint

	GetNActivities() int
	GetNConstraints() ConstraintIndex
	GetConstraint_Types() []ConstraintType // ordered list of constraint types
	// GetConstraintMaps() returns the hard- and soft constraint maps.
	GetConstraintMaps() (map[ConstraintType]ConstraintIndex, map[ConstraintType]ConstraintIndex)

	//TODO: probably not here ...
	// Prepare the "source" for a run with a set of enabled constraints:
	//PrepareRun([]bool, any)
}
