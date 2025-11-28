package makefet

import (
	"fetrunner/base"
	"strconv"
)

func (fetbuild *FetBuild) blocked_slots() map[NodeRef][]base.TimeSlot {
	tt_data := fetbuild.ttdata
	db0 := tt_data.BaseData.Db
	rundata := fetbuild.rundata
	sclist := fetbuild.space_constraints_list
	tclist := fetbuild.time_constraints_list
	namap := map[NodeRef][]base.TimeSlot{} // needed for lunch-break constraints

	// Rooms
	for _, c0 := range db0.Constraints[base.C_RoomNotAvailable] {
		// The weight is presumably 100% ...
		w := rundata.FetWeight(c0.Weight)
		data := c0.Data.(base.ResourceNotAvailable)
		rref := data.Resource
		// `NotAvailable` is an ordered list of time-slots in which the
		// room is to be regarded as not available for the timetable.

		if len(data.NotAvailable) != 0 {
			cna := sclist.CreateElement("ConstraintRoomNotAvailableTimes")
			cna.CreateElement("Weight_Percentage").SetText(w)
			cna.CreateElement("Room").SetText(db0.Ref2Tag(rref))
			cna.CreateElement("Number_of_Not_Available_Times").
				SetText(strconv.Itoa(len(data.NotAvailable)))
			for _, slot := range data.NotAvailable {
				nat := cna.CreateElement("Not_Available_Time")
				nat.CreateElement("Day").SetText(rundata.DayIds[slot.Day].Backend)
				nat.CreateElement("Hour").SetText(rundata.HourIds[slot.Hour].Backend)
			}
			cna.CreateElement("Active").SetText("true")

			fetbuild.add_space_constraint(cna, param_constraint(
				c0.CType, c0.Id, tt_data.RoomIndex[rref], c0.Weight))
		}
	}

	// Teachers
	for _, c0 := range db0.Constraints[base.C_TeacherNotAvailable] {
		// The weight is presumably 100% ...
		w := rundata.FetWeight(c0.Weight)
		data := c0.Data.(base.ResourceNotAvailable)
		tref := data.Resource
		namap[tref] = data.NotAvailable
		// `NotAvailable` is an ordered list of time-slots in which the
		// teacher is to be regarded as not available for the timetable.
		if len(data.NotAvailable) != 0 {
			cna := tclist.CreateElement("ConstraintTeacherNotAvailableTimes")
			cna.CreateElement("Weight_Percentage").SetText(w)
			cna.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			cna.CreateElement("Number_of_Not_Available_Times").
				SetText(strconv.Itoa(len(data.NotAvailable)))
			for _, slot := range data.NotAvailable {
				nat := cna.CreateElement("Not_Available_Time")
				nat.CreateElement("Day").SetText(rundata.DayIds[slot.Day].Backend)
				nat.CreateElement("Hour").SetText(rundata.HourIds[slot.Hour].Backend)
			}
			cna.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(cna, param_constraint(
				c0.CType, c0.Id, tt_data.TeacherIndex[tref], c0.Weight))
		}
	}

	// Classes
	for _, c0 := range db0.Constraints[base.C_ClassNotAvailable] {
		// The weight is presumably 100% ...
		w := rundata.FetWeight(c0.Weight)
		data := c0.Data.(base.ResourceNotAvailable)
		cref := data.Resource
		namap[cref] = data.NotAvailable
		// `NotAvailable` is an ordered list of time-slots in which the
		// class is to be regarded as not available for the timetable.
		if len(data.NotAvailable) != 0 {
			cna := tclist.CreateElement("ConstraintStudentsSetNotAvailableTimes")
			cna.CreateElement("Weight_Percentage").SetText(w)
			cna.CreateElement("Students").SetText(db0.Ref2Tag(cref))
			cna.CreateElement("Number_of_Not_Available_Times").
				SetText(strconv.Itoa(len(data.NotAvailable)))
			for _, slot := range data.NotAvailable {
				nat := cna.CreateElement("Not_Available_Time")
				nat.CreateElement("Day").SetText(rundata.DayIds[slot.Day].Backend)
				nat.CreateElement("Hour").SetText(rundata.HourIds[slot.Hour].Backend)
			}
			cna.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(cna, param_constraint(
				c0.CType, c0.Id, tt_data.ClassIndex[cref], c0.Weight))
		}
	}

	return namap
}
