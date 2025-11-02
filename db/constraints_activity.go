package db

var (
	C_ActivityStartTime = "ActivityStartTime"
	C_ActivityRooms     = "ActivityRooms"
)

// ++ ActivityStartTime

type ActivityStartTime struct {
	Activity NodeRef
	Day      int
	Hour     int
	Fixed    bool
}

func (db *DbTopLevel) NewActivityStartTime(
	id NodeRef, weight int, aid NodeRef, day int, hour int, fixed bool,
) *Constraint {
	c := &Constraint{
		CType:  C_ActivityStartTime,
		Id:     id,
		Weight: weight,
		Data:   ActivityStartTime{aid, day, hour, fixed},
	}
	db.addConstraint(c)
	return c
}

// ++ ActivityRooms

type ActivityRooms struct {
	Activity    NodeRef
	Rooms       []NodeRef   // "real" rooms
	RoomChoices [][]NodeRef // also here, the nodes are "real" rooms
}

func (db *DbTopLevel) NewActivityRooms(
	id NodeRef, weight int,
	aid NodeRef, rooms []NodeRef, roomChoices [][]NodeRef,
) *Constraint {
	c := &Constraint{
		CType:  C_ActivityRooms,
		Id:     id,
		Weight: weight,
		Data:   ActivityRooms{aid, rooms, roomChoices},
	}
	db.addConstraint(c)
	return c
}
