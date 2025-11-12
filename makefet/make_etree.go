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
const ATOMIC_GROUP_SEP1 = "#"
const ATOMIC_GROUP_SEP2 = "~"
const VIRTUAL_ROOM_PREFIX = "!"

// const LUNCH_BREAK_TAG = "-lb-"
// const LUNCH_BREAK_NAME = "Lunch Break"

func FetTree(tt_data *timetable.TtData) *etree.Document {
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
	set_rooms(fetroot, db0)
	set_classes(fetroot, tt_data)

	return doc
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
	fetdays := fetroot.CreateElement("Teachers_List")
	for _, n := range db0.Teachers {
		fetday := fetdays.CreateElement("Teacher")
		fetday.CreateElement("Name").SetText(n.GetTag())
		fetday.CreateElement("Long_Name").SetText(
			fmt.Sprintf("%s %s", n.Firstname, n.Name))
		fetday.CreateElement("Comments").SetText(string(n.GetRef()))
	}
}

func set_subjects(fetroot *etree.Element, db0 *db.DbTopLevel) {
	fetdays := fetroot.CreateElement("Subjects_List")
	for _, n := range db0.Subjects {
		fetday := fetdays.CreateElement("Subject")
		fetday.CreateElement("Name").SetText(n.GetTag())
		fetday.CreateElement("Long_Name").SetText(n.Name)
		fetday.CreateElement("Comments").SetText(string(n.GetRef()))
	}
}

func set_rooms(fetroot *etree.Element, db0 *db.DbTopLevel) {
	fetdays := fetroot.CreateElement("Rooms_List")
	for _, n := range db0.Rooms {
		fetday := fetdays.CreateElement("Room")
		fetday.CreateElement("Name").SetText(n.GetTag())
		fetday.CreateElement("Long_Name").SetText(n.Name)
		fetday.CreateElement("Capacity").SetText("30000")
		fetday.CreateElement("Virtual").SetText("false")
		fetday.CreateElement("Comments").SetText(string(n.GetRef()))
	}
}
