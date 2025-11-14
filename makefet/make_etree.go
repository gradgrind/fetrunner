package makefet

import (
	"fetrunner/autotimetable"
	"fetrunner/base"
	"fetrunner/db"
	"fetrunner/fet"
	"fetrunner/timetable"
	"fmt"
	"math"
	"regexp"
	"strconv"

	"github.com/beevik/etree"
)

type NodeRef = db.NodeRef

const CLASS_GROUP_SEP = "."
const VIRTUAL_ROOM_PREFIX = "!"

// const LUNCH_BREAK_TAG = "-lb-"
// const LUNCH_BREAK_NAME = "Lunch Break"

// Use a `FetBuild` as basis for constructing a `fet.TtRunDataFet`. In addition,
// some fields of the `autotimetable.BasicData` are initialized.
func FetTree(
	basic_data *autotimetable.BasicData,
	tt_data *timetable.TtData,
) *fet.TtRunDataFet {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	rundata := &fet.TtRunDataFet{Doc: doc}
	basic_data.Source = rundata

	fetbuild := &FetBuild{
		ttdata:             tt_data,
		rundata:            rundata,
		fet_virtual_rooms:  map[string]string{},
		fet_virtual_room_n: map[string]int{},
	}

	//TODO
	institution := "The School"
	fet_version := "6.28.2"

	fetroot := doc.CreateElement("fet")
	fetbuild.fetroot = fetroot
	fetroot.CreateAttr("version", fet_version)
	fetroot.CreateElement("Mode").SetText("Official")
	fetroot.CreateElement("Institution_Name").SetText(institution)

	//TODO?
	source_ref := ""
	fetroot.CreateElement("Comments").SetText(source_ref)

	fetbuild.set_days_hours()
	fetbuild.set_teachers()

	fetbuild.set_subjects()
	fetbuild.set_rooms()
	fetbuild.set_classes()

	fetbuild.activity_tag_list = fetroot.CreateElement("Activity_Tags_List")

	fetbuild.set_activities()

	tclist := fetroot.CreateElement("Time_Constraints_List")
	fetbuild.time_constraints_list = tclist
	bctime := tclist.CreateElement("ConstraintBasicCompulsoryTime")
	bctime.CreateElement("Weight_Percentage").SetText("100")
	bctime.CreateElement("Active").SetText("true")

	sclist := fetroot.CreateElement("Space_Constraints_List")
	fetbuild.space_constraints_list = sclist
	bcspace := sclist.CreateElement("ConstraintBasicCompulsorySpace")
	bcspace.CreateElement("Weight_Percentage").SetText("100")
	bcspace.CreateElement("Active").SetText("true")

	// Add "NotAvailable" constraints for all resources, returning a map
	// linking a resource to its blocked slot list:
	//   db.NodeRef -> []db.TimeSlot
	namap := fetbuild.blocked_slots()

	//TODO: Handle WITHOUT_ROOM_CONSTRAINTS
	fetbuild.add_placement_constraints(false)

	fetbuild.add_activity_constraints()

	fetbuild.add_class_constraints(namap)
	fetbuild.add_teacher_constraints(namap)

	//TODO: The remaining constraints

	//TODO: integrate this somehow ...

	// Number of activities
	basic_data.NActivities = len(rundata.ActivityIds)

	// Collect the constraints, dividing into soft and hard groups.
	r_constraint_number := regexp.MustCompile(`^[0-9]+[)].*`)
	constraint_counter := 0

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

			// Ensure that the constraints are numbered in their Comments.
			// This is to ease referencing in the results object.
			comments := e.SelectElement("Comments")
			comment := ""
			if comments == nil {
				comments = e.CreateElement("Comments")
			} else {
				comment = comments.Text()
				if r_constraint_number.MatchString(comment) {
					goto skip1
				}
			}
			constraint_counter++
			comments.SetText(
				fmt.Sprintf("%d)%s", constraint_counter, comment))
		skip1:

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

			// Ensure that the constraints are numbered in their Comments.
			// This is to ease referencing in the results object.
			comments := e.SelectElement("Comments")
			comment := ""
			if comments == nil {
				comments = e.CreateElement("Comments")
			} else {
				comment = comments.Text()
				if r_constraint_number.MatchString(comment) {
					goto skip2
				}
			}
			constraint_counter++
			comments.SetText(
				fmt.Sprintf("%d)%s", constraint_counter, comment))
		skip2:

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

	cdata.ConstraintTypes = sort_constraint_types(constraint_types)
	cdata.HardConstraintMap = hard_constraint_map
	cdata.SoftConstraintMap = soft_constraint_map

	//

	return fetbuild.rundata
}

func weight2fet(w int) string {
	if w <= 0 {
		return "0"
	}
	if w >= 100 {
		return "100"
	}
	wf := float64(w)
	n := wf + math.Pow(2, wf/12)
	wfet := 100.0 - 100.0/n
	return strconv.FormatFloat(wfet, 'f', 3, 64)
}

func (fetbuild *FetBuild) add_activity_tag(tag string) {
	atag := fetbuild.activity_tag_list.CreateElement("Activity_Tag")
	atag.CreateElement("Name").SetText(tag)
	atag.CreateElement("Printable").SetText("false")
}

func param_constraint(
	ctype string, id db.NodeRef, index int, weight int,
) Constraint {
	return Constraint{
		IdPair:     IdPair{Source: string(id)},
		Ctype:      ctype,
		Parameters: []int{index}}
}

func params_constraint(
	ctype string, id db.NodeRef, indexlist []int, weight int,
) Constraint {
	return Constraint{
		IdPair:     IdPair{Source: string(id)},
		Ctype:      ctype,
		Parameters: indexlist,
		Weight:     weight}
}

func (fetbuild *FetBuild) add_time_constraint(e *etree.Element, c Constraint) {
	rundata := fetbuild.rundata
	i := len(rundata.ConstraintElements)
	rundata.ConstraintElements = append(rundata.ConstraintElements, e)
	rundata.TimeConstraints = append(rundata.TimeConstraints, i)

	// Make a tag for the constraint
	fetbuild.constraint_counter++
	c.Backend = fmt.Sprintf("[%d]", fetbuild.constraint_counter)
	//e.CreateElement("Comments").SetText(c.Backend)

	rundata.Constraints = append(rundata.Constraints, c)
}

func (fetbuild *FetBuild) add_space_constraint(e *etree.Element, c Constraint) {
	rundata := fetbuild.rundata
	i := len(rundata.ConstraintElements)
	rundata.ConstraintElements = append(rundata.ConstraintElements, e)
	rundata.SpaceConstraints = append(rundata.SpaceConstraints, i)

	// Make a tag for the constraint
	fetbuild.constraint_counter++
	c.Backend = fmt.Sprintf("[%d]", fetbuild.constraint_counter)
	//e.CreateElement("Comments").SetText(c.Backend)

	rundata.Constraints = append(rundata.Constraints, c)
}
