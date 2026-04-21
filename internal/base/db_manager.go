package base

import (
	"fmt"
	"slices"
	"strconv"

	"github.com/gofrs/uuid/v5"
)

func NewDb() *DbTopLevel {
	return &DbTopLevel{
		Placements:  map[string][]*ActivityPlacement{},
		Constraints: map[string][]*BaseConstraint{},
		ElementMap:  map[NodeRef]Element{},
	}
}

func NewId() NodeRef {
	// Create a Version 4 UUID.
	u2, err := uuid.NewV4()
	if err != nil {
		panic(err)
	}
	return NodeRef(u2.String())
}

func addElement(ref NodeRef, element Element) NodeRef {
	if ref == "" {
		ref = NewId()
	} else {
		_, known := DataBase.Db.ElementMap[ref]
		if known {
			LogError("--ELEMENT_ID_DEFINED_MORE_THAN_ONCE %s", ref)
			ref = NewId()
		}
	}
	DataBase.Db.ElementMap[ref] = element
	return ref
}

func NewDay(ref NodeRef) *Day {
	e := &Day{}
	e.Id = addElement(ref, e)
	DataBase.Db.Days = append(DataBase.Db.Days, e)
	return e
}

func NewHour(ref NodeRef) *Hour {
	e := &Hour{}
	e.Id = addElement(ref, e)
	DataBase.Db.Hours = append(DataBase.Db.Hours, e)
	return e
}

func NewTeacher(ref NodeRef) *Teacher {
	e := &Teacher{}
	e.Id = addElement(ref, e)
	DataBase.Db.Teachers = append(DataBase.Db.Teachers, e)
	return e
}

func NewSubject(ref NodeRef) *Subject {
	e := &Subject{}
	e.Id = addElement(ref, e)
	DataBase.Db.Subjects = append(DataBase.Db.Subjects, e)
	return e
}

func NewRoom(ref NodeRef) *Room {
	e := &Room{}
	e.Id = addElement(ref, e)
	DataBase.Db.Rooms = append(DataBase.Db.Rooms, e)
	return e
}

func NewRoomGroup(ref NodeRef) *RoomGroup {
	e := &RoomGroup{}
	e.Id = addElement(ref, e)
	DataBase.Db.RoomGroups = append(DataBase.Db.RoomGroups, e)
	return e
}

func NewRoomChoiceGroup(ref NodeRef) *RoomChoiceGroup {
	e := &RoomChoiceGroup{}
	e.Id = addElement(ref, e)
	DataBase.Db.RoomChoiceGroups = append(DataBase.Db.RoomChoiceGroups, e)
	return e
}

func NewClass(ref NodeRef) *Class {
	e := &Class{}
	e.Id = addElement(ref, e)
	DataBase.Db.Classes = append(DataBase.Db.Classes, e)
	return e
}

func NewGroup(ref NodeRef) *Group {
	e := &Group{}
	e.Id = addElement(ref, e)
	DataBase.Db.Groups = append(DataBase.Db.Groups, e)
	return e
}

func NewCourse(ref NodeRef) *Course {
	e := &Course{}
	e.Id = addElement(ref, e)
	DataBase.Db.Courses = append(DataBase.Db.Courses, e)
	return e
}

func NewSuperCourse(ref NodeRef) *SuperCourse {
	e := &SuperCourse{}
	e.Id = addElement(ref, e)
	DataBase.Db.SuperCourses = append(DataBase.Db.SuperCourses, e)
	return e
}

func NewSubCourse(ref NodeRef) *SubCourse {
	e := &SubCourse{}
	e.Id = addElement(ref, e)
	DataBase.Db.SubCourses = append(DataBase.Db.SubCourses, e)
	return e
}

func NewActivity(ref NodeRef) *Activity {
	e := &Activity{}
	e.Id = addElement(ref, e)
	DataBase.Db.Activities = append(DataBase.Db.Activities, e)
	return e
}

