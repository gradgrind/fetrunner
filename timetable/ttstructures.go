// Package timetable provides structures and associated functions for
// supporting the management of timetables.
package timetable

import (
	"fetrunner/db"
)

type NodeRef = db.NodeRef // node reference (UUID)

type ActivityIndex int
type TeacherIndex int
type RoomIndex int
type AtomicIndex int

type TtData struct {
	Db           *db.DbTopLevel
	NDays        int
	NHours       int
	HoursPerWeek int

	//TODO??
	AtomicGroups []*AtomicGroup

	TeacherIndex map[NodeRef]TeacherIndex
	RoomIndex    map[NodeRef]RoomIndex

	// `AtomicGroupIndex` maps a class or group NodeRef to its list of atomic
	// group indexes.
	AtomicGroupIndex map[NodeRef][]AtomicIndex

	// `ClassDivisions` is a list with an entry for each class, containing a
	// list of its divisions ([][]NodeRef).
	ClassDivisions []ClassDivision

	// Set up by `CollectCourses`, which calls `makeActivities`
	// Note that activity 0 is invalid, the first activity has index 1.
	Activities     []*TtActivity
	CourseInfoList []*CourseInfo
	Ref2CourseInfo map[NodeRef]*CourseInfo

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
	Groups       []*db.Group // a `Class` is represented by its ClassGroup
	AtomicGroups []AtomicIndex
	Teachers     []TeacherIndex
	FixedRooms   []RoomIndex
	RoomChoices  [][]RoomIndex
	TtActivities []ActivityIndex
}

type TtActivity struct {
	CourseInfo *CourseInfo
	Activity   *db.Activity
	Day        int
	Hour       int
	Fixed      bool
}

type ClassDivision struct {
	Class     *db.Class
	Divisions [][]NodeRef
}

// BasicSetup performs the initialization of a TtSharedData structure, collecting
// "resources" (atomic student groups, teachers and rooms) and "activities".
func BasicSetup(db *db.DbTopLevel) *TtData {
	days := len(db.Days)
	hours := len(db.Hours)
	tt_shared_data := &TtData{
		Db:           db,
		NDays:        days,
		NHours:       hours,
		HoursPerWeek: days * hours,
	}

	// Collect ClassDivisions
	tt_shared_data.FilterDivisions()

	// Atomic groups: an atomic group is a "resource", it is an ordered list
	// of single groups, one from each division.
	// The atomic groups take the lowest resource indexes (starting at 0).
	// `AtomicGroups` maps the classes and groups to a list of their resource
	// indexes.
	tt_shared_data.MakeAtomicGroups()

	// Add teachers and rooms to resource array
	tt_shared_data.TeacherResources()
	tt_shared_data.RoomResources()

	// Get the courses (-> CourseInfo) and activities for the timetable
	tt_shared_data.CollectCourses()

	//TODO: Perhaps this should be called from the back-end, in preparation
	// for a generator run?
	tt_shared_data.preprocessConstraints()

	return tt_shared_data
}

func (tt_data *TtData) TeacherResources() {
	tt_data.TeacherIndex = map[NodeRef]TeacherIndex{}
	for i, t := range tt_data.Db.Teachers {
		tt_data.TeacherIndex[t.Id] = TeacherIndex(i)
	}
}

func (tt_data *TtData) RoomResources() {
	tt_data.RoomIndex = map[NodeRef]RoomIndex{}
	for i, r := range tt_data.Db.Rooms {
		tt_data.RoomIndex[r.Id] = RoomIndex(i)
	}
}

type MinDaysBetweenActivities struct {
	// Result of processing constraints DifferentDays and DaysBetween
	Weight               int
	ConsecutiveIfSameDay bool
	Activities           []ActivityIndex
	MinDays              int
}

type ParallelActivities struct {
	Weight         int
	ActivityGroups [][]ActivityIndex
}

// This structure is used to return the placement results from the
// timetable back-end.
type ActivityPlacement struct {
	Id    ActivityIndex
	Day   int
	Hour  int
	Rooms []RoomIndex
}
