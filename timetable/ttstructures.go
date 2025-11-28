// Package timetable provides structures and associated functions for
// supporting the management of timetables.
package timetable

import (
	"fetrunner/base"
)

type NodeRef = base.NodeRef // node reference (UUID)

type ActivityIndex = int
type TeacherIndex = int
type RoomIndex = int
type ClassIndex = int
type AtomicIndex = int

type TtData struct {
	BaseData     *base.BaseData
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

	// Set up by `CollectCourses`, which calls `makeActivities`
	// Note that activity 0 is invalid, the first activity has index 1.
	Activities        []*TtActivity
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

type TtActivity struct {
	CourseInfo     int            // index to `TtData.CourseInfoList`
	FixedStartTime *base.TimeSlot // needed for days-between preparation
}

type ClassDivision struct {
	Class     *base.Class
	Divisions [][]NodeRef
}

// BasicSetup performs the initialization of a TtData structure, collecting
// "resources" (atomic student groups, teachers and rooms) and "activities".
func BasicSetup(bd *base.BaseData) *TtData {
	days := len(bd.Db.Days)
	hours := len(bd.Db.Hours)
	tt_data := &TtData{
		BaseData:     bd,
		NDays:        days,
		NHours:       hours,
		HoursPerWeek: days * hours,
	}

	// Collect ClassDivisions
	tt_data.FilterDivisions()

	// Atomic groups: an atomic group is a "resource", it is an ordered list
	// of single groups, one from each division.
	// The atomic groups take the lowest resource indexes (starting at 0).
	// `AtomicGroups` maps the classes and groups to a list of their resource
	// indexes.
	tt_data.MakeAtomicGroups()

	// Add teachers and rooms to resource array
	tt_data.TeacherResources()
	tt_data.RoomResources()

	// Get the courses (-> CourseInfo) and activities for the timetable
	tt_data.CollectCourses()

	//TODO: Perhaps this should be called from the back-end, in preparation
	// for a generator run?
	tt_data.preprocessConstraints()

	return tt_data
}

func (tt_data *TtData) TeacherResources() {
	tt_data.TeacherIndex = map[NodeRef]TeacherIndex{}
	for i, t := range tt_data.BaseData.Db.Teachers {
		tt_data.TeacherIndex[t.Id] = i
	}
}

func (tt_data *TtData) RoomResources() {
	tt_data.RoomIndex = map[NodeRef]RoomIndex{}
	for i, r := range tt_data.BaseData.Db.Rooms {
		tt_data.RoomIndex[r.Id] = i
	}
}

// This structure is used to return the placement results from the
// timetable back-end. It differs from `base.ActivityPlacement` in that it
// uses indexes rather than NodeRefs.
type TtActivityPlacement struct {
	Activity ActivityIndex
	Day      int
	Hour     int
	Rooms    []RoomIndex
}
