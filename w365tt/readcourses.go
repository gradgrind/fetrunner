package w365tt

import (
	"fetrunner/base"
	"fetrunner/db"
	"strings"
)

func (dbi *W365TopLevel) readSubjects(newdb *db.DbTopLevel) {
	dbi.SubjectMap = map[NodeRef]*db.Subject{}
	dbi.SubjectTags = map[string]NodeRef{}
	for _, e := range dbi.Subjects {
		// Perform some checks and add to the SubjectTags map.
		_, nok := dbi.SubjectTags[e.Tag]
		if nok {
			base.Error.Fatalf("Subject Tag (Shortcut) defined twice: %s\n",
				e.Tag)
		}
		dbi.SubjectTags[e.Tag] = e.Id
		//Copy data to base db.
		n := newdb.NewSubject(e.Id)
		n.Tag = e.Tag
		n.Name = e.Name
		dbi.SubjectMap[e.Id] = n
	}
}

func (dbi *W365TopLevel) makeNewSubject(
	newdb *db.DbTopLevel,
	tag string,
	name string,
) NodeRef {
	s := newdb.NewSubject("")
	s.Tag = tag
	s.Name = name
	dbi.SubjectTags[tag] = s.Id
	return s.Id
}

func (dbi *W365TopLevel) readCourses(newdb *db.DbTopLevel) {
	dbi.CourseMap = map[NodeRef]struct{}{}
	for _, e := range dbi.Courses {
		subject := dbi.getCourseSubject(newdb, e.Subjects, e.Id)
		room := dbi.getCourseRoom(newdb, e.PreferredRooms, e.Id)
		groups := dbi.getCourseGroups(e.Groups, e.Id)
		teachers := dbi.getCourseTeachers(e.Teachers, e.Id)
		n := newdb.NewCourse(e.Id)
		n.Subject = subject
		n.Groups = groups
		n.Teachers = teachers
		n.Room = room
		dbi.CourseMap[e.Id] = struct{}{}
	}
}

func (dbi *W365TopLevel) readSuperCourses(newdb *db.DbTopLevel) {
	// In the input from W365 the subjects for the SuperCourses must be
	// taken from the linked EpochPlan.
	// The EpochPlans are otherwise not needed.
	epochPlanSubjects := map[NodeRef]NodeRef{}
	if dbi.EpochPlans != nil {
		for _, n := range dbi.EpochPlans {
			sref, ok := dbi.SubjectTags[n.Tag]
			if !ok {
				sref = dbi.makeNewSubject(newdb, n.Tag, n.Name)
			}
			epochPlanSubjects[n.Id] = sref
		}
	}

	sbcMap := map[NodeRef]*db.SubCourse{}
	for _, spc := range dbi.SuperCourses {
		// Read the SubCourses.
		for _, e := range spc.SubCourses {
			sbc, ok := sbcMap[e.Id]
			if ok {
				// Assume the SubCourse really is the same.
				sbc.SuperCourses = append(sbc.SuperCourses, spc.Id)
			} else {
				subject := dbi.getCourseSubject(newdb, e.Subjects, e.Id)
				room := dbi.getCourseRoom(newdb, e.PreferredRooms, e.Id)
				groups := dbi.getCourseGroups(e.Groups, e.Id)
				teachers := dbi.getCourseTeachers(e.Teachers, e.Id)
				// Use a new Id for the SubCourse because it can also be
				// the Id of a Course.
				n := newdb.NewSubCourse("$$" + e.Id)
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
			base.Error.Fatalf("Unknown EpochPlan in SuperCourse %s:\n  %s\n",
				spc.Id, spc.EpochPlan)
		}
		n := newdb.NewSuperCourse(spc.Id)
		n.Subject = subject
		dbi.CourseMap[n.Id] = struct{}{}
	}
}

func (dbi *W365TopLevel) getCourseSubject(
	newdb *db.DbTopLevel,
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
	msg := "Course %s:\n  Not a Subject: %s\n"
	var subject NodeRef
	if len(srefs) == 1 {
		wsid := srefs[0]
		_, ok := dbi.SubjectMap[wsid]
		if !ok {
			base.Error.Fatalf(msg, courseId, wsid)
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
				base.Error.Fatalf(msg, courseId, wsid)
			}
		}
		sktag := strings.Join(sklist, "/")
		wsid, ok := dbi.SubjectTags[sktag]
		if ok {
			// The name has already been used.
			subject = wsid
		} else {
			// Need a new Subject.
			subject = dbi.makeNewSubject(newdb, sktag, "Compound Subject")
		}
	} else {
		base.Error.Printf("Course/SubCourse has no subject: %s\n", courseId)
		// Use a dummy Subject.
		var ok bool
		subject, ok = dbi.SubjectTags["?"]
		if !ok {
			subject = dbi.makeNewSubject(newdb, "?", "No Subject")
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
	newdb *db.DbTopLevel,
	rrefs []NodeRef,
	courseId NodeRef,
) NodeRef {
	room := NodeRef("")
	if len(rrefs) > 1 {
		// Make a RoomChoiceGroup
		var estr string
		room, estr = dbi.makeRoomChoiceGroup(newdb, rrefs)
		if estr != "" {
			base.Error.Printf("In Course %s:\n%s", courseId, estr)
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
				base.Error.Printf("Invalid room in Course/SubCourse %s:\n  %s\n",
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
			base.Error.Fatalf("Invalid group in Course/SubCourse %s:\n  %s\n",
				courseId, gref)
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
			base.Error.Fatalf("Unknown teacher in Course %s:\n  %s\n",
				courseId, tref)
		}
		tlist = append(tlist, tref)
	}
	return tlist
}
