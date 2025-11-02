package db

// ++ TeacherNotAvailable

// TimeSlots in which the teacher is not available.
func (db *DbTopLevel) NewTeacherNotAvailable(
	id NodeRef, weight int, tid NodeRef, notAvailable []TimeSlot,
) *Constraint {
	c := &Constraint{
		CType:  "TeacherNotAvailable",
		Id:     id,
		Weight: weight,
		Data:   ResourceNotAvailable{tid, notAvailable},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ TeacherMinActivitiesPerDay

// Double time-slots count as two activities, etc.
func (db *DbTopLevel) NewTeacherMinActivitiesPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  "TeacherMinActivitiesPerDay",
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ TeacherMaxActivitiesPerDay

// Double time-slots count as two activities, etc.
func (db *DbTopLevel) NewTeacherMaxActivitiesPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  "TeacherMaxActivitiesPerDay",
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ TeacherMaxDays

// Double time-slots count as two activities, etc.
func (db *DbTopLevel) NewTeacherMaxDays(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  "TeacherMaxDays",
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ TeacherMaxGapsPerDay

func (db *DbTopLevel) NewTeacherMaxGapsPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  "TeacherMaxGapsPerDay",
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ TeacherMaxGapsPerWeek

func (db *DbTopLevel) NewTeacherMaxGapsPerWeek(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  "TeacherMaxGapsPerWeek",
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ TeacherMaxAfternoons

func (db *DbTopLevel) NewTeacherMaxAfternoons(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  "TeacherMaxAfternoons",
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}

// ++ TeacherLunchBreak

func (db *DbTopLevel) NewTeacherLunchBreak(
	id NodeRef, weight int, tid NodeRef,
) *Constraint {
	c := &Constraint{
		CType:  "TeacherLunchBreak",
		Id:     id,
		Weight: weight,
		Data:   tid,
	}
	r := db.GetElement(tid).(Resource)
	r.addConstraint(c)
	return c
}
