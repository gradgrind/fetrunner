// Package timetable provides structures and associated functions for
// supporting the management of timetables.
package timetable

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
)

type NodeRef = base.NodeRef // node reference (UUID)
type element = base.ElementBase

type ActivityIndex = autotimetable.ActivityIndex
type TeacherIndex = autotimetable.TeacherIndex
type RoomIndex = autotimetable.RoomIndex
type ClassIndex = autotimetable.ClassIndex
type AtomicIndex = autotimetable.AtomicIndex
type TtClass = autotimetable.TtClass
type TtGroup = autotimetable.TtGroup

type constraint = autotimetable.TtConstraint
type constraintIndex = autotimetable.ConstraintIndex
type autoTtData = autotimetable.AutoTtData
type constraintType = autotimetable.ConstraintType

type TtData struct {
	db *base.DbTopLevel

	constraints        []*constraint     // ordered constraint list for "autotimetable"
	hard_not_available []constraintIndex // list of hard "not available" constraints
	nConstraints       constraintIndex
	constraintTypes    []constraintType
	hardConstraintMap  map[constraintType][]constraintIndex
	softConstraintMap  map[constraintType][]constraintIndex

	ndays        int
	nhours       int
	hoursperweek int

	AtomicGroups []*AtomicGroup

	Teacher2Index map[NodeRef]TeacherIndex
	teachers      []element
	Room2Index    map[NodeRef]RoomIndex
	rooms         []element
	Class2Index   map[NodeRef]ClassIndex

	// `AtomicGroup2Indexes` maps a class or group NodeRef to its list of atomic
	// group indexes.
	AtomicGroup2Indexes map[NodeRef][]AtomicIndex

	// `ClassDivisions` is a list with an entry for each class, containing a
	// list of its divisions ([][]NodeRef).
	ClassDivisions []ClassDivision

	// Sorted list of constraint names (only those presently used in the data)
	ConstraintTypes []constraintType

	// Set up by `CollectCourses`, which calls `makeActivities`
	TtActivities      []*TtActivity
	Ref2ActivityIndex map[NodeRef]ActivityIndex
	CourseInfoList    []*CourseInfo
	Ref2CourseInfo    map[NodeRef]*CourseInfo

	// Transformed constraints
	MinDaysBetweenActivities []*TtDaysBetween
	ParallelActivities       []*TtParallelActivities
}

func (tt_data *TtData) GetDays() []element {
	dlist := []element{}
	for _, d := range tt_data.db.Days {
		dlist = append(dlist, element{Id: d.Id, Tag: d.Tag})
	}
	return dlist
}

func (tt_data *TtData) GetHours() []element {
	hlist := []element{}
	for _, h := range tt_data.db.Hours {
		hlist = append(hlist, element{Id: h.Id, Tag: h.Tag})
	}
	return hlist
}

func (tt_data *TtData) GetClasses() []*TtClass {
	db := tt_data.db
	clist := make([]*TtClass, len(tt_data.ClassDivisions))
	for i, c := range tt_data.ClassDivisions {
		glist := []*TtGroup{}
		for _, d := range c.Divisions {
			for _, g := range d {
				e := db.GetElement(g)
				id := e.GetRef()
				glist = append(glist, &TtGroup{
					Id:            id,
					Tag:           e.GetTag(),
					ClassIndex:    i,
					AtomicIndexes: tt_data.AtomicGroup2Indexes[id],
				})
			}
		}
		clist[i] = &TtClass{
			Id:            c.Class.Id,
			Tag:           c.Class.Tag,
			AtomicIndexes: tt_data.AtomicGroup2Indexes[c.Class.Id],
			Groups:        glist,
		}
	}
	return clist
}

func (tt_data *TtData) GetAtomicGroups() []string {
	aglist := []string{}
	for _, ag := range tt_data.AtomicGroups {
		aglist = append(aglist, ag.Tag)
	}
	return aglist
}

func (tt_data *TtData) GetRooms() []element {
	return tt_data.rooms
}

