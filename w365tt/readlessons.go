package w365tt

import (
	"fetrunner/base"
	"fetrunner/db"
)

func (dbi *DbTopLevel) readLessons(newdb *db.DbTopLevel) {
	for _, e := range dbi.Lessons {
		// The course must be Course or Supercourse.
		_, ok := dbi.CourseMap[e.Course]
		if !ok {
			base.Error.Fatalf(
				"Lesson %s:\n  Invalid course: %s\n",
				e.Id, e.Course)
		}
		// Check the Rooms.
		reflist := []Ref{}
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
		n.Day = e.Day
		n.Hour = e.Hour
		n.Fixed = e.Fixed
		n.Rooms = reflist
		//n.Flags = e.Flags
		//n.Background = e.Background
		//n.Footnote = e.Footnote
	}
}
