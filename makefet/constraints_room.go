package fet

import (
	"fetrunner/db"
)

func (fetinfo *fetInfo) handle_room_constraints() {
	tt_data := fetinfo.tt_data
	db0 := tt_data.Db

	natimes := []roomNotAvailable{}
	for _, rna := range db0.Constraints[db.C_RoomNotAvailable] {
		// The weight is assumed to be 100%.
		data := rna.Data.(db.ResourceNotAvailable)
		// `NotAvailable` is an ordered list of time-slots in which the
		// teacher is to be regarded as not available for the timetable.
		nats := []notAvailableTime{}
		for _, slot := range data.NotAvailable {
			nats = append(nats,
				notAvailableTime{
					Day:  db0.Days[slot.Day].GetTag(),
					Hour: db0.Hours[slot.Hour].GetTag()})
		}
		if len(nats) > 0 {
			natimes = append(natimes,
				roomNotAvailable{
					Weight_Percentage:             100,
					Room:                          db0.Ref2Tag(data.Resource),
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
					Comments:                      string(rna.Id),
				})
		}
	}
	fetinfo.fetdata.Space_Constraints_List.
		ConstraintRoomNotAvailableTimes = natimes
}
