package makefet

import (
	"fetrunner/db"
	"strconv"
)

//TODO: add to Constraints list

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
