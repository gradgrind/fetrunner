package fet

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"math"
	"slices"
	"strconv"

	"github.com/beevik/etree"
)

// In FET the IDs and "tags" (short names) are generally the same, and only
// unique within the repective category (teacher, room, etc.).

// type TtSourceItem = autotimetable.TtSourceItem
type element = base.ElementBase

type Constraint = autotimetable.AttConstraint
type ConstraintIndex = autotimetable.ConstraintIndex
type AutoTtData = autotimetable.AutoTtData
type ConstraintType = autotimetable.ConstraintType

type SoftWeight struct {
	Index  ConstraintIndex
	Weight string
}

type TtSourceFet struct {
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
	//--ActivityList []TtSourceItem

	WeightTable []float64

	NConstraints      ConstraintIndex
	ConstraintTypes   []ConstraintType
	HardConstraintMap map[ConstraintType][]ConstraintIndex
	SoftConstraintMap map[ConstraintType][]ConstraintIndex
}

func (sourcefet *TtSourceFet) Prepare(real_soft bool) {
	for _, cw := range sourcefet.SoftWeights {
		e := sourcefet.ConstraintElements[cw.Index]
		if real_soft {
			e.SelectElement("Weight_Percentage").SetText(cw.Weight)
		} else {
			e.SelectElement("Weight_Percentage").SetText("100")
		}
	}
}

func (sourcefet *TtSourceFet) GetDays() []element {
	items := []element{}
	for _, e := range sourcefet.Doc.Root().SelectElement("Days_List").SelectElements("Day") {
		id := e.SelectElement("Name").Text()
		items = append(items, element{
			// In FET the ID is also the tag
			Id: base.NodeRef(id), Tag: id})
	}
	return items
}

func (sourcefet *TtSourceFet) GetHours() []element {
	items := []element{}
	for _, e := range sourcefet.Doc.Root().SelectElement("Hours_List").SelectElements("Hour") {
		id := e.SelectElement("Name").Text()
		items = append(items, element{
			// In FET the ID is also the tag
			Id: base.NodeRef(id), Tag: id})
	}
	return items
}

func (sourcefet *TtSourceFet) GetTeachers() []element {
	items := []element{}
	for _, e := range sourcefet.Doc.Root().SelectElement("Teachers_List").SelectElements("Teacher") {
		id := e.SelectElement("Name").Text()
		items = append(items, element{
			// In FET the ID is also the tag
			Id: base.NodeRef(id), Tag: id})
	}
	return items
}

func (sourcefet *TtSourceFet) GetSubjects() []element {
	items := []element{}
	for _, e := range sourcefet.Doc.Root().SelectElement("Subjects_List").SelectElements("Subject") {
		id := e.SelectElement("Name").Text()
		items = append(items, element{
			// In FET the ID is also the tag
			Id: base.NodeRef(id), Tag: id})
	}
	return items
}

func (sourcefet *TtSourceFet) GetRooms() []element {
	items := []element{}
	for _, e := range sourcefet.Doc.Root().SelectElement("Rooms_List").SelectElements("Room") {
		// Only include real rooms, skip virtual ones.
		if e.SelectElement("Virtual").Text() == "false" {
			id := e.SelectElement("Name").Text()
			items = append(items, element{
				// In FET the ID is also the tag
				Id: base.NodeRef(id), Tag: id})
		}
	}
	return items
}

func (sourcefet *TtSourceFet) GetClasses() []element {
	items := []element{}
	for _, e := range sourcefet.Doc.Root().SelectElement("Students_List").SelectElements("Year") {
		id := e.SelectElement("Name").Text()
		items = append(items, element{
			// In FET the ID is also the tag
			Id: base.NodeRef(id), Tag: id})
	}
	return items
}

func (sourcefet *TtSourceFet) GetActivities() []element {
	aidlist := make([]element, len(sourcefet.ActivityElements))
	for i, a := range sourcefet.ActivityElements {
		aidlist[i] = element{
			//No Id
			Tag: a.SelectElement("Id").Text()}
	}
	return aidlist
}

func (sourcefet *TtSourceFet) GetConstraints() []Constraint { return sourcefet.Constraints }

/*TODO--?
func (sourcefet *TtSourceFet) ConstraintRef(index int) string {
    // A `FET` source file doesn't have any particular labelling of constraints.
    // A representation based on the constraint itself might be constructed, but as –
    // currently – only the `FET` back-end is under consideration for a `FET` source,
    // and this has extra comments as labels, it is probably unnecessary.
    return ""
}
*/

func (sourcefet *TtSourceFet) GetNActivities() int {
	return len(sourcefet.ActivityElements)
}

func (sourcefet *TtSourceFet) GetNConstraints() ConstraintIndex { return sourcefet.NConstraints }
func (sourcefet *TtSourceFet) GetConstraint_Types() []ConstraintType {
	return sourcefet.ConstraintTypes
}
func (sourcefet *TtSourceFet) GetHardConstraintMap() map[ConstraintType][]ConstraintIndex {
	return sourcefet.HardConstraintMap
}
func (sourcefet *TtSourceFet) GetSoftConstraintMap() map[ConstraintType][]ConstraintIndex {
	return sourcefet.SoftConstraintMap
}

// Rebuild the FET file given an array detailing which constraints are enabled.
// The `xmlp` argument is a pointer to a byte slice, to receive the
// XML FET-file.
// TODO: The xmlp parameter is specifically for FET, so it shouldn't be in the
// interface in this form. Also this should be a method on the back-end data.
// So the back-end data needs ConstraintElements and Doc, presumably also the
// byte buffer, or the method should handle the file writing.
func (sourcefet *TtSourceFet) PrepareRun(enabled []bool, xmlp any) {
	for i, c := range sourcefet.ConstraintElements {
		active := c.SelectElement("Active")
		if enabled[i] {
			active.SetText("true")
		} else {
			active.SetText("false")
		}
	}
	/* TODO: What was the point of all this? !!!
	   root := sourcefet.Doc.Root()
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
	*/
	sourcefet.Doc.Indent(2)
	var err error
	*(xmlp.(*[]byte)), err = sourcefet.Doc.WriteToBytes()
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

func (sourcefet *TtSourceFet) FetWeight(w int) string {
	if w <= 0 {
		return "0"
	}
	if w >= 100 {
		return "100"
	}
	return strconv.FormatFloat(sourcefet.WeightTable[w], 'f', 3, 64)
}

func (sourcefet *TtSourceFet) DbWeight(w string) int {
	wf, err := strconv.ParseFloat(w, 64)
	if err != nil {
		panic(err)
	}
	wdb, _ := slices.BinarySearch(sourcefet.WeightTable, wf)
	return wdb
}
