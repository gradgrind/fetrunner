package w365tt

import (
	"fetrunner/base"
	"fetrunner/db"
)

func (dbi *W365TopLevel) readLessons(newdb *db.DbTopLevel) {
	for _, e := range dbi.Lessons {
		// The course must be Course or SuperCourse.
		_, ok := dbi.CourseMap[e.Course]
		if !ok {
			base.Error.Fatalf(
				"Lesson %s:\n  Invalid course: %s\n",
				e.Id, e.Course)
		}
		// Check the Rooms.
		reflist := []NodeRef{}
		for _, rref := range e.Rooms {
			_, ok := dbi.RealRooms[rref]
			if ok {
				reflist = append(reflist, rref)
			} else {
				base.Error.Printf(
					"Invalid Room in Lesson %s:\n  %s\n",
					e.Id, rref)
			}
		}
		n := newdb.NewActivity(e.Id)
		n.Course = e.Course
		n.Duration = e.Duration

		// +++ Add constraints ...

		if e.Day >= 0 && e.Hour >= 0 {
			if e.Fixed {
				newdb.NewActivityStartTime(
					"", db.MAXWEIGHT, e.Id, e.Day, e.Hour)
			}
			newdb.AddActivityPlacement("", e.Id, e.Day, e.Hour, reflist)
		}
	}
}
