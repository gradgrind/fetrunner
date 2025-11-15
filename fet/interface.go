package fet

import (
	"fetrunner/autotimetable"

	"github.com/beevik/etree"
)

type IdPair = autotimetable.IdPair
type Constraint = autotimetable.Constraint

type TtRunDataFet struct {
	Doc                *etree.Document
	ConstraintElements []*etree.Element

	// FET has time and space constraints separate. It might be useful in
	// some way to have that information here.
	TimeConstraints  []int // indexes into `ConstraintElements`
	SpaceConstraints []int // indexes into `ConstraintElements`

	Constraints []Constraint
	ActivityIds []IdPair

	DayIds     []IdPair
	HourIds    []IdPair
	TeacherIds []IdPair
	RoomIds    []IdPair
	SubjectIds []IdPair
	ClassIds   []IdPair
}

func (rundata *TtRunDataFet) GetDays() []IdPair            { return rundata.DayIds }
func (rundata *TtRunDataFet) GetHours() []IdPair           { return rundata.HourIds }
func (rundata *TtRunDataFet) GetTeachers() []IdPair        { return rundata.TeacherIds }
func (rundata *TtRunDataFet) GetSubjects() []IdPair        { return rundata.SubjectIds }
func (rundata *TtRunDataFet) GetRooms() []IdPair           { return rundata.RoomIds }
func (rundata *TtRunDataFet) GetClasses() []IdPair         { return rundata.ClassIds }
func (rundata *TtRunDataFet) GetActivities() []IdPair      { return rundata.ActivityIds }
func (rundata *TtRunDataFet) GetConstraints() []Constraint { return rundata.Constraints }

// Rebuild the FET file given an array detailing which constraints are enabled.
// The `xmlp` argument is a pointer to a byte slice, to receive the
// XML FET-file.
func (rundata *TtRunDataFet) PrepareRun(enabled []bool, xmlp any) {
	for i, c := range rundata.ConstraintElements {
		active := c.SelectElement("Active")
		if enabled[i] {
			active.SetText("true")
		} else {
			active.SetText("false")
		}
	}
	root := rundata.Doc.Root()
	et := root.SelectElement("Time_Constraints_List")
	active := 0
	n := 0
	for _, e := range et.ChildElements() {
		// Count and skip if inactive
		if e.SelectElement("Active").Text() == "true" {
			active++ // count active constraints
		}
		n++
	}
	var err error
	*(xmlp.(*[]byte)), err = rundata.Doc.WriteToBytes()
	if err != nil {
		panic(err)
	}
}
