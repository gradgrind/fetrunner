package makefet

import (
	"fetrunner/db"
	"fmt"
	"strconv"

	"github.com/beevik/etree"
)

func FetTree(db *db.DbTopLevel) *etree.Document {
	institution := "The School"

	doc := etree.NewDocument()
	doc.CreateProcInst("xml", `version="1.0" encoding="UTF-8"`)

	fetroot := doc.CreateElement("fet")
	fetroot.CreateAttr("version", "6.28.2")
	fetroot.CreateElement("Institution_Name").SetText(institution)

	//TODO?
	//fetroot.CreateElement("Comments").SetText(source_ref)

	set_days_hours(fetroot, db)
	set_teachers(fetroot, db)
	set_subjects(fetroot, db)
	set_rooms(fetroot, db)

	return doc
}

func set_days_hours(fetroot *etree.Element, db *db.DbTopLevel) {
	fetdays := fetroot.CreateElement("Days_List")
	fetdays.CreateElement("Number_of_Days").SetText(strconv.Itoa(len(db.Days)))
	for _, n := range db.Days {
		fetday := fetdays.CreateElement("Day")
		fetday.CreateElement("Name").SetText(n.GetTag())
		fetday.CreateElement("Long_Name").SetText(string(n.GetRef()))
	}

	fethours := fetroot.CreateElement("Hours_List")
	fethours.CreateElement("Number_of_Hours").SetText(strconv.Itoa(len(db.Hours)))
	for _, n := range db.Hours {
		fethour := fethours.CreateElement("Hour")
		fethour.CreateElement("Name").SetText(n.GetTag())
		fethour.CreateElement("Long_Name").SetText(string(n.GetRef()))
	}
}

func set_teachers(fetroot *etree.Element, db *db.DbTopLevel) {
	fetdays := fetroot.CreateElement("Teachers_List")
	for _, n := range db.Teachers {
		fetday := fetdays.CreateElement("Teacher")
		fetday.CreateElement("Name").SetText(n.GetTag())
		fetday.CreateElement("Long_Name").SetText(
			fmt.Sprintf("%s %s", n.Firstname, n.Name))
		fetday.CreateElement("Comments").SetText(string(n.GetRef()))
	}
}

func set_subjects(fetroot *etree.Element, db *db.DbTopLevel) {
	fetdays := fetroot.CreateElement("Subjects_List")
	for _, n := range db.Subjects {
		fetday := fetdays.CreateElement("Subject")
		fetday.CreateElement("Name").SetText(n.GetTag())
		fetday.CreateElement("Long_Name").SetText(n.Name)
		fetday.CreateElement("Comments").SetText(string(n.GetRef()))
	}
}

func set_rooms(fetroot *etree.Element, db *db.DbTopLevel) {
	fetdays := fetroot.CreateElement("Rooms_List")
	for _, n := range db.Rooms {
		fetday := fetdays.CreateElement("Room")
		fetday.CreateElement("Name").SetText(n.GetTag())
		fetday.CreateElement("Long_Name").SetText(n.Name)
		fetday.CreateElement("Capacity").SetText("30000")
		fetday.CreateElement("Virtual").SetText("false")
		fetday.CreateElement("Comments").SetText(string(n.GetRef()))
	}
}
