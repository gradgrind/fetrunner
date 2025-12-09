package makefet

import (
	"fmt"
	"strconv"
)

func (fetbuild *FetBuild) set_days_hours() {
	db := fetbuild.basedata.Db
	ids := []IdPair{}
	fetdays := fetbuild.fetroot.CreateElement("Days_List")
	fetdays.CreateElement("Number_of_Days").SetText(strconv.Itoa(len(db.Days)))
	for _, n := range db.Days {
		id := n.GetTag()
		fetday := fetdays.CreateElement("Day")
		fetday.CreateElement("Name").SetText(id)
		fetday.CreateElement("Long_Name").SetText(n.Name)

		ids = append(ids, IdPair{Source: string(n.GetRef()), Backend: id})
	}
	fetbuild.rundata.DayIds = ids

	ids = []IdPair{}
	fethours := fetbuild.fetroot.CreateElement("Hours_List")
	fethours.CreateElement("Number_of_Hours").SetText(strconv.Itoa(len(db.Hours)))
	for _, n := range db.Hours {
		id := n.GetTag()
		fethour := fethours.CreateElement("Hour")
		fethour.CreateElement("Name").SetText(id)
		fethour.CreateElement("Long_Name").SetText(n.Name)

		ids = append(ids, IdPair{Source: string(n.GetRef()), Backend: id})
	}
	fetbuild.rundata.HourIds = ids
}

func (fetbuild *FetBuild) set_teachers() {
	db := fetbuild.basedata.Db
	ids := []IdPair{}
	fetteachers := fetbuild.fetroot.CreateElement("Teachers_List")
	for _, n := range db.Teachers {
		id := n.GetTag()
		fetteacher := fetteachers.CreateElement("Teacher")
		fetteacher.CreateElement("Name").SetText(id)
		fetteacher.CreateElement("Long_Name").SetText(
			fmt.Sprintf("%s %s", n.Firstname, n.Name))

		ids = append(ids, IdPair{Source: string(n.GetRef()), Backend: id})
	}
	fetbuild.rundata.TeacherIds = ids
}

func (fetbuild *FetBuild) set_subjects() {
	db := fetbuild.basedata.Db
	ids := []IdPair{}
	fetsubjects := fetbuild.fetroot.CreateElement("Subjects_List")
	for _, n := range db.Subjects {
		id := n.GetTag()
		fetsubject := fetsubjects.CreateElement("Subject")
		fetsubject.CreateElement("Name").SetText(id)
		fetsubject.CreateElement("Long_Name").SetText(n.Name)

		ids = append(ids, IdPair{Source: string(n.GetRef()), Backend: id})
	}
	fetbuild.rundata.SubjectIds = ids
}

func (fetbuild *FetBuild) set_rooms() {
	db := fetbuild.basedata.Db
	ids := []IdPair{}
	fetrooms := fetbuild.fetroot.CreateElement("Rooms_List")
	fetbuild.room_list = fetrooms
	for _, n := range db.Rooms {
		id := n.GetTag()
		fetroom := fetrooms.CreateElement("Room")
		fetroom.CreateElement("Name").SetText(id)
		fetroom.CreateElement("Long_Name").SetText(n.Name)
		fetroom.CreateElement("Capacity").SetText("30000")
		fetroom.CreateElement("Virtual").SetText("false")

		ids = append(ids, IdPair{Source: string(n.GetRef()), Backend: id})
	}
	fetbuild.rundata.RoomIds = ids
}
