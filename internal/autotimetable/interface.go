package autotimetable

import "fetrunner/internal/base"

// In autotimetable primarily indexes should be used to refer to elements
// of all sorts, including constraints, all 0-based (including
// the activities!). If some other sort of reference is needed (distinguish
// node-ref and tag) it should be available in the source and/or back-end
// interfaces.

type TtBackend interface {
    RunBackend(*AutoTtData, *TtInstance) TtInstanceBackend
    Tidy(*base.BaseData)
    ConstraintName(*TtInstance) string
    Results(*base.Logger, *TtInstance) []TtActivityPlacement
}

type TtInstanceBackend interface {
    Abort()
    DoTick(*base.BaseData, *AutoTtData, *TtInstance)
    Clear()
    FinalizeResult(*base.BaseData, *AutoTtData)
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
    GetResourceUnavailableConstraintTypes() []ConstraintType
    GetConstraintTypes() []ConstraintType // ordered list of constraint types
    // GetConstraintMaps() returns the hard- and soft constraint maps.
    GetConstraintMaps() (map[ConstraintType][]ConstraintIndex, map[ConstraintType][]ConstraintIndex)
}
