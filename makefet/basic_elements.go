package makefet

import (
	"fetrunner/db"
	"fmt"
	"strconv"
)

func (fetbuild *FetBuild) blocked_slots() map[db.NodeRef][]db.TimeSlot {
	tt_data := fetbuild.ttdata
	db0 := tt_data.Db
	sclist := fetbuild.space_constraints_list
	tclist := fetbuild.time_constraints_list
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
			fetbuild.add_space_constraint(cna)
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
			fetbuild.add_time_constraint(cna)
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
			fetbuild.add_time_constraint(cna)
		}
	}

	return namap
}

func (fetbuild *FetBuild) set_days_hours() {
	db0 := fetbuild.ttdata.Db
	fetdays := fetbuild.fetroot.CreateElement("Days_List")
	fetdays.CreateElement("Number_of_Days").SetText(strconv.Itoa(len(db0.Days)))
	for _, n := range db0.Days {
		id := n.GetTag()
		fetday := fetdays.CreateElement("Day")
		fetday.CreateElement("Name").SetText(id)
		fetday.CreateElement("Long_Name").SetText(n.Name)

		fetbuild.rundata.DayIds = append(fetbuild.rundata.DayIds, IdPair{
			Source: string(n.GetRef()), Backend: id,
		})
	}

	fethours := fetbuild.fetroot.CreateElement("Hours_List")
	fethours.CreateElement("Number_of_Hours").SetText(strconv.Itoa(len(db0.Hours)))
	for _, n := range db0.Hours {
		id := n.GetTag()
		fethour := fethours.CreateElement("Hour")
		fethour.CreateElement("Name").SetText(id)
		fethour.CreateElement("Long_Name").SetText(n.Name)

		fetbuild.rundata.HourIds = append(fetbuild.rundata.DayIds, IdPair{
			Source: string(n.GetRef()), Backend: id,
		})
	}
}

func (fetbuild *FetBuild) set_teachers() {
	fetteachers := fetbuild.fetroot.CreateElement("Teachers_List")
	for _, n := range fetbuild.ttdata.Db.Teachers {
		id := n.GetTag()
		fetteacher := fetteachers.CreateElement("Teacher")
		fetteacher.CreateElement("Name").SetText(id)
		fetteacher.CreateElement("Long_Name").SetText(
			fmt.Sprintf("%s %s", n.Firstname, n.Name))
		fetteacher.CreateElement("Comments").SetText(string(n.GetRef()))

		fetbuild.rundata.TeacherIds = append(fetbuild.rundata.DayIds, IdPair{
			Source: string(n.GetRef()), Backend: id,
		})
	}
}

//TODO: make lists for all elements?

func (fetbuild *FetBuild) set_subjects() {
	fetsubjects := fetbuild.fetroot.CreateElement("Subjects_List")
	for _, n := range fetbuild.ttdata.Db.Subjects {
		fetsubject := fetsubjects.CreateElement("Subject")
		fetsubject.CreateElement("Name").SetText(n.GetTag())
		fetsubject.CreateElement("Long_Name").SetText(n.Name)
		fetsubject.CreateElement("Comments").SetText(string(n.GetRef()))
	}
}

func (fetbuild *FetBuild) set_rooms() {
	fetrooms := fetbuild.fetroot.CreateElement("Rooms_List")
	fetbuild.room_list = fetrooms
	for _, n := range fetbuild.ttdata.Db.Rooms {
		fetroom := fetrooms.CreateElement("Room")
		fetroom.CreateElement("Name").SetText(n.GetTag())
		fetroom.CreateElement("Long_Name").SetText(n.Name)
		fetroom.CreateElement("Capacity").SetText("30000")
		fetroom.CreateElement("Virtual").SetText("false")
		fetroom.CreateElement("Comments").SetText(string(n.GetRef()))
	}
}
