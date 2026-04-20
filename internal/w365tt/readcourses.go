package w365tt

import (
    "fetrunner/internal/base"
    "strings"
)

func (dbi *W365TopLevel) readSubjects() {
    dbi.SubjectMap = map[NodeRef]*base.Subject{}
    dbi.SubjectTags = map[string]NodeRef{}
    for _, e := range dbi.Subjects {
        // Perform some checks and add to the SubjectTags map.
    sloop:
        _, nok := dbi.SubjectTags[e.Tag]
        if nok {
            base.LogError("--SUBJECT_TAG_DEFINED_TWICE %s", e.Tag)
            e.Tag += "$"
            goto sloop
        }
        dbi.SubjectTags[e.Tag] = e.Id
        //Copy data to base db.
        n := base.NewSubject(e.Id)
        n.Tag = e.Tag
        n.Name = e.Name
        dbi.SubjectMap[e.Id] = n
    }
}

func (dbi *W365TopLevel) makeNewSubject(
    tag string,
    name string,
) NodeRef {
    s := base.NewSubject("")
    s.Tag = tag
    s.Name = name
    dbi.SubjectTags[tag] = s.Id
    return s.Id
}

func (dbi *W365TopLevel) readCourses() {
    dbi.CourseMap = map[NodeRef]struct{}{}
    for _, e := range dbi.Courses {
        subject := dbi.getCourseSubject(e.Subjects, e.Id)
        room := dbi.getCourseRoom(e.PreferredRooms, e.Id)
        groups := dbi.getCourseGroups(e.Groups, e.Id)
        teachers := dbi.getCourseTeachers(e.Teachers, e.Id)
        n := base.NewCourse(e.Id)
        n.Subject = subject
        n.Groups = groups
        n.Teachers = teachers
        n.Room = room
        dbi.CourseMap[e.Id] = struct{}{}
    }
}

func (dbi *W365TopLevel) readSuperCourses() {
    // In the input from W365 the subjects for the SuperCourses must be
    // taken from the linked EpochPlan.
    // The EpochPlans are otherwise not needed.
    epochPlanSubjects := map[NodeRef]NodeRef{}
    if dbi.EpochPlans != nil {
        for _, n := range dbi.EpochPlans {
            sref, ok := dbi.SubjectTags[n.Tag]
            if !ok {
                sref = dbi.makeNewSubject(n.Tag, n.Name)
            }
            epochPlanSubjects[n.Id] = sref
        }
    }

    sbcMap := map[NodeRef]*base.SubCourse{}
    for _, spc := range dbi.SuperCourses {
        // Read the SubCourses.
        for _, e := range spc.SubCourses {
            sbc, ok := sbcMap[e.Id]
            if ok {
                // Assume the SubCourse really is the same.
                sbc.SuperCourses = append(sbc.SuperCourses, spc.Id)
            } else {
                subject := dbi.getCourseSubject(e.Subjects, e.Id)
                room := dbi.getCourseRoom(e.PreferredRooms, e.Id)
                groups := dbi.getCourseGroups(e.Groups, e.Id)
                teachers := dbi.getCourseTeachers(e.Teachers, e.Id)
                // Use a new Id for the SubCourse because it can also be
                // the Id of a Course.
                n := base.NewSubCourse("$$" + e.Id)
                n.SuperCourses = []NodeRef{spc.Id}
                n.Subject = subject
                n.Groups = groups
                n.Teachers = teachers
                n.Room = room
                sbcMap[e.Id] = n
            }
        }

        // Now add the SuperCourse.
        subject, ok := epochPlanSubjects[spc.EpochPlan]
        if !ok {
            base.LogError("--W365_UNKNOWN_COURCE_BLOCK SuperCourse: %s, EpochPlan: %s",
                spc.Id, spc.EpochPlan)
            continue
        }
        n := base.NewSuperCourse(spc.Id)
        n.Subject = subject
        dbi.CourseMap[n.Id] = struct{}{}
    }
}

