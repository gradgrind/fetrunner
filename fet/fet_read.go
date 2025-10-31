package fet

import (
	"encoding/json"
	"fetrunner/autotimetable"
	"fetrunner/base"
	"strconv"

	"github.com/beevik/etree"
)

type ConstraintIndex = autotimetable.ConstraintIndex
type BasicData = autotimetable.BasicData
type ConstraintType = autotimetable.ConstraintType

// In FET there are "time" constraints and "space" constraints. The
// `ConstraintData` structure lumps them all together, so there is just
// one constraint list in `FetDoc`. However, the "time" constraints are
// placed first in the list, so by recording the start index of the "space"
// constraints (`NTimeConstraints`) they can be differentiated.
type FetDoc struct {
	Doc              *etree.Document
	Activities       []*etree.Element // list of active activity elements
	Constraints      []*etree.Element // list of active constraint elements
	NTimeConstraints ConstraintIndex
	// The "non-negotiable" constraints are not dealt with in the
	// "autotimetable" package, but they are included in the `Constraints`
	// list. The `Necessary` list can be used to ensure they are always
	// enabled.
	Necessary []ConstraintIndex
}

func FetRead(cdata *BasicData, fetpath string) (*FetDoc, error) {
	base.Message.Printf("SOURCE: %s\n", fetpath)
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fetpath); err != nil {
		panic(err)
	}
	root := doc.Root()

	/*
		fmt.Printf("ROOT element: %s (%+v)\n", root.Tag, root.Attr)
		for i, e := range root.ChildElements() {
			fmt.Printf(" -- %02d: %s\n", i, e.Tag)
		}
		fmt.Println("\n  --------------------------")
	*/

	// Get active activities, count inactive ones
	activities := []*etree.Element{}
	{
		ael := root.SelectElement("Activities_List")
		inactive := 0
		for _, a := range ael.ChildElements() {
			if a.SelectElement("Active").Text() == "true" {
				activities = append(activities, a)
			} else {
				inactive++
			}
		}
		cdata.NActivities = len(activities)
		if inactive != 0 {
			base.Message.Printf("-A- %d inactive activities", inactive)
		}
	}

	// Collect the constraints, dividing into soft and hard groups.
	// Note that non-negotiable "basic" constraints are not included
	// in the maps, but they are included in the `constraints` list.
	// Inactive constraints will be removed

	constraints := []*etree.Element{}
	hard_constraint_map := map[ConstraintType][]ConstraintIndex{}
	soft_constraint_map := map[ConstraintType][]ConstraintIndex{}
	constraint_types := []ConstraintType{}
	necessary := []ConstraintIndex{}
	// Collect active time constraints
	var n_time_constraints int
	{
		et := root.SelectElement("Time_Constraints_List")
		inactive := 0
		for _, e := range et.ChildElements() {
			// Count and skip if inactive
			if e.SelectElement("Active").Text() == "false" {
				inactive++ // count inactive constraints
				continue
			}
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
		if inactive != 0 {
			base.Message.Printf("-T- %d inactive time constraints", inactive)
		}
		n_time_constraints = len(constraints)
	}
	// Collect active space constraints
	{
		et := root.SelectElement("Space_Constraints_List")
		inactive := 0
		for _, e := range et.ChildElements() {
			// Count and skip if inactive
			if e.SelectElement("Active").Text() == "false" {
				et.RemoveChild(e)
				inactive++ // count removed constraints
				continue
			}
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
		if inactive != 0 {
			base.Message.Printf("-S- %d inactive space constraints", inactive)
		}
	}

	fetdoc := &FetDoc{
		Doc:              doc,
		Activities:       activities,
		Constraints:      constraints,
		NTimeConstraints: ConstraintIndex(n_time_constraints),
		Necessary:        necessary,
	}

	cdata.NConstraints = ConstraintIndex(len(constraints))
	cdata.ConstraintTypes = sort_constraint_types(constraint_types)
	cdata.HardConstraintMap = hard_constraint_map
	cdata.SoftConstraintMap = soft_constraint_map

	//doc.Indent(2)
	return fetdoc, nil
}

// The ActivityIds are the (FET) Activity Id (int) and the content of the
// "Comments" field (string).
func (fetdoc *FetDoc) GetActivityIds() []autotimetable.ActivityId {
	alist := []autotimetable.ActivityId{}
	for _, a := range fetdoc.Activities {
		id := a.SelectElement("Id").Text()
		i, err := strconv.Atoi(id)
		if err != nil {
			panic("Activity Id is not an integer: " + id)
		}
		alist = append(alist, autotimetable.ActivityId{
			Id: i, Ref: a.SelectElement("Comments").Text()})
	}
	return alist
}

func (fetdoc *FetDoc) GetDayTags() []string {
	root := fetdoc.Doc.Root()
	days := []string{}
	for _, e := range root.SelectElement("Days_List").SelectElements("Day") {
		days = append(days, e.SelectElement("Name").Text())
	}
	return days
}

func (fetdoc *FetDoc) GetHourTags() []string {
	root := fetdoc.Doc.Root()
	hours := []string{}
	for _, e := range root.SelectElement("Hours_List").SelectElements("Hour") {
		hours = append(hours, e.SelectElement("Name").Text())
	}
	return hours
}

func (fetdoc *FetDoc) GetRooms() []autotimetable.TtItem {
	root := fetdoc.Doc.Root()
	rooms := []autotimetable.TtItem{}
	i := 0
	for _, e := range root.SelectElement("Rooms_List").ChildElements() {
		if e.SelectElement("Virtual").Text() == "false" {
			tag := e.SelectElement("Name").Text()
			data := e.SelectElement("Comments").Text()
			rooms = append(rooms, autotimetable.TtItem{
				Key:  data,
				Text: tag,
			})
			i++
		}
	}
	return rooms
}

// Get a key and string representation of the constraints.
func (fetdoc *FetDoc) GetConstraintItems() []autotimetable.TtItem {
	clist := []autotimetable.TtItem{}
	for _, c := range fetdoc.Constraints {
		// Make a JSON version of the constraint's XML
		s := WriteElement(c)
		key := ""
		ce := c.SelectElement("Comments")
		if ce != nil {
			key = ce.Text()
		}
		clist = append(clist, autotimetable.TtItem{
			Key:  key,
			Text: s,
		})
	}
	return clist
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
	for i, c := range fetdoc.Constraints {
		active := c.SelectElement("Active")
		if enabled[i] {
			active.SetText("true")
		} else {
			active.SetText("false")
		}
	}
	root := fetdoc.Doc.Root()
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
	*(xmlp.(*[]byte)), err = fetdoc.Doc.WriteToBytes()
	if err != nil {
		panic(err)
	}
}

// Generate a JSON version of the given element. Only a simple subset of
// XML is covered, but it should be enough for a FET file.
func WriteElement(e *etree.Element) string {
	k, v := jsonElement(e)
	if v == nil {
		return ""
	}
	m := map[string]any{}
	m[k] = v
	//jsonBytes, err := json.MarshalIndent(m, "", "   ")
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func jsonElement(e *etree.Element) (string, any) {
	children := e.ChildElements()
	if len(children) == 0 {
		v := e.Text()
		if len(v) == 0 {
			return "", nil
		}
		return e.FullTag(), v
	}
	m0 := map[string][]any{}
	for _, c := range children {
		k, v := jsonElement(c)
		if v == nil {
			continue
		}
		m0[k] = append(m0[k], v)
	}
	if len(m0) == 0 {
		return "", nil
	}
	m := map[string]any{}
	for k, v := range m0 {
		if len(v) == 1 {
			m[k] = v[0]
		} else {
			m[k] = v
		}
	}
	return e.FullTag(), m
}
