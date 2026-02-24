package fet

import (
	"encoding/json"
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fmt"
	"regexp"
	"strconv"

	"github.com/beevik/etree"
)

// In FET there are "time" constraints and "space" constraints. They are
// all lumped together in the `ConstraintElements` list, but their indexes
// are also recorded in the `TimeConstraints` and `SpaceConstraints` lists.

func FetRead(
	bdata *base.BaseData,
	fetpath string,
) *TtSourceFet {
	logger := bdata.Logger
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fetpath); err != nil {
		logger.Error("%s", err)
		return nil
	}
	rundata := &TtSourceFet{Doc: doc, WeightTable: MakeFetWeights()}
	//fmt.Printf("rundata.WeightTable = %+v\n\n", rundata.WeightTable)
	fetroot := doc.Root()

	/*
	   fmt.Printf("ROOT element: %s (%+v)\n", root.Tag, root.Attr)
	   for i, e := range root.ChildElements() {
	       fmt.Printf(" -- %02d: %s\n", i, e.Tag)
	   }
	   fmt.Println("\n  --------------------------")
	*/

	// Get active activities, count inactive ones
	{
		activities := []*etree.Element{}
		aidlist := []TtSourceItem{}
		ael := fetroot.SelectElement("Activities_List")
		inactive := 0
		i := 0
		for _, a := range ael.ChildElements() {
			if a.SelectElement("Active").Text() == "true" {
				activities = append(activities, a)
				id := a.SelectElement("Id").Text()
				aidlist = append(aidlist, TtSourceItem{Index: i, Tag: id})
				i++
			} else {
				inactive++
			}
		}
		if inactive != 0 {
			logger.Result("INACTIVE_ACTIVITIES", strconv.Itoa(inactive))
		}
		rundata.ActivityElements = activities
		rundata.ActivityList = aidlist
	}

	// Get resource lists, etc.
	rundata.DayList = get_days(fetroot)
	rundata.HourList = get_hours(fetroot)
	rundata.RoomList = get_rooms(fetroot)
	rundata.TeacherList = get_teachers(fetroot)
	rundata.SubjectList = get_subjects(fetroot)
	rundata.ClassList = get_classes(fetroot)

	// Collect the constraints, dividing into soft and hard groups.
	// Inactive constraints will be removed.
	r_constraint_number := regexp.MustCompile(`^\[[0-9]+\](.*)$`)
	constraint_counter := 0
	hard_constraint_map := map[ConstraintType][]ConstraintIndex{}
	soft_constraint_map := map[ConstraintType][]ConstraintIndex{}
	constraint_types := []ConstraintType{}

	for timespace := range 2 {
		// First (timespace == 0) collect active time constraints,
		// then (timespace == 1) collect active space constraints.

		var et *etree.Element
		var bc string
		if timespace == 0 {
			et = fetroot.SelectElement("Time_Constraints_List")
			bc = "ConstraintBasicCompulsoryTime"
		} else {
			et = fetroot.SelectElement("Space_Constraints_List")
			bc = "ConstraintBasicCompulsorySpace"
		}
		inactive := 0
		for _, e := range et.ChildElements() {
			// Count and skip if inactive
			if e.SelectElement("Active").Text() == "false" {
				inactive++ // count inactive constraints
				continue
			}
			ctype := ConstraintType(e.Tag)
			if ctype == bc {
				// Basic, non-negotiable constraint
				continue
			}
			i := len(rundata.ConstraintElements)
			rundata.ConstraintElements = append(rundata.ConstraintElements, e)
			if timespace == 0 {
				rundata.TimeConstraints = append(rundata.TimeConstraints, i)
			} else {
				rundata.SpaceConstraints = append(rundata.SpaceConstraints, i)
			}

			w := e.SelectElement("Weight_Percentage").Text()
			wdb := rundata.DbWeight(w)
			//fmt.Printf(" ++ %02d: %s (%s -> %02d)\n", i, ctype, w, wdb)
			if w == "100" {
				// Hard constraint
				hard_constraint_map[ctype] = append(hard_constraint_map[ctype],
					ConstraintIndex(i))
			} else {
				// Soft constraint
				wctype := fmt.Sprintf("%02d:%s", wdb, ctype)
				soft_constraint_map[wctype] = append(soft_constraint_map[wctype],
					ConstraintIndex(i))
				rundata.SoftWeights = append(rundata.SoftWeights,
					SoftWeight{i, w})
			}
			constraint_types = append(constraint_types, ctype)
			// ... duplicates wil be removed in `sort_constraint_types`

			// Ensure that the constraints are numbered in their Comments.
			// This is to ease referencing in the results object.
			comments := e.SelectElement("Comments")
			comment := ""
			if comments == nil {
				comments = e.CreateElement("Comments")
			} else {
				// Remove any existing comment id
				comment = comments.Text()
				parts := r_constraint_number.FindStringSubmatch(comment)
				if parts != nil {
					comment = parts[1]
				}
			}
			wtag := ""
			if w != "100" {
				wtag = ":" + w
			}
			// In FET, the constraints have no identifiers/tags, so one is
			// added in the "Comments"  field.
			cid := fmt.Sprintf("[%d%s]", constraint_counter, wtag)
			comments.SetText(cid + comment)
			rundata.Constraints = append(rundata.Constraints, Constraint{
				TtSourceItem: TtSourceItem{Index: constraint_counter, Tag: cid},
				Ctype:        ctype,
				Weight:       wdb,
			})
			constraint_counter++
		}
		if inactive != 0 {
			if timespace == 0 {
				logger.Result("INACTIVE_TIME_CONSTRAINTS", strconv.Itoa(inactive))
			} else {
				logger.Result("INACTIVE_SPACE_CONSTRAINTS", strconv.Itoa(inactive))
			}
		}
	}

	rundata.NConstraints = ConstraintIndex(constraint_counter)
	rundata.ConstraintTypes = autotimetable.SortConstraintTypes(
		constraint_types, ConstraintPriority)
	rundata.HardConstraintMap = hard_constraint_map
	rundata.SoftConstraintMap = soft_constraint_map

	return rundata
}