func (tt_data *TtData) GetSubjects() []element {
	subjects := make([]element, len(tt_data.db.Subjects))
	for i, s := range tt_data.db.Subjects {
		subjects[i] = element{Id: s.Id, Tag: s.Tag}
	}
	return subjects
}

func (tt_data *TtData) GetTeachers() []element {
	return tt_data.teachers
}

// A `CourseInfo` is a representation of a course (Course or SuperCourse) for
// the timetable.
// Activities within a course are (already) ordered, highest duration first,
// and the Activities field has the same order.
type CourseInfo struct {
	Id                 NodeRef // Course or SuperCourse
	Subject            string
	Groups             []*base.Group // a `Class` is represented by its ClassGroup
	AtomicGroupIndexes []AtomicIndex
	Teachers           []TeacherIndex
	FixedRooms         []RoomIndex
	RoomChoices        [][]RoomIndex
	Activities         []ActivityIndex
}

// TODO: Add node id field? And where is the other info, like duration?
type TtActivity struct {
	CourseInfo     int            // index to `TtData.CourseInfoList`
	FixedStartTime *base.TimeSlot // needed for days-between preparation
}

type ClassDivision struct {
	Class     *base.Class
	Divisions [][]NodeRef
}

// MakeTimetableData performs the initialization of a TtData structure, collecting
// "resources" (atomic student groups, teachers and rooms) and "activities".
func MakeTimetableData(bd *base.BaseData) *TtData {
	db := bd.Db
	days := len(db.Days)
	hours := len(db.Hours)
	tt_data := &TtData{
		db: db,

		ndays:        days,
		nhours:       hours,
		hoursperweek: days * hours,
	}

	// Collect ClassDivisions
	tt_data.FilterDivisions()

	// Atomic groups: an atomic group is a "resource", it is an ordered list
	// of single groups, one from each division.
	tt_data.MakeAtomicGroups()

	// Build maps to convert teacher and room ids to indexes
	tt_data.TeacherResources()
	tt_data.RoomResources()

	// Get the courses (-> CourseInfo) and activities for the timetable
	tt_data.CollectCourses(bd)

	// Collect constraints
	tt_data.prepare_constraints()

	//TODO: Perhaps this should be called from the back-end, in preparation
	// for a generator run?
	tt_data.preprocessConstraints(bd)

	return tt_data
}

func (tt_data *TtData) TeacherResources() {
	tt_data.Teacher2Index = map[NodeRef]TeacherIndex{}
	tt_data.teachers = make([]element, len(tt_data.db.Teachers))
	for i, t := range tt_data.db.Teachers {
		tt_data.Teacher2Index[t.Id] = i
		tt_data.teachers[i] = element{Id: t.Id, Tag: t.Tag}
	}
}

func (tt_data *TtData) RoomResources() {
	tt_data.Room2Index = map[NodeRef]RoomIndex{}
	tt_data.rooms = make([]element, len(tt_data.db.Rooms))
	for i, r := range tt_data.db.Rooms {
		tt_data.Room2Index[r.Id] = i
		tt_data.rooms[i] = element{Id: r.Id, Tag: r.Tag}
	}
}

func (tt_data *TtData) GetActivities() []element {
	activities := tt_data.db.Activities
	alist := make([]element, len(activities))
	for i, a := range activities {
		alist[i] = element{Id: a.Id} // no Tag field
	}
	return alist
}

// TODO: Is this really needed?
func (tt_data *TtData) GetNActivities() int {
	return len(tt_data.TtActivities)
}

func (tt_data *TtData) GetConstraint_Types() []constraintType {
	return tt_data.ConstraintTypes
}

// TODO: Is this really needed?
func (tt_data *TtData) GetNConstraints() constraintIndex {
	return len(tt_data.constraints)
}

func (tt_data *TtData) GetHardConstraintMap() map[constraintType][]constraintIndex {
	return tt_data.hardConstraintMap
}

func (tt_data *TtData) GetSoftConstraintMap() map[constraintType][]constraintIndex {
	return tt_data.softConstraintMap
}

// TODO: This seems to be used only for the result presentation.
// Is it really necessary? If so, what should it contain?
func (tt_data *TtData) GetConstraints() []*constraint {
	return tt_data.constraints
}
