package fet

import (
	"strconv"
)

func (fetbuild *fet_build) set_days_hours() {
	ids := []string{}
	fetdays := fetbuild.fetroot.CreateElement("Days_List")
	dlist := fetbuild.ttsource.GetDays()
	fetdays.CreateElement("Number_of_Days").SetText(strconv.Itoa(len(dlist)))
	for _, day := range dlist {
		id := day.Tag
		fetday := fetdays.CreateElement("Day")
		fetday.CreateElement("Name").SetText(id)
		ids = append(ids, id)
	}
	fetbuild.DayList = ids

	ids = []string{}
	fethours := fetbuild.fetroot.CreateElement("Hours_List")
	hlist := fetbuild.ttsource.GetHours()
	fethours.CreateElement("Number_of_Hours").SetText(strconv.Itoa(len(hlist)))
	for _, hour := range hlist {
		id := hour.Tag
		fethour := fethours.CreateElement("Hour")
		fethour.CreateElement("Name").SetText(id)
		ids = append(ids, id)
	}
	fetbuild.HourList = ids
}

func (fetbuild *fet_build) set_teachers() {
	ids := []string{}
	fetteachers := fetbuild.fetroot.CreateElement("Teachers_List")
	for _, t := range fetbuild.ttsource.GetTeachers() {
		id := t.Tag
		fetteacher := fetteachers.CreateElement("Teacher")
		fetteacher.CreateElement("Name").SetText(id)
		ids = append(ids, id)
	}
	fetbuild.TeacherList = ids
}

func (fetbuild *fet_build) set_subjects() {
	ids := []string{}
	fetsubjects := fetbuild.fetroot.CreateElement("Subjects_List")
	for _, s := range fetbuild.ttsource.GetSubjects() {
		id := s.Tag
		fetsubject := fetsubjects.CreateElement("Subject")
		fetsubject.CreateElement("Name").SetText(id)
		ids = append(ids, id)
	}
	fetbuild.SubjectList = ids
}

func (fetbuild *fet_build) set_rooms() {
	ids := []string{}
	fetrooms := fetbuild.fetroot.CreateElement("Rooms_List")
	fetbuild.room_list = fetrooms
	for _, r := range fetbuild.ttsource.GetRooms() {
		id := r.Tag
		fetroom := fetrooms.CreateElement("Room")
		fetroom.CreateElement("Name").SetText(id)
		fetroom.CreateElement("Capacity").SetText("30000")
		fetroom.CreateElement("Virtual").SetText("false")
		ids = append(ids, id)
	}
	fetbuild.RoomList = ids
}
