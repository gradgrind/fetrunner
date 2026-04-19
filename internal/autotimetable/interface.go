package autotimetable

import "fetrunner/internal/base"

// In autotimetable primarily indexes should be used to refer to elements
// of all sorts, including constraints, all 0-based (including
// the activities!). If some other sort of reference is needed (distinguish
// node-ref and tag) it should be available in the source and/or back-end
// interfaces.

type TtBackend interface {
	RunBackend(*AutoTtData, *TtInstance)
	Tidy()
	ConstraintName(*TtInstance) string
	Results(*TtInstance) []TtActivityPlacement
}

type TtInstanceBackend interface {
	Abort()
	DoTick(*AutoTtData, *TtInstance)
	Clear()
	FinalizeResult(*AutoTtData)
}

type TtSource interface {
	GetDays() []base.ElementBase
	GetHours() []base.ElementBase
	GetTeachers() []base.ElementBase
	GetClasses() []*TtClass
	GetAtomicGroups() []string
	GetSubjects() []base.ElementBase
	GetRooms() []base.ElementBase
	GetActivities() []*TtActivity
	GetConstraints() []*TtConstraint
	GetPhase0ConstraintTypes() []ConstraintType
	GetConstraintTypes() []ConstraintType // ordered list of constraint types
	// GetConstraintMaps() returns the hard- and soft constraint maps.
	GetConstraintMaps() (map[ConstraintType][]ConstraintIndex, map[ConstraintType][]ConstraintIndex)
}
