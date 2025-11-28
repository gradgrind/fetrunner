package w365tt

import (
	"fetrunner/base"
)

func (dbi *W365TopLevel) readLessons(newdb *base.BaseData) {
	logger := newdb.Logger
	for _, e := range dbi.Lessons {
		// The course must be Course or SuperCourse.
		_, ok := dbi.CourseMap[e.Course]
		if !ok {
			logger.Error(
				"Lesson %s:\n  Invalid course: %s\n",
				e.Id, e.Course)
			continue
		}
		// Check the Rooms.
		reflist := []NodeRef{}
		for _, rref := range e.Rooms {
			_, ok := dbi.RealRooms[rref]
			if ok {
				reflist = append(reflist, rref)
			} else {
				logger.Error(
					"Invalid Room in Lesson %s:\n  %s\n",
					e.Id, rref)
			}
		}
		n := newdb.NewActivity(e.Id)
		n.Course = e.Course
		n.Duration = e.Duration

		// +++ Add constraints ...
		ndb := newdb.Db

		if e.Day >= 0 && e.Hour >= 0 {
			if e.Fixed {
				ndb.NewActivityStartTime(
					"", base.MAXWEIGHT, e.Id, e.Day, e.Hour)
			}
			ndb.AddActivityPlacement("", e.Id, e.Day, e.Hour, reflist)
		}
	}
}
