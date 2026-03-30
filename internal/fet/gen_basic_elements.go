package fet

import (
	"fetrunner/internal/base"
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
	fetbuild.teacher_hard_blocked = make([][][]bool, len(ids))
	fetbuild.teacher_max_days = make([]int, len(ids))
	fetbuild.teacher_max_afternoons = make([]int, len(ids))
	fetbuild.teacher_lunch_break_days = make([][]int, len(ids))
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

func (fetbuild *fet_build) set_classes() {
	source := fetbuild.ttsource
	ids := []string{}
	aglist := source.GetAtomicGroups() // needed for Subgroups
	fetyears := fetbuild.fetroot.CreateElement("Students_List")
	for _, cl := range source.GetClasses() {
		cname := cl.Tag
		// Skip "special" classes.
		if cname == "" {
			continue
		}
		ids = append(ids, cname)
		fetyear := fetyears.CreateElement("Year")
		fetyear.CreateElement("Name").SetText(cname)

		// Construct the Groups and Subgroups
		for _, g := range cl.Groups {
			// Need to construct group name with class, CLASS_GROUP_SEP
			// and group
			fetgroup := fetyear.CreateElement("Group")
			fetgroup.CreateElement("Name").SetText(cl.Tag + CLASS_GROUP_SEP + g.Tag)
			for _, agix := range g.AtomicIndexes {
				fetsubgroup := fetgroup.CreateElement("Subgroup")
				fetsubgroup.CreateElement("Name").SetText(aglist[agix])
			}
		}
	}
	fetbuild.ClassList = ids
	fetbuild.class_hard_blocked = make([][][]bool, len(ids))
	fetbuild.class_max_afternoons = make([]int, len(ids))
	fetbuild.class_lunch_break_days = make([][]int, len(ids))
}

// In FET the group identifier is constructed from the class tag,
// CLASS_GROUP_SEP and the group tag. However, if the group is the
// whole class, just the class tag is used.
func fetGroupTag(g *base.Group) string {
	gt := g.Class.Tag
	if g.Tag != "" {
		gt += CLASS_GROUP_SEP + g.Tag
	}
	return gt
}
