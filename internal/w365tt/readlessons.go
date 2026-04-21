package w365tt

import (
    "fetrunner/internal/base"
)

func (dbi *W365TopLevel) readLessons() {
    for _, e := range dbi.Lessons {
        // The course must be Course or SuperCourse.
        _, ok := dbi.CourseMap[e.Course]
        if !ok {
            base.LogError(
                "--W365_LESSON_HAS_INVALID_COURSE Lesson: %s, Course: %s",
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
                base.LogError(
                    "--W365_INVALID_ROOM Lesson: %s, Room: %s",
                    e.Id, rref)
            }
        }
        n := base.NewActivity(e.Id)
        n.Course = e.Course
        n.Duration = e.Duration

        // +++ Add constraints ...
        ndb := base.DataBase.Db

        if e.Day >= 0 && e.Hour >= 0 {
            if e.Fixed {
                ndb.NewActivityStartTime(
                    "", base.MAXWEIGHT, e.Id, e.Day, e.Hour)
            }
            ndb.AddActivityPlacement("", e.Id, e.Day, e.Hour, reflist)
        }
    }
}
