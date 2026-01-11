package fet

import (
	"fetrunner/internal/autotimetable"
	"math"
	"slices"
	"strconv"

	"github.com/beevik/etree"
)

type IdPair = autotimetable.IdPair
type Constraint = autotimetable.Constraint
type ConstraintIndex = autotimetable.ConstraintIndex
type AutoTtData = autotimetable.AutoTtData
type ConstraintType = autotimetable.ConstraintType

type SoftWeight struct {
	Index  ConstraintIndex
	Weight string
}

type TtRunDataFet struct {
	Doc                *etree.Document
	ConstraintElements []*etree.Element
	SoftWeights        []SoftWeight

	// ActivityElements is currently not used
	ActivityElements []*etree.Element

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

	WeightTable []float64

	NConstraints      ConstraintIndex
	ConstraintTypes   []ConstraintType
	HardConstraintMap map[ConstraintType][]ConstraintIndex
	SoftConstraintMap map[ConstraintType][]ConstraintIndex
}

func (rundata *TtRunDataFet) Prepare(real_soft bool) {
	for _, cw := range rundata.SoftWeights {
		e := rundata.ConstraintElements[cw.Index]
		if real_soft {
			e.SelectElement("Weight_Percentage").SetText(cw.Weight)
		} else {
			e.SelectElement("Weight_Percentage").SetText("100")
		}
	}
}

func (rundata *TtRunDataFet) GetDays() []IdPair            { return rundata.DayIds }
func (rundata *TtRunDataFet) GetHours() []IdPair           { return rundata.HourIds }
func (rundata *TtRunDataFet) GetTeachers() []IdPair        { return rundata.TeacherIds }
func (rundata *TtRunDataFet) GetSubjects() []IdPair        { return rundata.SubjectIds }
func (rundata *TtRunDataFet) GetRooms() []IdPair           { return rundata.RoomIds }
func (rundata *TtRunDataFet) GetClasses() []IdPair         { return rundata.ClassIds }
func (rundata *TtRunDataFet) GetActivities() []IdPair      { return rundata.ActivityIds }
func (rundata *TtRunDataFet) GetConstraints() []Constraint { return rundata.Constraints }

func (rundata *TtRunDataFet) GetNActivities() int                  { return len(rundata.ActivityIds) }
func (rundata *TtRunDataFet) GetNConstraints() ConstraintIndex     { return rundata.NConstraints }
func (rundata *TtRunDataFet) GetConstraintTypes() []ConstraintType { return rundata.ConstraintTypes }
func (rundata *TtRunDataFet) GetHardConstraintMap() map[ConstraintType][]ConstraintIndex {
	return rundata.HardConstraintMap
}
func (rundata *TtRunDataFet) GetSoftConstraintMap() map[ConstraintType][]ConstraintIndex {
	return rundata.SoftConstraintMap
}

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
	rundata.Doc.Indent(2)
	var err error
	*(xmlp.(*[]byte)), err = rundata.Doc.WriteToBytes()
	if err != nil {
		panic(err)
	}
}

func MakeFetWeights() []float64 {
	wtable := make([]float64, 101)
	wtable[0] = 0.0
	wtable[100] = 100.0
	for w := 1; w < 100; w++ {
		wf := float64(w + 1)
		denom := wf + math.Pow(2, (wf-50.0)*0.2)
		wtable[w] = 100.0 - 100.0/denom
	}
	return wtable
}

func (rundata *TtRunDataFet) FetWeight(w int) string {
	if w <= 0 {
		return "0"
	}
	if w >= 100 {
		return "100"
	}
	return strconv.FormatFloat(rundata.WeightTable[w], 'f', 3, 64)
}

func (rundata *TtRunDataFet) DbWeight(w string) int {
	wf, err := strconv.ParseFloat(w, 64)
	if err != nil {
		panic(err)
	}
	wdb, _ := slices.BinarySearch(rundata.WeightTable, wf)
	return wdb
}
