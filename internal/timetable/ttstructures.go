// Package timetable provides structures and associated functions for
// supporting the management of timetables.
package timetable

import (
	"cmp"
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"maps"
	"slices"
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

type ttActivity = autotimetable.TtActivity
type ttConstraint = autotimetable.TtConstraint
type constraintIndex = autotimetable.ConstraintIndex
type autoTtData = autotimetable.AutoTtData
type constraintType = autotimetable.ConstraintType

type TtData struct {
	db     *base.DbTopLevel
	ndays  int
	nhours int

	constraints   []*ttConstraint // ordered constraint list for "autotimetable"
	teacher2Index map[nodeRef]teacherIndex
	teachers      []element
	room2Index    map[nodeRef]roomIndex
	rooms         []element
	class2Index   map[nodeRef]classIndex
	// `atomicGroup2Indexes` maps a class or group NodeRef to its list of atomic
	// group indexes.
	atomicGroup2Indexes map[nodeRef][]atomicIndex
	atomicGroups        []*AtomicGroup

	// `classDivisions` is a list with an entry for each class, containing a
	// list of its divisions ([][]NodeRef).
	classDivisions []classDivision

	// Set up by `CollectCourses`, which calls `makeActivities`
	ttActivities    []*ttActivity
	fixedActivities []*base.TimeSlot

	ref2ActivityIndex map[nodeRef]activityIndex
	courseInfoList    []*courseInfo
	ref2courseInfo    map[nodeRef]*courseInfo
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
	clist := make([]*ttClass, len(tt_data.classDivisions))
	for i, c := range tt_data.classDivisions {
		glist := []*ttGroup{}
		for _, d := range c.Divisions {
			for _, g := range d {
				e := db.GetElement(g)
				id := e.GetRef()
				glist = append(glist, &ttGroup{
					Id:            id,
					Tag:           e.GetTag(),
					ClassIndex:    i,
					AtomicIndexes: tt_data.atomicGroup2Indexes[id],
				})
			}
		}
		clist[i] = &ttClass{
			Id:            c.Class.Id,
			Tag:           c.Class.Tag,
			AtomicIndexes: tt_data.atomicGroup2Indexes[c.Class.Id],
			Groups:        glist,
		}
	}
	return clist
}

func (tt_data *TtData) GetAtomicGroups() []string {
	aglist := []string{}
	for _, ag := range tt_data.atomicGroups {
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

// A `courseInfo` is a representation of a course (Course or SuperCourse) for
// the timetable.
// Activities within a course are (already) ordered, highest duration first,
// and the Activities field has the same order.
type courseInfo struct {
	Id                 nodeRef // Course or SuperCourse
	Subject            string
	Groups             []*base.Group // a `Class` is represented by its ClassGroup
	AtomicGroupIndexes []atomicIndex
	Teachers           []teacherIndex
	FixedRooms         []roomIndex
	RoomChoices        [][]roomIndex
	Activities         []activityIndex
}

type classDivision struct {
	Class     *base.Class
	Divisions [][]nodeRef
}

// MakeTimetableData performs the initialization of a TtData structure, collecting
// "resources" (atomic student groups, teachers and rooms) and "activities".
func MakeTimetableData(bd *base.BaseData) autotimetable.TtSource {
	db := bd.Db
	tt_data := &TtData{
		db: db,

		ndays:  len(db.Days),
		nhours: len(db.Hours),
	}

	// Collect ClassDivisions
	tt_data.FilterDivisions()

	// Atomic groups: an atomic group is a "resource", it is an ordered list
	// of single groups, one from each division.
	tt_data.MakeAtomicGroups()

	// Build maps to convert teacher and room ids to indexes
	tt_data.TeacherResources()
	tt_data.RoomResources()

	// Get the courses (-> courseInfo) and activities for the timetable
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
	tt_data.activity_constraints(constraint_map)
	tt_data.class_constraints(constraint_map)
	tt_data.teacher_constraints(constraint_map)
	for c := range constraint_map {
		bd.Logger.Error("UnhandledConstraintType: %s", c)
	}

	return tt_data
}

func (tt_data *TtData) TeacherResources() {
	tt_data.teacher2Index = map[nodeRef]teacherIndex{}
	tt_data.teachers = make([]element, len(tt_data.db.Teachers))
	for i, t := range tt_data.db.Teachers {
		tt_data.teacher2Index[t.Id] = i
		tt_data.teachers[i] = element{Id: t.Id, Tag: t.Tag}
	}
}

func (tt_data *TtData) RoomResources() {
	tt_data.room2Index = map[nodeRef]roomIndex{}
	tt_data.rooms = make([]element, len(tt_data.db.Rooms))
	for i, r := range tt_data.db.Rooms {
		tt_data.room2Index[r.Id] = i
		tt_data.rooms[i] = element{Id: r.Id, Tag: r.Tag}
	}
}

func (tt_data *TtData) GetActivities() []*ttActivity {
	return tt_data.ttActivities
}

func (tt_data *TtData) GetConstraintTypes() []constraintType {
	// First sort alphabetically, remove duplicates, and then sort according
	// to priority. This allows constraint types with the same priority to
	// have a stable order.
	ctlist := make([]constraintType, len(tt_data.constraints))
	for i, c := range tt_data.constraints {
		ctlist[i] = c.CType
	}
	slices.Sort(ctlist)
	ctlist = slices.Compact(ctlist)
	priority := base.ConstraintPriority
	slices.SortStableFunc(ctlist,
		func(a, b constraintType) int {
			return cmp.Compare(priority[b], priority[a])
		})
	return ctlist
}

func (tt_data *TtData) GetConstraintMaps() (
	map[constraintType][]constraintIndex,
	map[constraintType][]constraintIndex,
) {
	hardConstraintMap := map[constraintType][]constraintIndex{}
	softConstraintMap := map[constraintType][]constraintIndex{}
	for i, c := range tt_data.constraints {
		if c.IsHard() {
			hardConstraintMap[c.CType] = append(hardConstraintMap[c.CType], i)
		} else {
			softConstraintMap[c.CType] = append(softConstraintMap[c.CType], i)
		}
	}
	return hardConstraintMap, softConstraintMap
}

func (tt_data *TtData) GetConstraints() []*ttConstraint {
	return tt_data.constraints
}