// `PrepareDb` must be called after the data has been initially loaded into
// the `DbTopLevel` structure. It processes the data by performing checks and
// completing the initialization of the internal data structures.
func PrepareDb() {
	db := DataBase.Db
	// Collect the SubCourses for each SuperCourse
	for _, sbc := range db.SubCourses {
		for _, spcref := range sbc.SuperCourses {
			spc := db.ElementMap[spcref].(*SuperCourse)
			spc.SubCourses = append(spc.SubCourses, sbc)
		}
	}

	// Collect the Activities for each Course and SuperCourse, the list being
	// ordered with the longest durations first
	for _, l := range db.Activities {
		c := db.ElementMap[l.Course].(ActivityCourse)
		d := l.Duration
		var i int = 0
		ll := c.GetActivityList()
		for _, a := range ll {
			if a.Duration <= d {
				break
			}
			i++
		}
		ll = slices.Insert(ll, i, l)
		c.SetActivityList(ll)
	}

	// Expand Group information
	for _, c := range db.Classes {
		if c.ClassGroup == "" {
			// Not a real class
			continue
		}
		db.ElementMap[c.ClassGroup].(*Group).Class = c // Tag is empty.
		for _, d := range c.Divisions {
			for _, gref := range d.Groups {
				db.ElementMap[gref].(*Group).Class = c
			}
		}
	}
	// Check that all groups belong to a class
	for _, g := range db.Groups {
		if g.Class == nil {
			// This is a loader failure, it should not be possible.
			panic(fmt.Sprintf("Group not in Class: %s", g.Id))
		}
	}

	// Check that element tags are unique
	newtags("Subject", db.Subjects)
	newtags("Room", db.Rooms)
	newtags("Teacher", db.Teachers)

	// Check that the Rooms in RoomGroups and RoomChoiceGroups are valid.
	for _, rg := range db.RoomGroups {
		rlist := []NodeRef{}
		for _, r := range rg.Rooms {
			if _, ok := db.ElementMap[r].(*Room); ok {
				rlist = append(rlist, r)
			} else {
				LogError("--INVALID_ROOM_IN_ROOM_GROUP Room: %s, RoomGroup: %s", r, rg.Tag)
			}
		}
		rg.Rooms = rlist
	}
	for _, rg := range db.RoomChoiceGroups {
		rlist := []NodeRef{}
		for _, r := range rg.Rooms {
			if _, ok := db.ElementMap[r].(*Room); ok {
				rlist = append(rlist, r)
			} else {
				LogError("--INVALID_ROOM_IN_ROOM_CHOICE_GROUP Room: %s, RoomChoiceGroup: %s", r, rg.Tag)
			}
		}
		rg.Rooms = rlist
	}
}

func newtags[T Element](etype string, elist []T) {
	checktags := map[string]bool{}
	errortags := []Element{}
	for _, e0 := range elist {
		tag := e0.GetTag()
		if checktags[tag] {
			errortags = append(errortags, e0)
		} else {
			checktags[tag] = true
		}
	}
	for _, e := range errortags {
		tag0 := e.GetTag()
		i := 1
		var tag string
		for {
			tag = tag0 + strconv.Itoa(i)
			_, nok := checktags[tag]
			if !nok {
				break
			}
		}
		checktags[tag] = true
		e.setTag(tag)
		LogError(
			"--ELEMENT_TAG_NOT_UNIQUE Type: %s, Tag: %s, Element %s -> %s",
			etype, tag0, e.GetRef(), tag)
	}
}

func CheckDbBasics() bool {
	db := DataBase.Db
	// This function is provided for use by code which needs the following
	// Elements to be provided.
	if len(db.Days) == 0 {
		LogError("--NO_DAYS")
		return false
	}
	if len(db.Hours) == 0 {
		LogError("--NO_HOURS")
		return false
	}
	if len(db.Teachers) == 0 {
		LogError("--NO_TEACHERS")
		return false
	}
	if len(db.Subjects) == 0 {
		LogError("--NO_SUBJECTS")
		return false
	}
	if len(db.Rooms) == 0 {
		LogError("--NO_ROOMS")
		return false
	}
	if len(db.Classes) == 0 {
		LogError("--NO_CLASSES")
		return false
	}
	return true
}

// Interface for Course and SubCourse elements
type CourseInterface interface {
	GetId() NodeRef
	GetGroups() []NodeRef
	GetTeachers() []NodeRef
	GetSubject() NodeRef
	GetRoom() NodeRef
}

func (c *Course) GetId() NodeRef            { return c.Id }
func (c *SubCourse) GetId() NodeRef         { return c.Id }
func (c *Course) GetGroups() []NodeRef      { return c.Groups }
func (c *SubCourse) GetGroups() []NodeRef   { return c.Groups }
func (c *Course) GetTeachers() []NodeRef    { return c.Teachers }
func (c *SubCourse) GetTeachers() []NodeRef { return c.Teachers }
func (c *Course) GetSubject() NodeRef       { return c.Subject }
func (c *SubCourse) GetSubject() NodeRef    { return c.Subject }
func (c *Course) GetRoom() NodeRef          { return c.Room }
func (c *SubCourse) GetRoom() NodeRef       { return c.Room }
