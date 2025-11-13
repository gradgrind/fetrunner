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
	TimeConstraints  []int // indexes into `Constraints`
	SpaceConstraints []int // indexes into `Constraints`

	//TODO: Do I need a "necessary" list here for those "basic" constraints
	// which must always be enabled? Or can they just be left out of the
	// other lists?

	DayIds     []IdPair
	HourIds    []IdPair
	TeacherIds []IdPair
}

func (rundata *TtRunDataFet) GetDays() []IdPair     { return rundata.DayIds }
func (rundata *TtRunDataFet) GetHours() []IdPair    { return rundata.HourIds }
func (rundata *TtRunDataFet) GetTeachers() []IdPair { return rundata.TeacherIds }
