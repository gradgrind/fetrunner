package makefet

import (
	"fetrunner/db"
	"fetrunner/timetable"
	"strconv"
)

func add_room_constraints(tt_data *timetable.TtData) {
	db0 := tt_data.Db
	sclist := tt_data.BackendData.(*FetData).space_constraints_list

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
		}
	}
}
