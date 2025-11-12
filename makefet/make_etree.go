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

type FetData struct {
	room_list              *etree.Element // needed for adding virtual rooms
	activity_tag_list      *etree.Element // in case these are needed
	time_constraints_list  *etree.Element
	space_constraints_list *etree.Element

	// Cache for FET virtual rooms, "hash" -> FET-virtual-room tag
	fet_virtual_rooms  map[string]string
	fet_virtual_room_n map[string]int // FET-virtual-room tag -> number of room sets
}

func FetTree(tt_data *timetable.TtData) *etree.Document {
	fetdata := &FetData{
		fet_virtual_rooms:  map[string]string{},
		fet_virtual_room_n: map[string]int{},
	}
	tt_data.BackendData = fetdata
	db0 := tt_data.Db
	institution := "The School"

	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

	fetroot := doc.CreateElement("fet")
	fetroot.CreateAttr("version", "6.28.2")
	fetroot.CreateElement("Mode").SetText("Official")
	fetroot.CreateElement("Institution_Name").SetText(institution)

	//TODO?
	source_ref := ""
	fetroot.CreateElement("Comments").SetText(source_ref)

	set_days_hours(fetroot, db0)
	set_teachers(fetroot, db0)
	set_subjects(fetroot, db0)
	set_rooms(fetroot, tt_data)
	set_classes(fetroot, tt_data)

	fetdata.activity_tag_list = fetroot.CreateElement("Activity_Tags_List")

	set_activities(fetroot, tt_data)

	tclist := fetroot.CreateElement("Time_Constraints_List")
	fetdata.time_constraints_list = tclist
	bctime := tclist.CreateElement("ConstraintBasicCompulsoryTime")
	bctime.CreateElement("Weight_Percentage").SetText("100")
	bctime.CreateElement("Active").SetText("true")

	sclist := fetroot.CreateElement("Space_Constraints_List")
	fetdata.space_constraints_list = sclist
	bcspace := sclist.CreateElement("ConstraintBasicCompulsorySpace")
	bcspace.CreateElement("Weight_Percentage").SetText("100")
	bcspace.CreateElement("Active").SetText("true")

	// Add "NotAvailable" constraints for all resources, returning a map
	// linking a resource to its blocked slot list:
	//   db.NodeRef -> []db.TimeSlot
	namap := blocked_slots(tt_data)

	//TODO: Handle WITHOUT_ROOM_CONSTRAINTS
	add_placement_constraints(tt_data, false)

	add_activity_constraints(tt_data)

	add_class_constraints(tt_data, namap)
	add_teacher_constraints(tt_data, namap)

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

func add_activity_tag(tt_data *timetable.TtData, tag string) {
	atag := tt_data.BackendData.(*FetData).
		activity_tag_list.CreateElement("Activity_Tag")
	atag.CreateElement("Name").SetText(tag)
	atag.CreateElement("Printable").SetText("false")
}

func set_days_hours(fetroot *etree.Element, db0 *db.DbTopLevel) {
	fetdays := fetroot.CreateElement("Days_List")
	fetdays.CreateElement("Number_of_Days").SetText(strconv.Itoa(len(db0.Days)))
	for _, n := range db0.Days {
		fetday := fetdays.CreateElement("Day")
		fetday.CreateElement("Name").SetText(n.GetTag())
		fetday.CreateElement("Long_Name").SetText(string(n.GetRef()))
	}

	fethours := fetroot.CreateElement("Hours_List")
	fethours.CreateElement("Number_of_Hours").SetText(strconv.Itoa(len(db0.Hours)))
	for _, n := range db0.Hours {
		fethour := fethours.CreateElement("Hour")
		fethour.CreateElement("Name").SetText(n.GetTag())
		fethour.CreateElement("Long_Name").SetText(string(n.GetRef()))
	}
}

func set_teachers(fetroot *etree.Element, db0 *db.DbTopLevel) {
	fetteachers := fetroot.CreateElement("Teachers_List")
	for _, n := range db0.Teachers {
		fetteacher := fetteachers.CreateElement("Teacher")
		fetteacher.CreateElement("Name").SetText(n.GetTag())
		fetteacher.CreateElement("Long_Name").SetText(
			fmt.Sprintf("%s %s", n.Firstname, n.Name))
		fetteacher.CreateElement("Comments").SetText(string(n.GetRef()))
	}
}

func set_subjects(fetroot *etree.Element, db0 *db.DbTopLevel) {
	fetsubjects := fetroot.CreateElement("Subjects_List")
	for _, n := range db0.Subjects {
		fetsubject := fetsubjects.CreateElement("Subject")
		fetsubject.CreateElement("Name").SetText(n.GetTag())
		fetsubject.CreateElement("Long_Name").SetText(n.Name)
		fetsubject.CreateElement("Comments").SetText(string(n.GetRef()))
	}
}

func set_rooms(fetroot *etree.Element, tt_data *timetable.TtData) {
	fetrooms := fetroot.CreateElement("Rooms_List")
	tt_data.BackendData.(*FetData).room_list = fetrooms
	for _, n := range tt_data.Db.Rooms {
		fetroom := fetrooms.CreateElement("Room")
		fetroom.CreateElement("Name").SetText(n.GetTag())
		fetroom.CreateElement("Long_Name").SetText(n.Name)
		fetroom.CreateElement("Capacity").SetText("30000")
		fetroom.CreateElement("Virtual").SetText("false")
		fetroom.CreateElement("Comments").SetText(string(n.GetRef()))
	}
}

func blocked_slots(tt_data *timetable.TtData) map[db.NodeRef][]db.TimeSlot {
	db0 := tt_data.Db
	sclist := tt_data.BackendData.(*FetData).space_constraints_list
	tclist := tt_data.BackendData.(*FetData).time_constraints_list
	namap := map[db.NodeRef][]db.TimeSlot{} // needed for lunch-break constraints

	// Rooms
	for _, c0 := range db0.Constraints[db.C_RoomNotAvailable] {
		// The weight is assumed to be 100%.
		data := c0.Data.(db.ResourceNotAvailable)
		rref := data.Resource
		// `NotAvailable` is an ordered list of time-slots in which the
		// room is to be regarded as not available for the timetable.

		if len(data.NotAvailable) != 0 {
			cna := sclist.CreateElement("ConstraintRoomNotAvailableTimes")
			cna.CreateElement("Weight_Percentage").SetText("100")
			cna.CreateElement("Room").SetText(db0.Ref2Tag(rref))
			cna.CreateElement("Number_of_Not_Available_Times").
				SetText(strconv.Itoa(len(data.NotAvailable)))
			for _, slot := range data.NotAvailable {
				nat := cna.CreateElement("Not_Available_Time")
				nat.CreateElement("Day").SetText(db0.Days[slot.Day].GetTag())
				nat.CreateElement("Hour").SetText(db0.Hours[slot.Hour].GetTag())
			}
			cna.CreateElement("Active").SetText("true")
			cna.CreateElement("Comments").SetText(resource_constraint(
				db.C_RoomNotAvailable, c0.Id, rref))
		}
	}

	// Teachers
	for _, c0 := range db0.Constraints[db.C_TeacherNotAvailable] {
		// The weight is assumed to be 100%.
		data := c0.Data.(db.ResourceNotAvailable)
		tref := data.Resource
		namap[tref] = data.NotAvailable
		// `NotAvailable` is an ordered list of time-slots in which the
		// teacher is to be regarded as not available for the timetable.
		if len(data.NotAvailable) != 0 {
			cna := tclist.CreateElement("ConstraintTeacherNotAvailableTimes")
			cna.CreateElement("Weight_Percentage").SetText("100")
			cna.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			cna.CreateElement("Number_of_Not_Available_Times").
				SetText(strconv.Itoa(len(data.NotAvailable)))
			for _, slot := range data.NotAvailable {
				nat := cna.CreateElement("Not_Available_Time")
				nat.CreateElement("Day").SetText(db0.Days[slot.Day].GetTag())
				nat.CreateElement("Hour").SetText(db0.Hours[slot.Hour].GetTag())
			}
			cna.CreateElement("Active").SetText("true")
			cna.CreateElement("Comments").SetText(resource_constraint(
				db.C_TeacherNotAvailable, c0.Id, tref))
		}
	}

	// Classes
	for _, c0 := range db0.Constraints[db.C_ClassNotAvailable] {
		// The weight is assumed to be 100%.
		data := c0.Data.(db.ResourceNotAvailable)
		cref := data.Resource
		namap[cref] = data.NotAvailable
		// `NotAvailable` is an ordered list of time-slots in which the
		// class is to be regarded as not available for the timetable.
		if len(data.NotAvailable) != 0 {
			cna := tclist.CreateElement("ConstraintStudentsSetNotAvailableTimes")
			cna.CreateElement("Weight_Percentage").SetText("100")
			cna.CreateElement("Students").SetText(db0.Ref2Tag(cref))
			cna.CreateElement("Number_of_Not_Available_Times").
				SetText(strconv.Itoa(len(data.NotAvailable)))
			for _, slot := range data.NotAvailable {
				nat := cna.CreateElement("Not_Available_Time")
				nat.CreateElement("Day").SetText(db0.Days[slot.Day].GetTag())
				nat.CreateElement("Hour").SetText(db0.Hours[slot.Hour].GetTag())
			}
			cna.CreateElement("Active").SetText("true")
			cna.CreateElement("Comments").SetText(resource_constraint(
				db.C_ClassNotAvailable, c0.Id, cref))
		}
	}

	return namap
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
