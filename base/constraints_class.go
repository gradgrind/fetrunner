package base

var (
	C_ClassNotAvailable        = "ClassNotAvailable"
	C_ClassMinActivitiesPerDay = "ClassMinActivitiesPerDay"
	C_ClassMaxActivitiesPerDay = "ClassMaxActivitiesPerDay"
	C_ClassMaxGapsPerDay       = "ClassMaxGapsPerDay"
	C_ClassMaxGapsPerWeek      = "ClassMaxGapsPerWeek"
	C_ClassMaxAfternoons       = "ClassMaxAfternoons"
	C_ClassLunchBreak          = "ClassLunchBreak"
	C_ClassForceFirstHour      = "ClassForceFirstHour"
)

// ++ ClassNotAvailable

// TimeSlots in which the class is not available.
func (db *DbTopLevel) NewClassNotAvailable(
	id NodeRef, weight int, tid NodeRef, notAvailable []TimeSlot,
) *Constraint {
	c := &Constraint{
		CType:  C_ClassNotAvailable,
		Id:     id,
		Weight: weight,
		Data:   ResourceNotAvailable{tid, notAvailable},
	}
	db.addConstraint(c)
	return c
}

// ++ ClassMinActivitiesPerDay

// Double time-slots count as two activities, etc.
func (db *DbTopLevel) NewClassMinActivitiesPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  C_ClassMinActivitiesPerDay,
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	db.addConstraint(c)
	return c
}

// ++ ClassMaxActivitiesPerDay

// Double time-slots count as two activities, etc.
func (db *DbTopLevel) NewClassMaxActivitiesPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  C_ClassMaxActivitiesPerDay,
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	db.addConstraint(c)
	return c
}

// ++ ClassMaxGapsPerDay

func (db *DbTopLevel) NewClassMaxGapsPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  C_ClassMaxGapsPerDay,
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	db.addConstraint(c)
	return c
}

// ++ ClassMaxGapsPerWeek

func (db *DbTopLevel) NewClassMaxGapsPerWeek(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  C_ClassMaxGapsPerWeek,
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	db.addConstraint(c)
	return c
}

// ++ ClassMaxAfternoons

func (db *DbTopLevel) NewClassMaxAfternoons(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  C_ClassMaxAfternoons,
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	db.addConstraint(c)
	return c
}

// ++ ClassLunchBreak

func (db *DbTopLevel) NewClassLunchBreak(
	id NodeRef, weight int, tid NodeRef,
) *Constraint {
	c := &Constraint{
		CType:  C_ClassLunchBreak,
		Id:     id,
		Weight: weight,
		Data:   tid,
	}
	db.addConstraint(c)
	return c
}

// ++ ClassForceFirstHour

func (db *DbTopLevel) NewClassForceFirstHour(
	id NodeRef, weight int, tid NodeRef,
) *Constraint {
	c := &Constraint{
		CType:  C_ClassForceFirstHour,
		Id:     id,
		Weight: weight,
		Data:   tid,
	}
	db.addConstraint(c)
	return c
}
