package base

// ++ ActivityStartTime

type ActivityStartTime struct {
    Activity NodeRef
    Day      int
    Hour     int
}

func (db *DbTopLevel) NewActivityStartTime(
    id NodeRef, weight int, aid NodeRef, day int, hour int,
) *BaseConstraint {
    c := &BaseConstraint{
        CType:  C_ActivityStartTime,
        Id:     id,
        Weight: weight,
        Data:   ActivityStartTime{aid, day, hour},
    }
    db.addConstraint(c)
    return c
}

// +++++ ActivityPlacement +++++
// This is not really a constraint, it is the result of a placement.

func (db *DbTopLevel) AddActivityPlacement(
    placementTag string, aid NodeRef, day int, hour int, rooms []NodeRef,
) {
    db.Placements[placementTag] = append(db.Placements[placementTag],
        &ActivityPlacement{
            Activity: aid,
            Day:      day,
            Hour:     hour,
            Rooms:    rooms,
        })
}
