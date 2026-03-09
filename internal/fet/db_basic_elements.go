package fet

import (
	"fmt"
	"strconv"
)

func (fetbuild *fet_build) set_days_hours() {
	db := fetbuild.basedata.Db
	ids := []string{}
	fetdays := fetbuild.fetroot.CreateElement("Days_List")
	fetdays.CreateElement("Number_of_Days").SetText(strconv.Itoa(len(db.Days)))
	for _, n := range db.Days {
		id := n.GetTag()
		fetday := fetdays.CreateElement("Day")
		fetday.CreateElement("Name").SetText(id)
		fetday.CreateElement("Long_Name").SetText(n.Name)

		ids = append(ids, id)
	}
	fetbuild.DayList = ids

	ids = []string{}
	fethours := fetbuild.fetroot.CreateElement("Hours_List")
	fethours.CreateElement("Number_of_Hours").SetText(strconv.Itoa(len(db.Hours)))
	for _, n := range db.Hours {
		id := n.GetTag()
		fethour := fethours.CreateElement("Hour")
		fethour.CreateElement("Name").SetText(id)
		fethour.CreateElement("Long_Name").SetText(n.Name)

		ids = append(ids, id)
	}
	fetbuild.HourList = ids
}

func (fetbuild *fet_build) set_teachers() {
	db := fetbuild.basedata.Db
	ids := []string{}
	fetteachers := fetbuild.fetroot.CreateElement("Teachers_List")
	for _, n := range db.Teachers {
		id := n.GetTag()
		fetteacher := fetteachers.CreateElement("Teacher")
		fetteacher.CreateElement("Name").SetText(id)
		fetteacher.CreateElement("Long_Name").SetText(
			fmt.Sprintf("%s %s", n.Firstname, n.Name))

		ids = append(ids, id)
	}
	fetbuild.TeacherList = ids
}

func (fetbuild *fet_build) set_subjects() {
	db := fetbuild.basedata.Db
	ids := []string{}
	fetsubjects := fetbuild.fetroot.CreateElement("Subjects_List")
	for _, n := range db.Subjects {
		id := n.GetTag()
		fetsubject := fetsubjects.CreateElement("Subject")
		fetsubject.CreateElement("Name").SetText(id)
		fetsubject.CreateElement("Long_Name").SetText(n.Name)

		ids = append(ids, id)
	}
	fetbuild.SubjectList = ids
}

func (fetbuild *fet_build) set_rooms() {
	db := fetbuild.basedata.Db
	ids := []string{}
	fetrooms := fetbuild.fetroot.CreateElement("Rooms_List")
	fetbuild.room_list = fetrooms
	for _, n := range db.Rooms {
		id := n.GetTag()
		fetroom := fetrooms.CreateElement("Room")
		fetroom.CreateElement("Name").SetText(id)
		fetroom.CreateElement("Long_Name").SetText(n.Name)
		fetroom.CreateElement("Capacity").SetText("30000")
		fetroom.CreateElement("Virtual").SetText("false")

		ids = append(ids, id)
	}
	fetbuild.RoomList = ids
}
