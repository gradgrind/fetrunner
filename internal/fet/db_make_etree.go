package fet

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"strconv"

	"github.com/beevik/etree"
)

type NodeRef = base.NodeRef

const CLASS_GROUP_SEP = "."
const VIRTUAL_ROOM_PREFIX = "!"

// const LUNCH_BREAK_TAG = "-lb-"
// const LUNCH_BREAK_NAME = "Lunch Break"

// TODO ...
// Construct a `fet_build` from the timetable data available via the
// `autotimetable.TtSource` interface.
// This `fet_build` needs to contain all the information for generating a
// `FET` file with a subset of constraints, as an implementation of the
// `autotimetable.TtBackend` interface. The constraints are
// determined by the source and the possibility to map to a variable number
// of `FET` constraints should be supported.
// TODO? Some fields of the `autotimetable.BasicData` are initialized.
func FetTree(attdata *autotimetable.AutoTtData) *fet_build {
	source := attdata.Source // TtSource interface
	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)
	source_constraints := source.GetConstraints()
	fetbuild := &fet_build{
		real_soft:           attdata.Parameters.REAL_SOFT,
		no_room_constraints: attdata.Parameters.WITHOUT_ROOM_CONSTRAINTS,
		ttsource:            source,

		Doc:                doc,
		WeightTable:        MakeFetWeights(),
		ConstraintElements: make([][]*etree.Element, len(source_constraints)),

		fet_virtual_rooms:  map[string]string{},
		fet_virtual_room_n: map[string]int{},
	}
	attdata.Backend = fetbuild

	fetroot := doc.CreateElement("fet")
	fetbuild.fetroot = fetroot
	fetroot.CreateAttr("version", fet_version)
	fetroot.CreateElement("Mode").SetText("Official")
	//fetroot.CreateElement("Institution_Name").SetText(source.GetInstitution())

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

	// Convert the source constraints to FET constraints
	for i, sc := range source_constraints {
		base_constraint_fet[sc.CType](fetbuild, i, sc)
	}
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

func (fetbuild *fet_build) DbWeight2Fet(w int) string {
	if w <= 0 {
		return "0"
	}
	if w >= 100 {
		return "100"
	}
	return strconv.FormatFloat(fetbuild.WeightTable[w], 'f', 3, 64)
}
