package db

// ++ ClassNotAvailable

// TimeSlots in which the class is not available.
func (db *DbTopLevel) NewClassNotAvailable(
	id NodeRef, weight int, tid NodeRef, notAvailable []TimeSlot,
) *Constraint {
	c := &Constraint{
		CType:  "ClassNotAvailable",
		Id:     id,
		Weight: weight,
		Data:   ResourceNotAvailable{tid, notAvailable},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ ClassMinActivitiesPerDay

// Double time-slots count as two activities, etc.
func (db *DbTopLevel) NewClassMinActivitiesPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  "ClassMinActivitiesPerDay",
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ ClassMaxActivitiesPerDay

// Double time-slots count as two activities, etc.
func (db *DbTopLevel) NewClassMaxActivitiesPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  "ClassMaxActivitiesPerDay",
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ ClassMaxGapsPerDay

func (db *DbTopLevel) NewClassMaxGapsPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  "ClassMaxGapsPerDay",
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ ClassMaxGapsPerWeek

func (db *DbTopLevel) NewClassMaxGapsPerWeek(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  "ClassMaxGapsPerWeek",
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ ClassMaxAfternoons

func (db *DbTopLevel) NewClassMaxAfternoons(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  "ClassMaxAfternoons",
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ ClassLunchBreak

func (db *DbTopLevel) NewClassLunchBreak(
	id NodeRef, weight int, tid NodeRef,
) *Constraint {
	c := &Constraint{
		CType:  "ClassLunchBreak",
		Id:     id,
		Weight: weight,
		Data:   tid,
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ ClassForceFirstHour

func (db *DbTopLevel) NewClassForceFirstHour(
	id NodeRef, weight int, tid NodeRef,
) *Constraint {
	c := &Constraint{
		CType:  "ClassForceFirstHour",
		Id:     id,
		Weight: weight,
		Data:   tid,
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}
