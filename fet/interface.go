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

	//TODO: Do I need a "necessary" list here for those "basic" constraints
	// which must always be enabled? Or can they just be left out of the
	// other lists?

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