func get_days(fetroot *etree.Element) []TtSourceItem {
	items := []TtSourceItem{}
	for i, e := range fetroot.SelectElement("Days_List").SelectElements("Day") {
		id := e.SelectElement("Name").Text()
		items = append(items, TtSourceItem{Index: i, Tag: id})
	}
	return items
}

func get_hours(fetroot *etree.Element) []TtSourceItem {
	hours := []TtSourceItem{}
	for i, e := range fetroot.SelectElement("Hours_List").SelectElements("Hour") {
		id := e.SelectElement("Name").Text()
		hours = append(hours, TtSourceItem{Index: i, Tag: id})
	}
	return hours
}

func get_rooms(fetroot *etree.Element) []TtSourceItem {
	rooms := []TtSourceItem{}
	i := 0
	for _, e := range fetroot.SelectElement("Rooms_List").SelectElements("Room") {
		if e.SelectElement("Virtual").Text() == "false" {
			id := e.SelectElement("Name").Text()
			rooms = append(rooms, TtSourceItem{Index: i, Tag: id})
			i++
		}
	}
	return rooms
}

func get_teachers(fetroot *etree.Element) []TtSourceItem {
	items := []TtSourceItem{}
	for i, e := range fetroot.SelectElement("Teachers_List").SelectElements("Teacher") {
		id := e.SelectElement("Name").Text()
		items = append(items, TtSourceItem{Index: i, Tag: id})
	}
	return items
}

func get_classes(fetroot *etree.Element) []TtSourceItem {
	items := []TtSourceItem{}
	for i, e := range fetroot.SelectElement("Students_List").SelectElements("Year") {
		id := e.SelectElement("Name").Text()
		items = append(items, TtSourceItem{Index: i, Tag: id})
	}
	return items
}

func get_subjects(fetroot *etree.Element) []TtSourceItem {
	items := []TtSourceItem{}
	for i, e := range fetroot.SelectElement("Subjects_List").SelectElements("Subject") {
		id := e.SelectElement("Name").Text()
		items = append(items, TtSourceItem{Index: i, Tag: id})
	}
	return items
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
