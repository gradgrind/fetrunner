package makefet

import (
	"fetrunner/autotimetable"
	"fetrunner/base"
	"fetrunner/fet"
	"fetrunner/timetable"
	"fmt"
	"math"
	"strconv"

	"github.com/beevik/etree"
)

type NodeRef = base.NodeRef

const CLASS_GROUP_SEP = "."
const VIRTUAL_ROOM_PREFIX = "!"

// const LUNCH_BREAK_TAG = "-lb-"
// const LUNCH_BREAK_NAME = "Lunch Break"

// Use a `FetBuild` as basis for constructing a `fet.TtRunDataFet`. In addition,
// some fields of the `autotimetable.BasicData` are initialized.
func FetTree(
	attdata *autotimetable.AutoTtData,
	bdata *base.BaseData,
	tt_data *timetable.TtData,
) *fet.TtRunDataFet {
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	rundata := &fet.TtRunDataFet{
		Doc:         doc,
		WeightTable: fet.MakeFetWeights(),
	}
	attdata.Source = rundata

	fetbuild := &FetBuild{
		basedata:           bdata,
		ttdata:             tt_data,
		rundata:            rundata,
		fet_virtual_rooms:  map[string]string{},
		fet_virtual_room_n: map[string]int{},
	}

	fetroot := doc.CreateElement("fet")
	fetbuild.fetroot = fetroot
	fetroot.CreateAttr("version", fet_version)
	fetroot.CreateElement("Mode").SetText("Official")
	fetroot.CreateElement("Institution_Name").SetText(
		bdata.Db.Info.Institution)

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
	//   NodeRef -> []db.TimeSlot
	namap := fetbuild.blocked_slots()

	//TODO: Handle WITHOUT_ROOM_CONSTRAINTS
	fetbuild.add_placement_constraints(false)

	fetbuild.add_activity_constraints()

	fetbuild.add_class_constraints(namap)
	fetbuild.add_teacher_constraints(namap)

	//TODO: The remaining constraints

	// Number of activities
	attdata.NActivities = len(rundata.ActivityIds)

	// Collect the constraints, dividing into soft and hard groups.
	hard_constraint_map := map[string][]int{}
	soft_constraint_map := map[string][]int{}
	constraint_types := []string{}
	for i, c := range rundata.Constraints {

		constraint_types = append(constraint_types, c.Ctype)
		// ... duplicates wil be removed in `sort_constraint_types`

		if c.Weight == base.MAXWEIGHT {
			// Hard constraint
			hard_constraint_map[c.Ctype] = append(
				hard_constraint_map[c.Ctype], i)
		} else {
			// Soft constraint
			soft_constraint_map[c.Ctype] = append(
				soft_constraint_map[c.Ctype], i)
		}
	}
	attdata.NConstraints = len(rundata.Constraints)
	attdata.ConstraintTypes = fet.SortConstraintTypes(constraint_types)
	attdata.HardConstraintMap = hard_constraint_map
	attdata.SoftConstraintMap = soft_constraint_map

	//

	return fetbuild.rundata
}

func oldweight2fet(w int) string {
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
	ctype string, id NodeRef, index int, weight int,
) Constraint {
	return Constraint{
		IdPair:     IdPair{Source: string(id)},
		Ctype:      ctype,
		Parameters: []int{index},
		Weight:     weight}
}

func params_constraint(
	ctype string, id NodeRef, indexlist []int, weight int,
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
	e.CreateElement("Comments").SetText(c.Backend)

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
	e.CreateElement("Comments").SetText(c.Backend)

	rundata.Constraints = append(rundata.Constraints, c)
}
