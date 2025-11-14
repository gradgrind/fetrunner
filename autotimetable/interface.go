package autotimetable

type TtSource interface {
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

	//TODO
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
