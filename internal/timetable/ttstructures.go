// Package timetable provides structures and associated functions for
// supporting the management of timetables.
package timetable

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
)

type NodeRef = base.NodeRef // node reference (UUID)

type ActivityIndex = autotimetable.ActivityIndex
type TeacherIndex = autotimetable.TeacherIndex
type RoomIndex = autotimetable.RoomIndex
type ClassIndex = autotimetable.ClassIndex
type AtomicIndex = autotimetable.AtomicIndex

type TtData struct {
	NDays        int
	NHours       int
	HoursPerWeek int
	//BackendData  any // available for the generator back-end

	//TODO??
	AtomicGroups []*AtomicGroup

	TeacherIndex map[NodeRef]TeacherIndex
	RoomIndex    map[NodeRef]RoomIndex
	ClassIndex   map[NodeRef]ClassIndex

	// `AtomicGroupIndex` maps a class or group NodeRef to its list of atomic
	// group indexes.
	AtomicGroupIndex map[NodeRef][]AtomicIndex

	// `ClassDivisions` is a list with an entry for each class, containing a
	// list of its divisions ([][]NodeRef).
	ClassDivisions []ClassDivision

	// Sorted list of constraint names (only those presently used in the data)
	ConstraintTypes []autotimetable.ConstraintType

	// Set up by `CollectCourses`, which calls `makeActivities`

	//TODO: Still valid? I think it is confusing and error prone! All other
	// indexes are 0-based.
	// Note that activity 0 is invalid, the first activity has index 1.
	TtActivities      []*TtActivity
	Ref2ActivityIndex map[NodeRef]ActivityIndex
	CourseInfoList    []*CourseInfo
	Ref2CourseInfo    map[NodeRef]*CourseInfo

	// Transformed constraints
	MinDaysBetweenActivities []*TtDaysBetween
	ParallelActivities       []*TtParallelActivities
}

// A `CourseInfo` is a representation of a course (Course or SuperCourse) for
// the timetable.
// Activities within a course are (already) ordered, highest duration first,
// and the Activities field has the same order.
type CourseInfo struct {
	Id           NodeRef // Course or SuperCourse
	Subject      string
	Groups       []*base.Group // a `Class` is represented by its ClassGroup
	AtomicGroups []AtomicIndex
	Teachers     []TeacherIndex
	FixedRooms   []RoomIndex
	RoomChoices  [][]RoomIndex
	Activities   []ActivityIndex
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
		NDays:        days,
		NHours:       hours,
		HoursPerWeek: days * hours,
	}

	// Collect ClassDivisions
	tt_data.FilterDivisions(db)

	// Atomic groups: an atomic group is a "resource", it is an ordered list
	// of single groups, one from each division.
	// The atomic groups take the lowest resource indexes (starting at 0).
	// `AtomicGroups` maps the classes and groups to a list of their resource
	// indexes.
	tt_data.MakeAtomicGroups(db)

	// Add teachers and rooms to resource array
	tt_data.TeacherResources(db)
	tt_data.RoomResources(db)

	// Get the courses (-> CourseInfo) and activities for the timetable
	tt_data.CollectCourses(bd)

	//TODO: Perhaps this should be called from the back-end, in preparation
	// for a generator run?
	tt_data.preprocessConstraints(bd)

	return tt_data
}

func (tt_data *TtData) TeacherResources(db *base.DbTopLevel) {
	tt_data.TeacherIndex = map[NodeRef]TeacherIndex{}
	for i, t := range db.Teachers {
		tt_data.TeacherIndex[t.Id] = i
	}
}

func (tt_data *TtData) RoomResources(db *base.DbTopLevel) {
	tt_data.RoomIndex = map[NodeRef]RoomIndex{}
	for i, r := range db.Rooms {
		tt_data.RoomIndex[r.Id] = i
	}
}

type TtSourceItem = autotimetable.TtSourceItem

func (tt_data *TtData) GetActivities() []TtSourceItem {
	alist := make([]TtSourceItem, len(tt_data.TtActivities))
	for i := range tt_data.TtActivities {
		alist[i] = TtSourceItem{Index: i}
	}
	return alist
}

func (tt_data *TtData) GetClasses() []TtSourceItem {
	clist := make([]TtSourceItem, len(tt_data.ClassDivisions))
	for i, c := range tt_data.ClassDivisions {
		clist[i] = TtSourceItem{Index: i, Tag: c.Class.Tag}
	}
	return clist
}

func (tt_data *TtData) GetConstraint_Types() []autotimetable.ConstraintType {
	return tt_data.ConstraintTypes
}

//TODO: tt_data should have lists of all the indexed things, providing the source ids,
// as these are not directly relevant to autotimetable.
