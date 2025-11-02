package db

var (
	C_RoomNotAvailable = "RoomNotAvailable"
)

// ++ RoomNotAvailable

// TimeSlots in which the room is not available.
func (db *DbTopLevel) NewRoomNotAvailable(
	cid NodeRef, weight int, rid NodeRef, notAvailable []TimeSlot,
) *Constraint {
	c := &Constraint{
		CType:  C_RoomNotAvailable,
		Id:     cid,
		Weight: weight,
		Data:   ResourceNotAvailable{rid, notAvailable},
	}
	db.addConstraint(c)
	return c
}
