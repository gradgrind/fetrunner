package db

// ++ RoomNotAvailable

// TimeSlots in which the room is not available.
func (db *DbTopLevel) NewRoomNotAvailable(
	id NodeRef, weight int, tid NodeRef, notAvailable []TimeSlot,
) *Constraint {
	c := &Constraint{
		CType:  "RoomNotAvailable",
		Id:     id,
		Weight: weight,
		Data:   ResourceNotAvailable{tid, notAvailable},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}
