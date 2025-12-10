package fet

import (
	"encoding/json"
	"fetrunner/internal/base"
	"fmt"
	"regexp"
	"strconv"

	"github.com/beevik/etree"
)

// In FET there are "time" constraints and "space" constraints. They are
// all lumped together in th `ConstraintElements` list, but their indexes
// are also recorded in the `TimeConstraints` and `SpaceConstraints` lists.

func FetRead(bdata *base.BaseData, fetpath string) *TtRunDataFet {
	logger := bdata.Logger
	logger.Info("SOURCE: %s\n", fetpath)
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fetpath); err != nil {
		logger.Error("%s", err)
		return nil
	}
	rundata := &TtRunDataFet{Doc: doc}
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
		aidlist := []IdPair{}
		ael := fetroot.SelectElement("Activities_List")
		inactive := 0
		for _, a := range ael.ChildElements() {
			if a.SelectElement("Active").Text() == "true" {
				activities = append(activities, a)
				id := a.SelectElement("Id").Text()
				// In this case both source and back-end are FET:
				aidlist = append(aidlist, IdPair{Source: id, Backend: id})
			} else {
				inactive++
			}
		}
		if inactive != 0 {
			logger.Result("INACTIVE_ACTIVITIES", strconv.Itoa(inactive))
		}
		rundata.ActivityElements = activities
		rundata.ActivityIds = aidlist
	}

	// Get resource lists, etc.
	rundata.DayIds = get_days(fetroot)
	rundata.HourIds = get_hours(fetroot)
	rundata.RoomIds = get_rooms(fetroot)
	rundata.TeacherIds = get_teachers(fetroot)
	rundata.SubjectIds = get_subjects(fetroot)
	rundata.ClassIds = get_classes(fetroot)

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
			//fmt.Printf(" ++ %02d: %s (%s)\n", i, ctype, w)
			if w == "100" {
				// Hard constraint
				hard_constraint_map[ctype] = append(hard_constraint_map[ctype],
					ConstraintIndex(i))
			} else {
				// Soft constraint
				soft_constraint_map[ctype] = append(soft_constraint_map[ctype],
					ConstraintIndex(i))
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
			constraint_counter++
			cid := fmt.Sprintf("[%d]", constraint_counter)
			comments.SetText(cid + comment)
			rundata.Constraints = append(rundata.Constraints, Constraint{
				IdPair: IdPair{Backend: cid},
				Ctype:  ctype,
				Weight: rundata.DbWeight(w),
			})
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
	rundata.ConstraintTypes = SortConstraintTypes(constraint_types)
	rundata.HardConstraintMap = hard_constraint_map
	rundata.SoftConstraintMap = soft_constraint_map

	return rundata
}

func get_days(fetroot *etree.Element) []IdPair {
	items := []IdPair{}
	for _, e := range fetroot.SelectElement("Days_List").SelectElements("Day") {
		id := e.SelectElement("Name").Text()
		items = append(items, IdPair{
			Backend: id,
			Source:  id,
		})
	}
	return items
}

func get_hours(fetroot *etree.Element) []IdPair {
	hours := []IdPair{}
	for _, e := range fetroot.SelectElement("Hours_List").SelectElements("Hour") {
		id := e.SelectElement("Name").Text()
		hours = append(hours, IdPair{
			Backend: id,
			Source:  id,
		})
	}
	return hours
}

func get_rooms(fetroot *etree.Element) []IdPair {
	rooms := []IdPair{}
	for _, e := range fetroot.SelectElement("Rooms_List").SelectElements("Room") {
		if e.SelectElement("Virtual").Text() == "false" {
			id := e.SelectElement("Name").Text()
			rooms = append(rooms, IdPair{
				Backend: id,
				Source:  id,
			})
		}
	}
	return rooms
}

func get_teachers(fetroot *etree.Element) []IdPair {
	items := []IdPair{}
	for _, e := range fetroot.SelectElement("Teachers_List").SelectElements("Teacher") {
		id := e.SelectElement("Name").Text()
		items = append(items, IdPair{
			Backend: id,
			Source:  id,
		})
	}
	return items
}

func get_classes(fetroot *etree.Element) []IdPair {
	items := []IdPair{}
	for _, e := range fetroot.SelectElement("Students_List").SelectElements("Year") {
		id := e.SelectElement("Name").Text()
		items = append(items, IdPair{
			Backend: id,
			Source:  id,
		})
	}
	return items
}

func get_subjects(fetroot *etree.Element) []IdPair {
	items := []IdPair{}
	for _, e := range fetroot.SelectElement("Subjects_List").SelectElements("Subject") {
		id := e.SelectElement("Name").Text()
		items = append(items, IdPair{
			Backend: id,
			Source:  id,
		})
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
