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

type element = base.ElementBase

type constraint = autotimetable.AttConstraint
type constraintIndex = autotimetable.ConstraintIndex
type autoTtData = autotimetable.AutoTtData
type constraintType = autotimetable.ConstraintType

type softWeight struct {
	Index  constraintIndex
	Weight string
}

type TtSourceFet struct {
	doc                *etree.Document
	constraintElements []*etree.Element
	softWeights        []softWeight

	activityElements []*etree.Element

	// FET has time and space constraints separate. It might be useful in
	// some way to have that information here.
	timeConstraints  []int // indexes into `ConstraintElements`
	spaceConstraints []int // indexes into `ConstraintElements`

	constraints []constraint

	nConstraints      constraintIndex
	constraintTypes   []constraintType
	hardConstraintMap map[constraintType][]constraintIndex
	softConstraintMap map[constraintType][]constraintIndex
}

func (sourcefet *TtSourceFet) Prepare(real_soft bool) {
	for _, cw := range sourcefet.softWeights {
		e := sourcefet.constraintElements[cw.Index]
		if real_soft {
			e.SelectElement("Weight_Percentage").SetText(cw.Weight)
		} else {
			e.SelectElement("Weight_Percentage").SetText("100")
		}
	}
}

func (sourcefet *TtSourceFet) GetDays() []element {
	items := []element{}
	for _, e := range sourcefet.doc.Root().SelectElement("Days_List").SelectElements("Day") {
		id := e.SelectElement("Name").Text()
		items = append(items, element{
			// In FET the ID is also the tag
			Id: base.NodeRef(id), Tag: id})
	}
	return items
}

func (sourcefet *TtSourceFet) GetHours() []element {
	items := []element{}
	for _, e := range sourcefet.doc.Root().SelectElement("Hours_List").SelectElements("Hour") {
		id := e.SelectElement("Name").Text()
		items = append(items, element{
			// In FET the ID is also the tag
			Id: base.NodeRef(id), Tag: id})
	}
	return items
}

func (sourcefet *TtSourceFet) GetTeachers() []element {
	items := []element{}
	for _, e := range sourcefet.doc.Root().SelectElement("Teachers_List").SelectElements("Teacher") {
		id := e.SelectElement("Name").Text()
		items = append(items, element{
			// In FET the ID is also the tag
			Id: base.NodeRef(id), Tag: id})
	}
	return items
}

func (sourcefet *TtSourceFet) GetSubjects() []element {
	items := []element{}
	for _, e := range sourcefet.doc.Root().SelectElement("Subjects_List").SelectElements("Subject") {
		id := e.SelectElement("Name").Text()
		items = append(items, element{
			// In FET the ID is also the tag
			Id: base.NodeRef(id), Tag: id})
	}
	return items
}

func (sourcefet *TtSourceFet) GetRooms() []element {
	items := []element{}
	for _, e := range sourcefet.doc.Root().SelectElement("Rooms_List").SelectElements("Room") {
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
	for _, e := range sourcefet.doc.Root().SelectElement("Students_List").SelectElements("Year") {
		id := e.SelectElement("Name").Text()
		items = append(items, element{
			// In FET the ID is also the tag
			Id: base.NodeRef(id), Tag: id})
	}
	return items
}

func (sourcefet *TtSourceFet) GetActivities() []element {
	aidlist := make([]element, len(sourcefet.activityElements))
	for i, a := range sourcefet.activityElements {
		aidlist[i] = element{
			//No Id
			Tag: a.SelectElement("Id").Text()}
	}
	return aidlist
}

func (sourcefet *TtSourceFet) GetConstraints() []constraint { return sourcefet.constraints }

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
	return len(sourcefet.activityElements)
}

func (sourcefet *TtSourceFet) GetNConstraints() constraintIndex { return sourcefet.nConstraints }
func (sourcefet *TtSourceFet) GetConstraint_Types() []constraintType {
	return sourcefet.constraintTypes
}
func (sourcefet *TtSourceFet) GetHardConstraintMap() map[constraintType][]constraintIndex {
	return sourcefet.hardConstraintMap
}
func (sourcefet *TtSourceFet) GetSoftConstraintMap() map[constraintType][]constraintIndex {
	return sourcefet.softConstraintMap
}

// Rebuild the FET file given an array detailing which constraints are enabled.
// The `xmlp` argument is a pointer to a byte slice, to receive the
// XML FET-file.
// TODO: The xmlp parameter is specifically for FET, so it shouldn't be in the
// interface in this form. Also this should be a method on the back-end data.
// So the back-end data needs ConstraintElements and Doc, presumably also the
// byte buffer, or the method should handle the file writing.
func (sourcefet *TtSourceFet) PrepareRun(enabled []bool, xmlp any) {
	for i, c := range sourcefet.constraintElements {
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
	sourcefet.doc.Indent(2)
	var err error
	*(xmlp.(*[]byte)), err = sourcefet.doc.WriteToBytes()
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

// TODO: Where is this needed?
func (sourcefet *TtSourceFet) DbWeight2Fet(w int, weightTable []float64) string {
	if w <= 0 {
		return "0"
	}
	if w >= 100 {
		return "100"
	}
	return strconv.FormatFloat(weightTable[w], 'f', 3, 64)
}

func FetWeight2Db(w string, weightTable []float64) int {
	wf, err := strconv.ParseFloat(w, 64)
	if err != nil {
		panic(err)
	}
	wdb, _ := slices.BinarySearch(weightTable, wf)
	return wdb
}
