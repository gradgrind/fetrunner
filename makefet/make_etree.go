package makefet

import (
	"fetrunner/db"
	"fetrunner/timetable"
	"fmt"
	"strconv"

	"github.com/beevik/etree"
)

type NodeRef = db.NodeRef

const CLASS_GROUP_SEP = "."
const VIRTUAL_ROOM_PREFIX = "!"

// const LUNCH_BREAK_TAG = "-lb-"
// const LUNCH_BREAK_NAME = "Lunch Break"

type FetData struct {
	room_list         *etree.Element // needed for adding virtual rooms
	activity_tag_list *etree.Element // in case these are needed
}

func FetTree(tt_data *timetable.TtData) *etree.Document {
	fetdata := &FetData{}
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

	return doc
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