func (dbi *W365TopLevel) getCourseSubject(
    srefs []NodeRef,
    courseId NodeRef,
) NodeRef {
    //
    // Deal with the Subjects field of a Course or SubCourse – W365
    // allows multiple subjects.
    // The base db expects one and only one subject (in the Subject field).
    // If there are multiple subjects in the input, these will be converted
    // to a single "composite" subject, using all the subject tags.
    // Repeated use of the same subject list will reuse the created subject.
    //
    msg := "--W365_UNKNOWN_SUBJECT Course: %s, Subject: %s"
    var subject NodeRef
    if len(srefs) == 1 {
        wsid := srefs[0]
        _, ok := dbi.SubjectMap[wsid]
        if !ok {
            base.LogError(msg, courseId, wsid)
            return ""
        }
        subject = wsid
    } else if len(srefs) > 1 {
        // Make a subject name
        sklist := []string{}
        for _, wsid := range srefs {
            // Need Tag/Shortcut field
            s, ok := dbi.SubjectMap[wsid]
            if ok {
                sklist = append(sklist, s.Tag)
            } else {
                base.LogError(msg, courseId, wsid)
                return ""
            }
        }
        sktag := strings.Join(sklist, "/")
        wsid, ok := dbi.SubjectTags[sktag]
        if ok {
            // The name has already been used.
            subject = wsid
        } else {
            // Need a new Subject.
            subject = dbi.makeNewSubject(sktag, "Compound Subject")
        }
    } else {
        base.LogError("--W365_COURSE_HAS_NO_SUBJECT %s", courseId)
        // Use a dummy Subject.
        var ok bool
        subject, ok = dbi.SubjectTags["?"]
        if !ok {
            subject = dbi.makeNewSubject("?", "No Subject")
        }
    }
    return subject
}

// Deal with rooms. W365 can have a single RoomGroup or a list of Rooms.
// If there is a list of Rooms, this is converted to a RoomChoiceGroup.
// The result should be a single Room, RoomChoiceGroup or RoomGroup
// in the "Room" field.
// If a list of rooms recurs, the same RoomChoiceGroup is used.
func (dbi *W365TopLevel) getCourseRoom(
    rrefs []NodeRef,
    courseId NodeRef,
) NodeRef {
    room := NodeRef("")
    if len(rrefs) > 1 {
        // Make a RoomChoiceGroup
        var estr string
        room, estr = dbi.makeRoomChoiceGroup(rrefs)
        if estr != "" {
            base.LogError("--W365_ROOM_ERROR Course: %s, Room: %s", courseId, estr)
        }
    } else if len(rrefs) == 1 {
        // Check that room is Room or RoomGroup.
        rref0 := rrefs[0]
        _, ok := dbi.RealRooms[rref0]
        if ok {
            room = rref0
        } else {
            _, ok := dbi.RoomGroupMap[rref0]
            if ok {
                room = rref0
            } else {
                base.LogError("--W365_INVALID_ROOM Course/SubCourse: %s, Room: %s",
                    courseId, rref0)
            }
        }
    }
    return room
}

func (dbi *W365TopLevel) getCourseGroups(
    grefs []NodeRef,
    courseId NodeRef,
) []NodeRef {
    //
    // Check the group references and replace Class references by the
    // corresponding whole-class base.Group references.
    //
    glist := []NodeRef{}
    for _, gref := range grefs {
        ngref, ok := dbi.GroupRefMap[gref]
        if !ok {
            base.LogError(
                "--W365_INVALID_GROUP Course/SubCourse: %s, Room: %s",
                courseId, gref)
            continue
        }
        glist = append(glist, ngref)
    }
    return glist
}

func (dbi *W365TopLevel) getCourseTeachers(
    trefs []NodeRef,
    courseId NodeRef,
) []NodeRef {
    //
    // Check the teacher references.
    //
    tlist := []NodeRef{}
    for _, tref := range trefs {
        _, ok := dbi.TeacherMap[tref]
        if !ok {
            base.LogError("--W365_UNKNOWN_TEACHER Course: %s, Teacher: %s",
                courseId, tref)
        }
        tlist = append(tlist, tref)
    }
    return tlist
}
