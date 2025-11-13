package autotimetable

type TtRunData interface {
	GetConstraintTypeSets() map[string][]int // ctype -> []constraint-index

	GetDays() []IdPair
	GetHours() []IdPair
	GetTeachers() []IdPair

	//TODO???
	GetClasses() []IdPair
	GetStudentGroups() []IdPair

	GetRooms() []IdPair
	GetActivities() []IdPair

	GetConstraints() []*Constraint
}

type IdPair struct {
	Source  string
	Backend string
}

type Constraint struct {
	Ctype      string
	SourceRef  string
	Parameters []int
}
