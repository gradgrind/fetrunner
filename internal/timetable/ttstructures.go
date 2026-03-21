// Package timetable provides structures and associated functions for
// supporting the management of timetables.
package timetable

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"maps"
)

type nodeRef = base.NodeRef // node reference (UUID)
type element = base.ElementBase

type activityIndex = autotimetable.ActivityIndex
type teacherIndex = autotimetable.TeacherIndex
type roomIndex = autotimetable.RoomIndex
type classIndex = autotimetable.ClassIndex
type atomicIndex = autotimetable.AtomicIndex
type ttClass = autotimetable.TtClass
type ttGroup = autotimetable.TtGroup

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

	Teacher2Index map[nodeRef]teacherIndex
	teachers      []element
	Room2Index    map[nodeRef]roomIndex
	rooms         []element
	Class2Index   map[nodeRef]classIndex

	// `AtomicGroup2Indexes` maps a class or group NodeRef to its list of atomic
	// group indexes.
	AtomicGroup2Indexes map[nodeRef][]atomicIndex

	// `ClassDivisions` is a list with an entry for each class, containing a
	// list of its divisions ([][]NodeRef).
	ClassDivisions []ClassDivision

	// Sorted list of constraint names (only those presently used in the data)
	ConstraintTypes []constraintType

	// Set up by `CollectCourses`, which calls `makeActivities`
	TtActivities      []*TtActivity
	Ref2ActivityIndex map[nodeRef]activityIndex
	CourseInfoList    []*CourseInfo
	Ref2CourseInfo    map[nodeRef]*CourseInfo
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

func (tt_data *TtData) GetClasses() []*ttClass {
	db := tt_data.db
	clist := make([]*ttClass, len(tt_data.ClassDivisions))
	for i, c := range tt_data.ClassDivisions {
		glist := []*ttGroup{}
		for _, d := range c.Divisions {
			for _, g := range d {
				e := db.GetElement(g)
				id := e.GetRef()
				glist = append(glist, &ttGroup{
					Id:            id,
					Tag:           e.GetTag(),
					ClassIndex:    i,
					AtomicIndexes: tt_data.AtomicGroup2Indexes[id],
				})
			}
		}
		clist[i] = &ttClass{
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
	Id                 nodeRef // Course or SuperCourse
	Subject            string
	Groups             []*base.Group // a `Class` is represented by its ClassGroup
	AtomicGroupIndexes []atomicIndex
	Teachers           []teacherIndex
	FixedRooms         []roomIndex
	RoomChoices        [][]roomIndex
	Activities         []activityIndex
}

// TODO: Add node id field? And where is the other info, like duration?
type TtActivity struct {
	CourseInfo     int            // index to `TtData.CourseInfoList`
	fixedStartTime *base.TimeSlot // needed for days-between preparation
}

type ClassDivision struct {
	Class     *base.Class
	Divisions [][]nodeRef
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

	// Use a copy of the constraints map so that it can be used destructively,
	// deleting entries as they are processed.
	constraint_map := maps.Clone(db.Constraints)
	tt_data.get_blocked_slots(constraint_map)
	tt_data.placement_constraints(constraint_map)

	// Prepare for the generation of new constraints where these are implied
	// by certain special constraints. This must be after the call to
	// `placement_constraints` as that sets up the hard fixed starting times,
	// which are needed for the generation of the days-between constraints.
	tt_data.prepare_days_between(bd, constraint_map)
	tt_data.prepare_parallels(bd, constraint_map)

	// Collect the remaining constraints.
	tt_data.prepare_constraints(constraint_map)
	for c := range constraint_map {
		bd.Logger.Error("UnhandledConstraintType: %s", c)
	}

	return tt_data
}

func (tt_data *TtData) TeacherResources() {
	tt_data.Teacher2Index = map[nodeRef]teacherIndex{}
	tt_data.teachers = make([]element, len(tt_data.db.Teachers))
	for i, t := range tt_data.db.Teachers {
		tt_data.Teacher2Index[t.Id] = i
		tt_data.teachers[i] = element{Id: t.Id, Tag: t.Tag}
	}
}

func (tt_data *TtData) RoomResources() {
	tt_data.Room2Index = map[nodeRef]roomIndex{}
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
