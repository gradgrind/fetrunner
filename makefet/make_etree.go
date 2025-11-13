package makefet

import (
	"fetrunner/db"
	"fetrunner/timetable"
	"fmt"
	"math"
	"strconv"
	"strings"

	"github.com/beevik/etree"
)

type NodeRef = db.NodeRef

const CLASS_GROUP_SEP = "."
const VIRTUAL_ROOM_PREFIX = "!"

// const LUNCH_BREAK_TAG = "-lb-"
// const LUNCH_BREAK_NAME = "Lunch Break"

func FetTree(tt_data *timetable.TtData) *etree.Document {
	fetbuild := &FetBuild{
		ttdata:             tt_data,
		fet_virtual_rooms:  map[string]string{},
		fet_virtual_room_n: map[string]int{},
	}

	//TODO
	institution := "The School"
	fet_version := "6.28.2"

	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

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

	return doc
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

func resource_constraint(
	ctype string, id db.NodeRef, resource db.NodeRef,
) string {
	return fmt.Sprintf("%s.%s:%s", ctype, id, resource)
}

func param_constraint(
	ctype string, id db.NodeRef, param string,
) string {
	return fmt.Sprintf("%s.%s:%s", ctype, id, param)
}

func activities_constraint(
	ctype string, id db.NodeRef, alist []timetable.ActivityIndex,
) string {
	ailist := []string{}
	for _, a := range alist {
		ailist = append(ailist, strconv.Itoa(int(a)))
	}
	return param_constraint(ctype, id, strings.Join(ailist, ","))
}

// TODO:
func (fetbuild *FetBuild) add_time_constraint(e *etree.Element) {
	rundata := fetbuild.rundata
	i := len(rundata.ConstraintElements)
	rundata.ConstraintElements = append(rundata.ConstraintElements, e)
	rundata.TimeConstraints = append(rundata.TimeConstraints, i)
}

// TODO:
func (fetbuild *FetBuild) add_space_constraint(e *etree.Element) {
	rundata := fetbuild.rundata
	i := len(rundata.ConstraintElements)
	rundata.ConstraintElements = append(rundata.ConstraintElements, e)
	rundata.SpaceConstraints = append(rundata.SpaceConstraints, i)
}
