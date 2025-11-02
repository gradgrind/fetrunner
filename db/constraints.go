package db

const MAXWEIGHT = 100

func (db *DbTopLevel) addConstraint(c *Constraint) {
	db.Constraints = append(db.Constraints, c)
}

// ++ ActivitiesEndDay

func (db *DbTopLevel) NewActivitiesEndDay(
	id Ref, weight int, course Ref,
) *Constraint {
	c := &Constraint{
		CType:  "ActivitiesEndDay",
		Id:     id,
		Weight: weight,
		Data:   course, // Course or SuperCourse
	}
	db.addConstraint(c)
	return c
}

// ++ BeforeAfterHour

// Permissible starting hours are before or after the specified hour,
// not including the specified hour.
type BeforeAfterHour struct {
	Courses []Ref // Courses or SuperCourses
	After   bool  // false => before given hour, true => after given hour
	Hour    int
}

func (db *DbTopLevel) NewBeforeAfterHour(
	id Ref, weight int, courses []Ref, after bool, hour int,
) *Constraint {
	c := &Constraint{
		CType:  "BeforeAfterHour",
		Id:     id,
		Weight: weight,
		Data:   BeforeAfterHour{courses, after, hour},
	}
	db.addConstraint(c)
	return c
}

// ++ AutomaticDifferentDays

// This Constraint applies to all courses (with more than one Activity).
// If not present, all courses will by default apply it as a hard constraint,
// except for courses which have an overriding DAYS_BETWEEN constraint.
func (db *DbTopLevel) NewAutomaticDifferentDays(
	id Ref, weight int, consecutiveIfSameDay bool,
) *Constraint {
	c := &Constraint{
		CType:  "AutomaticDifferentDays",
		Id:     id,
		Weight: weight,
		Data:   consecutiveIfSameDay,
	}
	db.addConstraint(c)
	return c
}

// ++ DaysBetween

// This constraint applies between the activitys of the individual courses.
// It does not connect the courses. If DaysBetween = 1, this constraint
// overrides the global AutomaticDifferentDays constraint for these courses.
type DaysBetween struct {
	Courses              []Ref // Courses or SuperCourses
	DaysBetween          int
	ConsecutiveIfSameDay bool
}

func (db *DbTopLevel) NewDaysBetween(
	id Ref, weight int,
	courses []Ref, daysBetween int, consecutiveIfSameDay bool,
) *Constraint {
	c := &Constraint{
		CType:  "DaysBetween",
		Id:     id,
		Weight: weight,
		Data: DaysBetween{
			Courses:              courses,
			DaysBetween:          daysBetween,
			ConsecutiveIfSameDay: consecutiveIfSameDay},
	}
	db.addConstraint(c)
	return c
}

// ++ DaysBetweenJoin

// This constraint applies between the individual activities of the two courses,
// not between the activities of a course itself. That is, between course 1,
// activity 1 and course 2 activity 1; between course 1, activity 1 and course 2,
// activity 2, etc.

type DaysBetweenJoin struct {
	Course1              Ref // Course or SuperCourse
	Course2              Ref // Course or SuperCourse
	DaysBetween          int
	ConsecutiveIfSameDay bool
}

func (db *DbTopLevel) NewDaysBetweenJoin(
	id Ref, weight int,
	course1 Ref, course2 Ref, daysBetween int, consecutiveIfSameDay bool,
) *Constraint {
	c := &Constraint{
		CType:  "DaysBetweenJoin",
		Id:     id,
		Weight: weight,
		Data: DaysBetweenJoin{
			Course1:              course1,
			Course2:              course2,
			DaysBetween:          daysBetween,
			ConsecutiveIfSameDay: consecutiveIfSameDay},
	}
	db.addConstraint(c)
	return c
}

// ++ ParallelCourses

// The activities of the courses specified here should be at the same time.
// To avoid complications, it is required that the number and lengths of
// activities be the same in each course.
func (db *DbTopLevel) NewParallelCourses(
	id Ref, weight int, courses []Ref,
) *Constraint {
	c := &Constraint{
		CType:  "ParallelCourses",
		Id:     id,
		Weight: weight,
		Data:   courses, // Courses or SuperCourses
	}
	db.addConstraint(c)
	return c
}

// ++ DoubleActivityNotOverBreaks

// There should be at most one of these. The breaks are immediately before
// the specified hours.
func (db *DbTopLevel) NewDoubleActivityNotOverBreaks(
	id Ref, weight int, hours []int,
) *Constraint {
	c := &Constraint{
		CType:  "DoubleActivityNotOverBreaks",
		Id:     id,
		Weight: weight,
		Data:   hours,
	}
	db.addConstraint(c)
	return c
}

// ++ MinHoursFollowing

// The start of an activity in `Course2` should be at least `Hours` after
// the end of an activity in `Course1`.
type MinHoursFollowing struct {
	Course1 Ref // Course or SuperCourse
	Course2 Ref // Course or SuperCourse
	Hours   int
}

func (db *DbTopLevel) NewMinHoursFollowing(
	id Ref, weight int,
	course1 Ref, course2 Ref, hours int,
) *Constraint {
	c := &Constraint{
		CType:  "MinHoursFollowing",
		Id:     id,
		Weight: weight,
		Data: MinHoursFollowing{
			Course1: course1,
			Course2: course2,
			Hours:   hours,
		},
	}
	db.addConstraint(c)
	return c
}

//TODO ... continue conversion to new Constraint structure ...

/* TODO: Is this really useful? The W365 front end doesn't currently support it
// and the MinHoursFollowing may be more useful.
// ++ NotOnSameDay

func (db *DbTopLevel) NewNotOnSameDay(
	id Ref, weight int, subjects []Ref,
) *Constraint {
	c := &Constraint{
		CType:  "NotOnSameDay",
		Id:     id,
		Weight: weight,
		Data:   subjects,
	}
	db.addConstraint(c)
	return c
}
*/

//TODO ... more?
