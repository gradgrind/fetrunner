package db

var (
	C_TeacherNotAvailable        = "TeacherNotAvailable"
	C_TeacherMinActivitiesPerDay = "TeacherMinActivitiesPerDay"
	C_TeacherMaxActivitiesPerDay = "TeacherMaxActivitiesPerDay"
	C_TeacherMaxDays             = "TeacherMaxDays"
	C_TeacherMaxGapsPerDay       = "TeacherMaxGapsPerDay"
	C_TeacherMaxGapsPerWeek      = "TeacherMaxGapsPerWeek"
	C_TeacherMaxAfternoons       = "TeacherMaxAfternoons"
	C_TeacherLunchBreak          = "TeacherLunchBreak"
)

// ++ TeacherNotAvailable

// TimeSlots in which the teacher is not available.
func (db *DbTopLevel) NewTeacherNotAvailable(
	id NodeRef, weight int, tid NodeRef, notAvailable []TimeSlot,
) *Constraint {
	c := &Constraint{
		CType:  C_TeacherNotAvailable,
		Id:     id,
		Weight: weight,
		Data:   ResourceNotAvailable{tid, notAvailable},
	}
	db.addConstraint(c)
	return c
}

// ++ TeacherMinActivitiesPerDay

// Double time-slots count as two activities, etc.
func (db *DbTopLevel) NewTeacherMinActivitiesPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  C_TeacherMinActivitiesPerDay,
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	db.addConstraint(c)
	return c
}

// ++ TeacherMaxActivitiesPerDay

// Double time-slots count as two activities, etc.
func (db *DbTopLevel) NewTeacherMaxActivitiesPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  C_TeacherMaxActivitiesPerDay,
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	db.addConstraint(c)
	return c
}

// ++ TeacherMaxDays

// Double time-slots count as two activities, etc.
func (db *DbTopLevel) NewTeacherMaxDays(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  C_TeacherMaxDays,
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	db.addConstraint(c)
	return c
}

// ++ TeacherMaxGapsPerDay

func (db *DbTopLevel) NewTeacherMaxGapsPerDay(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  C_TeacherMaxGapsPerDay,
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	db.addConstraint(c)
	return c
}

// ++ TeacherMaxGapsPerWeek

func (db *DbTopLevel) NewTeacherMaxGapsPerWeek(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  C_TeacherMaxGapsPerWeek,
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	db.addConstraint(c)
	return c
}

// ++ TeacherMaxAfternoons

func (db *DbTopLevel) NewTeacherMaxAfternoons(
	id NodeRef, weight int, tid NodeRef, n int,
) *Constraint {
	c := &Constraint{
		CType:  C_TeacherMaxAfternoons,
		Id:     id,
		Weight: weight,
		Data:   ResourceN{tid, n},
	}
	db.addConstraint(c)
	return c
}

// ++ TeacherLunchBreak

func (db *DbTopLevel) NewTeacherLunchBreak(
	id NodeRef, weight int, tid NodeRef,
) *Constraint {
	c := &Constraint{
		CType:  C_TeacherLunchBreak,
		Id:     id,
		Weight: weight,
		Data:   tid,
	}
	db.addConstraint(c)
	return c
}
