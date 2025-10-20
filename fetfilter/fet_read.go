package fetfilter

import (
	"fetrunner/autotimetable"
	"fmt"
	"slices"

	"github.com/beevik/etree"
)

type ConstraintIndex = autotimetable.ConstraintIndex
type ConstraintData = autotimetable.ConstraintData
type ConstraintType = autotimetable.ConstraintType

type FetDoc struct {
	Doc             *etree.Document
	Constraints     []*etree.Element
	TimeConstraints ConstraintIndex
}

//TODO: Possibility of sorting the constraint types?

func ReadFet(fetpath string) (*ConstraintData, error) {
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

	// Collect the constraints, dividing into soft and hard groups.
	// Note that non-negotiable "basic" constraints are not included
	// in the maps, but they are included in the `constraints` list.

	constraints := []*etree.Element{}
	hard_constraint_map := map[ConstraintType][]ConstraintIndex{}
	//var basic_time_constraints []*etree.Element
	//var basic_space_constraint *etree.Element
	soft_constraint_map := map[ConstraintType][]ConstraintIndex{}
	constraint_types := []ConstraintType{}
	// Collect time constraints
	et := root.SelectElement("Time_Constraints_List")
	for _, e := range et.ChildElements() {
		i := len(constraints)
		constraints = append(constraints, e)
		ctype := ConstraintType(e.Tag)
		w := e.SelectElement("Weight_Percentage").Text()
		fmt.Printf(" ++ %02d: %s (%s)\n", i, ctype, w)
		if ctype == "ConstraintBasicCompulsoryTime" ||
			(ctype == "ConstraintActivityPreferredStartingTime" &&
				w == "100") {
			// Basic, non-negotiable constraints
			continue
		}
		if !slices.Contains(constraint_types, ctype) {
			constraint_types = append(constraint_types, ctype)
		}
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
		fmt.Printf(" ++ %02d: %s (%s)\n", i, ctype, w)
		if ctype == "ConstraintBasicCompulsorySpace" {
			// Basic, non-negotiable constraints
			continue
		}
		if !slices.Contains(constraint_types, ctype) {
			constraint_types = append(constraint_types, ctype)
		}
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
		Doc:             doc,
		Constraints:     constraints,
		TimeConstraints: ConstraintIndex(n_time_constraints),
	}

	cdata := &ConstraintData{
		InputData:         fetdoc,
		Constraints:       ConstraintIndex(len(constraints)),
		ConstraintTypes:   constraint_types,
		HardConstraintMap: hard_constraint_map,
		SoftConstraintMap: soft_constraint_map,
		ConstraintString:  ConstraintString,
	}

	//doc.Indent(2)
	return cdata, nil
}

// Get a string representation of the given constraint.
func ConstraintString(cdata *ConstraintData, cix ConstraintIndex) string {
	fetdoc, ok := cdata.InputData.(*FetDoc)
	if !ok {
		panic("Bug, BackEndData not FetDoc")
	}
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

func WriteFET(fetfile string, cdata *ConstraintData) {
	fetdoc, ok := cdata.InputData.(*FetDoc)
	if !ok {
		panic("Bug, BackEndData not FetDoc")
	}
	err := fetdoc.Doc.WriteToFile(fetfile)
	if err != nil {
		panic(err)
	}
}
