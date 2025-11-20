package w365tt

import (
	"fetrunner/base"
	"fetrunner/db"
	"strings"
)

func (dbi *W365TopLevel) readSubjects(newdb *db.DbTopLevel) {
	logger := newdb.Logger
	dbi.SubjectMap = map[NodeRef]*db.Subject{}
	dbi.SubjectTags = map[string]NodeRef{}
	for _, e := range dbi.Subjects {
		// Perform some checks and add to the SubjectTags map.
	sloop:
		_, nok := dbi.SubjectTags[e.Tag]
		if nok {
			logger.Error("Subject Tag (Shortcut) defined twice: %s\n",
				e.Tag)
			e.Tag += "$"
			goto sloop
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
	logger := newdb.Logger
	dbi.CourseMap = map[NodeRef]struct{}{}
	for _, e := range dbi.Courses {
		subject := dbi.getCourseSubject(newdb, e.Subjects, e.Id)
		room := dbi.getCourseRoom(newdb, e.PreferredRooms, e.Id)
		groups := dbi.getCourseGroups(logger, e.Groups, e.Id)
		teachers := dbi.getCourseTeachers(logger, e.Teachers, e.Id)
		n := newdb.NewCourse(e.Id)
		n.Subject = subject
		n.Groups = groups
		n.Teachers = teachers
		n.Room = room
		dbi.CourseMap[e.Id] = struct{}{}
	}
}

func (dbi *W365TopLevel) readSuperCourses(newdb *db.DbTopLevel) {
	logger := newdb.Logger
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
				groups := dbi.getCourseGroups(logger, e.Groups, e.Id)
				teachers := dbi.getCourseTeachers(logger, e.Teachers, e.Id)
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
			logger.Error("Unknown EpochPlan in SuperCourse %s:\n  %s\n",
				spc.Id, spc.EpochPlan)
			continue
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
	logger := newdb.Logger
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
			logger.Error(msg, courseId, wsid)
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
				logger.Error(msg, courseId, wsid)
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
			subject = dbi.makeNewSubject(newdb, sktag, "Compound Subject")
		}
	} else {
		logger.Error("Course/SubCourse has no subject: %s\n", courseId)
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
	logger := newdb.Logger
	room := NodeRef("")
	if len(rrefs) > 1 {
		// Make a RoomChoiceGroup
		var estr string
		room, estr = dbi.makeRoomChoiceGroup(newdb, rrefs)
		if estr != "" {
			logger.Error("In Course %s:\n%s", courseId, estr)
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
				logger.Error("Invalid room in Course/SubCourse %s:\n  %s\n",
					courseId, rref0)
			}
		}
	}
	return room
}

func (dbi *W365TopLevel) getCourseGroups(
	logger *base.LogInstance,
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
			logger.Error("Invalid group in Course/SubCourse %s:\n  %s\n",
				courseId, gref)
			continue
		}
		glist = append(glist, ngref)
	}
	return glist
}

func (dbi *W365TopLevel) getCourseTeachers(
	logger *base.LogInstance,
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
			logger.Error("Unknown teacher in Course %s:\n  %s\n",
				courseId, tref)
		}
		tlist = append(tlist, tref)
	}
	return tlist
}
