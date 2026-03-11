package fet

import (
    "fetrunner/internal/autotimetable"
    "fetrunner/internal/base"
    "fetrunner/internal/timetable"
    "fmt"
    "strconv"

    "github.com/beevik/etree"
)

type NodeRef = base.NodeRef

const CLASS_GROUP_SEP = "."
const VIRTUAL_ROOM_PREFIX = "!"

// const LUNCH_BREAK_TAG = "-lb-"
// const LUNCH_BREAK_NAME = "Lunch Break"

// Construct a `fet_build` from the timetable data in a `timetable.TtData`.
// This `fet_build` needs to contain all the information for generating a
// `FET` file with a subset of constraints. The constraints are determined
// by the source (here `timetable.TtData`) and the possibility to map to
// a variable number of `FET` constraints should be supported.
// Some fields of the `timetable.TtData` and the`autotimetable.BasicData`
// are initialized.
func FetTree(
    bdata *base.BaseData,
    real_soft bool,
    tt_data *timetable.TtData,
) *fet_build {
    doc := etree.NewDocument()
    doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

    fetbuild := &fet_build{
        basedata: bdata,
        ttdata:   tt_data,

        Doc:         doc,
        WeightTable: MakeFetWeights(),

        fet_virtual_rooms:  map[string]string{},
        fet_virtual_room_n: map[string]int{},
        real_soft:          real_soft,
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

    // Collect the constraints, dividing into soft and hard groups.
    hard_constraint_map := map[string][]int{}
    soft_constraint_map := map[string][]int{}
    constraint_types := []string{}
    for i, c := range fetbuild.Constraints {

        constraint_types = append(constraint_types, c.Ctype)
        // ... duplicates wil be removed in `sort_constraint_types`

        if c.Weight == base.MAXWEIGHT {
            // Hard constraint
            hard_constraint_map[c.Ctype] = append(
                hard_constraint_map[c.Ctype], i)
        } else {
            // Soft constraint
            wctype := fmt.Sprintf("%02d:%s", c.Weight, c.Ctype)
            soft_constraint_map[wctype] = append(soft_constraint_map[wctype], i)
        }
    }
    fetbuild.NConstraints = len(fetbuild.Constraints)
    tt_data.ConstraintTypes = autotimetable.SortConstraintTypes(
        constraint_types, base.ConstraintPriority)
    fetbuild.HardConstraintMap = hard_constraint_map
    fetbuild.SoftConstraintMap = soft_constraint_map

    return fetbuild
}

/*
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
*/

// Currently unused
func (fetbuild *fet_build) add_activity_tag(tag string) {
    atag := fetbuild.activity_tag_list.CreateElement("Activity_Tag")
    atag.CreateElement("Name").SetText(tag)
    atag.CreateElement("Printable").SetText("false")
}

// TODO: Where to record the node ref?
func param_constraint(
    ctype string, id NodeRef, index int, weight int,
) constraint {
    return constraint{
        TtSourceTag: string(id),
        Ctype:       ctype,
        Parameters:  []int{index},
        Weight:      weight}
}

func params_constraint(
    ctype string, id NodeRef, indexlist []int, weight int,
) constraint {
    return constraint{
        TtSourceTag: string(id),
        Ctype:       ctype,
        Parameters:  indexlist,
        Weight:      weight}
}

//TODO: Possibility of multiple FET constraints for one DB constraint. That could
// be quite a radical change ...

func (fetbuild *fet_build) add_time_constraint(e *etree.Element, c constraint) {
    i := len(fetbuild.ConstraintElements)
    fetbuild.ConstraintElements = append(fetbuild.ConstraintElements, e)
    fetbuild.TimeConstraints = append(fetbuild.TimeConstraints, i)

    // Make a tag for the constraint
    fetbuild.constraint_counter++
    if c.Weight == 100 {
        c.Backend = fmt.Sprintf("[%d]", fetbuild.constraint_counter)
    } else {
        wfet := e.SelectElement("Weight_Percentage").Text()
        c.Backend = fmt.Sprintf("[%d:%s]", fetbuild.constraint_counter, wfet)
        if !fetbuild.real_soft {
            e.SelectElement("Weight_Percentage").SetText("100")
        }
    }
    e.CreateElement("Comments").SetText(c.Backend)

    fetbuild.Constraints = append(fetbuild.Constraints, c)
}

func (fetbuild *fet_build) add_space_constraint(e *etree.Element, c constraint) {
    i := len(fetbuild.ConstraintElements)
    fetbuild.ConstraintElements = append(fetbuild.ConstraintElements, e)
    fetbuild.SpaceConstraints = append(fetbuild.SpaceConstraints, i)

    // Make a tag for the constraint
    fetbuild.constraint_counter++
    if c.Weight == 100 {
        c.Backend = fmt.Sprintf("[%d]", fetbuild.constraint_counter)
    } else {
        wfet := e.SelectElement("Weight_Percentage").Text()
        c.Backend = fmt.Sprintf("[%d:%s]", fetbuild.constraint_counter, wfet)
        if !fetbuild.real_soft {
            e.SelectElement("Weight_Percentage").SetText("100")
        }
    }
    e.CreateElement("Comments").SetText(c.Backend)

    fetbuild.Constraints = append(fetbuild.Constraints, c)
}

func (fetbuild *fet_build) DbWeight2Fet(w int) string {
    if w <= 0 {
        return "0"
    }
    if w >= 100 {
        return "100"
    }
    return strconv.FormatFloat(fetbuild.WeightTable[w], 'f', 3, 64)
}
