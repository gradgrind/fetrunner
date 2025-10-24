package fet

import (
	"fetrunner/autotimetable"
	"fmt"

	"github.com/beevik/etree"
)

type ConstraintIndex = autotimetable.ConstraintIndex
type BasicData = autotimetable.BasicData
type ConstraintType = autotimetable.ConstraintType
type ActivityIndex = autotimetable.ActivityIndex

// In FET there are "time" constraints and "space" constraints. The
// `ConstraintData` structure lumps them all together, so there is just
// one constraint list in `FetDoc`. However, the "time" constraints are
// placed first in the list, so by recording the start index of the "space"
// constraints (`NTimeConstraints`) they can be differentiated.
type FetDoc struct {
	Doc              *etree.Document
	Constraints      []*etree.Element // list of actual constraint elements
	NTimeConstraints ConstraintIndex
	// The "non-negotiable" constraints are not dealt with in the
	// "autotimetable" package, but they are included in the `Constraints`
	// list. The `Necessary` list can be used to ensure they are always
	// enabled.
	Necessary []ConstraintIndex
}

func FetRead(cdata *BasicData, fetpath string) (*FetDoc, error) {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fetpath); err != nil {
		panic(err)
	}

	root := doc.Root()
	fmt.Printf("ROOT element: %s (%+v)\n", root.Tag, root.Attr)
	for i, e := range root.ChildElements() {
		fmt.Printf(" -- %02d: %s\n", i, e.Tag)
	}

	fmt.Println("\n  --------------------------")

	ael := root.SelectElement("Activities_List")
	cdata.NActivities = ActivityIndex(len(ael.ChildElements()))

	// Collect the constraints, dividing into soft and hard groups.
	// Note that non-negotiable "basic" constraints are not included
	// in the maps, but they are included in the `constraints` list.

	constraints := []*etree.Element{}
	hard_constraint_map := map[ConstraintType][]ConstraintIndex{}
	soft_constraint_map := map[ConstraintType][]ConstraintIndex{}
	constraint_types := []ConstraintType{}
	necessary := []ConstraintIndex{}
	// Collect time constraints
	et := root.SelectElement("Time_Constraints_List")
	for _, e := range et.ChildElements() {
		i := len(constraints)
		constraints = append(constraints, e)
		ctype := ConstraintType(e.Tag)
		w := e.SelectElement("Weight_Percentage").Text()
		//fmt.Printf(" ++ %02d: %s (%s)\n", i, ctype, w)
		if ctype == "ConstraintBasicCompulsoryTime" {
			// Basic, non-negotiable constraint
			necessary = append(necessary, ConstraintIndex(i))
			continue
		}
		constraint_types = append(constraint_types, ctype)
		// ... duplicates wil be removed in `sort_constraint_types`
		if w == "100" {
			// Hard constraint
			hard_constraint_map[ctype] = append(hard_constraint_map[ctype],
				ConstraintIndex(i))
		} else {
			// Soft constraint
			soft_constraint_map[ctype] = append(soft_constraint_map[ctype],
				ConstraintIndex(i))
		}
	}
	n_time_constraints := len(constraints)
	// Collect space constraints
	et = root.SelectElement("Space_Constraints_List")
	for _, e := range et.ChildElements() {
		i := len(constraints)
		constraints = append(constraints, e)
		ctype := ConstraintType(e.Tag)
		w := e.SelectElement("Weight_Percentage").Text()
		//fmt.Printf(" ++ %02d: %s (%s)\n", i, ctype, w)
		if ctype == "ConstraintBasicCompulsorySpace" {
			// Basic, non-negotiable constraint
			necessary = append(necessary, ConstraintIndex(i))
			continue
		}
		constraint_types = append(constraint_types, ctype)
		// ... duplicates wil be removed in `sort_constraint_types`
		if w == "100" {
			// Hard constraint
			hard_constraint_map[ctype] = append(hard_constraint_map[ctype],
				ConstraintIndex(i))
		} else {
			// Soft constraint
			soft_constraint_map[ctype] = append(soft_constraint_map[ctype],
				ConstraintIndex(i))
		}
	}

	fetdoc := &FetDoc{
		Doc:              doc,
		Constraints:      constraints,
		NTimeConstraints: ConstraintIndex(n_time_constraints),
		Necessary:        necessary,
	}

	cdata.NConstraints = ConstraintIndex(len(constraints))
	cdata.ConstraintTypes = sort_constraint_types(constraint_types)
	cdata.HardConstraintMap = hard_constraint_map
	cdata.SoftConstraintMap = soft_constraint_map
	cdata.Resources = get_resources(root)

	//doc.Indent(2)
	return fetdoc, nil
}

func get_resources(root *etree.Element) []autotimetable.Resource {
	resources := []autotimetable.Resource{}
	el := root.SelectElement("Rooms_List")
	i := 0
	for _, e := range el.ChildElements() {
		if e.SelectElement("Virtual").Text() == "false" {
			tag := e.SelectElement("Name").Text()
			resources = append(resources, autotimetable.Resource{
				Index: i,
				Type:  autotimetable.RoomResource,
				Tag:   tag,
			})
		}
	}

	//TODO: teachers and groups

	return resources
}

// Get a string representation of the given constraint.
func (fetdoc *FetDoc) ConstraintString(cix ConstraintIndex) string {
	e := fetdoc.Constraints[cix]
	d := etree.NewDocument()
	d.SetRoot(e)
	//d.Indent(2) // with newlines and indentation
	d.Unindent() // no newlines or indentation
	s, err := d.WriteToString()
	if err != nil {
		panic(err)
	}
	return s
}

func (fetdoc *FetDoc) WriteFET(fetfile string) {
	err := fetdoc.Doc.WriteToFile(fetfile)
	if err != nil {
		panic(err)
	}
}

// Rebuild the FET file given an array detailing which constraints are
// enabled.
// Because it modifies the data in the shared `FetDoc`, this function
// is not thread-safe!
// The `xmlp` argument is a pointer to a byte slice, to receive the
// XML FET-file.
func (fetdoc *FetDoc) PrepareRun(enabled []bool, xmlp any) {
	for _, i := range fetdoc.Necessary {
		enabled[i] = true
	}
	doc := fetdoc.Doc
	root := doc.Root()
	et := root.SelectElement("Time_Constraints_List")
	var i ConstraintIndex = 0
	var e *etree.Element
	for _, e = range et.ChildElements() {
		active := e.SelectElement("Active")
		if enabled[i] {
			active.SetText("true")
		} else {
			active.SetText("false")
		}
		i++
	}
	et = root.SelectElement("Space_Constraints_List")
	for _, e = range et.ChildElements() {
		active := e.SelectElement("Active")
		if enabled[i] {
			active.SetText("true")
		} else {
			active.SetText("false")
		}
		i++
	}
	var err error
	*(xmlp.(*[]byte)), err = fetdoc.Doc.WriteToBytes()
	if err != nil {
		panic(err)
	}
}
