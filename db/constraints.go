package db

const MAXWEIGHT = 100

func (db *DbTopLevel) addConstraint(c *Constraint) {
    db.Constraints[c.CType] = append(db.Constraints[c.CType], c)
}

// +++ For teacher, class and room constraints

type ResourceN struct {
    Resource NodeRef
    N        int
}

type ResourceNotAvailable struct {
    Resource NodeRef
    // NotAvailable is an ordered list of time-slots in which the teacher
    // is to be regarded as not available for the timetable.
    NotAvailable []TimeSlot
}

// ---

var (
    C_ActivitiesEndDay            = "ActivitiesEndDay"
    C_AfterHour                   = "AfterHour"
    C_BeforeHour                  = "BeforeHour"
    C_AutomaticDifferentDays      = "AutomaticDifferentDays"
    C_DaysBetween                 = "DaysBetween"
    C_DaysBetweenJoin             = "DaysBetweenJoin"
    C_MinHoursFollowing           = "MinHoursFollowing"
    C_DoubleActivityNotOverBreaks = "DoubleActivityNotOverBreaks"
    C_ParallelCourses             = "ParallelCourses"

    C_SetStartingTime = "SetStartingTime"
    C_SetRooms        = "SetRooms"
)

// ++ ActivitiesEndDay

func (db *DbTopLevel) NewActivitiesEndDay(
    id NodeRef, weight int, course NodeRef,
) *Constraint {
    c := &Constraint{
        CType:  C_ActivitiesEndDay,
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
    Courses []NodeRef // Courses or SuperCourses
    Hour    int
}

func (db *DbTopLevel) NewAfterHour(
    id NodeRef, weight int, courses []NodeRef, hour int,
) *Constraint {
    c := &Constraint{
        CType:  C_AfterHour,
        Id:     id,
        Weight: weight,
        Data:   BeforeAfterHour{courses, hour},
    }
    db.addConstraint(c)
    return c
}

func (db *DbTopLevel) NewBeforeHour(
    id NodeRef, weight int, courses []NodeRef, hour int,
) *Constraint {
    c := &Constraint{
        CType:  C_BeforeHour,
        Id:     id,
        Weight: weight,
        Data:   BeforeAfterHour{courses, hour},
    }
    db.addConstraint(c)
    return c
}

// ++ AutomaticDifferentDays

// This Constraint applies to all courses (with more than one Activity).
// If not present, all courses will by default apply it as a hard constraint,
// except for courses which have an overriding DAYS_BETWEEN constraint.
func (db *DbTopLevel) NewAutomaticDifferentDays(
    id NodeRef, weight int, consecutiveIfSameDay bool,
) *Constraint {
    c := &Constraint{
        CType:  C_AutomaticDifferentDays,
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
    Courses              []NodeRef // Courses or SuperCourses
    DaysBetween          int
    ConsecutiveIfSameDay bool
}

func (db *DbTopLevel) NewDaysBetween(
    id NodeRef, weight int,
    courses []NodeRef, daysBetween int, consecutiveIfSameDay bool,
) *Constraint {
    c := &Constraint{
        CType:  C_DaysBetween,
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
    Course1              NodeRef // Course or SuperCourse
    Course2              NodeRef // Course or SuperCourse
    DaysBetween          int
    ConsecutiveIfSameDay bool
}

func (db *DbTopLevel) NewDaysBetweenJoin(
    id NodeRef, weight int,
    course1 NodeRef, course2 NodeRef, daysBetween int, consecutiveIfSameDay bool,
) *Constraint {
    c := &Constraint{
        CType:  C_DaysBetweenJoin,
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
    id NodeRef, weight int, courses []NodeRef,
) *Constraint {
    c := &Constraint{
        CType:  C_ParallelCourses,
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
    id NodeRef, weight int, hours []int,
) *Constraint {
    c := &Constraint{
        CType:  C_DoubleActivityNotOverBreaks,
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
    Course1 NodeRef // Course or SuperCourse
    Course2 NodeRef // Course or SuperCourse
    Hours   int
}

func (db *DbTopLevel) NewMinHoursFollowing(
    id NodeRef, weight int,
    course1 NodeRef, course2 NodeRef, hours int,
) *Constraint {
    c := &Constraint{
        CType:  C_MinHoursFollowing,
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

//TODO ... more?
