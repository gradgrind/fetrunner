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
        Constraints: map[string][]*Constraint{},
        Elements:    map[NodeRef]Element{},
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

func (bd *BaseData) addElement(ref NodeRef, element Element) NodeRef {
    if ref == "" {
        ref = NewId()
    } else {
        _, known := bd.Db.Elements[ref]
        if known {
            bd.Logger.Error("Element Id defined more than once:  %s", ref)
            ref = NewId()
        }
    }
    bd.Db.Elements[ref] = element
    return ref
}

func (bd *BaseData) NewDay(ref NodeRef) *Day {
    e := &Day{}
    e.Id = bd.addElement(ref, e)
    bd.Db.Days = append(bd.Db.Days, e)
    return e
}

func (bd *BaseData) NewHour(ref NodeRef) *Hour {
    e := &Hour{}
    e.Id = bd.addElement(ref, e)
    bd.Db.Hours = append(bd.Db.Hours, e)
    return e
}

func (bd *BaseData) NewTeacher(ref NodeRef) *Teacher {
    e := &Teacher{}
    e.Id = bd.addElement(ref, e)
    bd.Db.Teachers = append(bd.Db.Teachers, e)
    return e
}

func (bd *BaseData) NewSubject(ref NodeRef) *Subject {
    e := &Subject{}
    e.Id = bd.addElement(ref, e)
    bd.Db.Subjects = append(bd.Db.Subjects, e)
    return e
}

func (bd *BaseData) NewRoom(ref NodeRef) *Room {
    e := &Room{}
    e.Id = bd.addElement(ref, e)
    bd.Db.Rooms = append(bd.Db.Rooms, e)
    return e
}

func (bd *BaseData) NewRoomGroup(ref NodeRef) *RoomGroup {
    e := &RoomGroup{}
    e.Id = bd.addElement(ref, e)
    bd.Db.RoomGroups = append(bd.Db.RoomGroups, e)
    return e
}

func (bd *BaseData) NewRoomChoiceGroup(ref NodeRef) *RoomChoiceGroup {
    e := &RoomChoiceGroup{}
    e.Id = bd.addElement(ref, e)
    bd.Db.RoomChoiceGroups = append(bd.Db.RoomChoiceGroups, e)
    return e
}

func (bd *BaseData) NewClass(ref NodeRef) *Class {
    e := &Class{}
    e.Id = bd.addElement(ref, e)
    bd.Db.Classes = append(bd.Db.Classes, e)
    return e
}

func (bd *BaseData) NewGroup(ref NodeRef) *Group {
    e := &Group{}
    e.Id = bd.addElement(ref, e)
    bd.Db.Groups = append(bd.Db.Groups, e)
    return e
}

func (bd *BaseData) NewCourse(ref NodeRef) *Course {
    e := &Course{}
    e.Id = bd.addElement(ref, e)
    bd.Db.Courses = append(bd.Db.Courses, e)
    return e
}

func (bd *BaseData) NewSuperCourse(ref NodeRef) *SuperCourse {
    e := &SuperCourse{}
    e.Id = bd.addElement(ref, e)
    bd.Db.SuperCourses = append(bd.Db.SuperCourses, e)
    return e
}

func (bd *BaseData) NewSubCourse(ref NodeRef) *SubCourse {
    e := &SubCourse{}
    e.Id = bd.addElement(ref, e)
    bd.Db.SubCourses = append(bd.Db.SubCourses, e)
    return e
}

func (bd *BaseData) NewActivity(ref NodeRef) *Activity {
    e := &Activity{}
    e.Id = bd.addElement(ref, e)
    bd.Db.Activities = append(bd.Db.Activities, e)
    return e
}

// `PrepareDb` must be called after the data has been initially loaded into
// the `DbTopLevel` structure. It processes the data by performing checks and
// completing the initialization of the internal data structures.
func (bd *BaseData) PrepareDb() {
    db := bd.Db
    logger := bd.Logger
    if db.Info.MiddayBreak == nil {
        db.Info.MiddayBreak = []int{}
    } else if len(db.Info.MiddayBreak) > 1 {
        // Sort and check contiguity.
        slices.Sort(db.Info.MiddayBreak)
        mb := db.Info.MiddayBreak
        if mb[len(mb)-1]-mb[0] >= len(mb) {
            logger.Error("MiddayBreak hours not contiguous")
            db.Info.MiddayBreak = []int{}
        }
    }

    // Collect the SubCourses for each SuperCourse
    for _, sbc := range db.SubCourses {
        for _, spcref := range sbc.SuperCourses {
            spc := db.Elements[spcref].(*SuperCourse)
            spc.SubCourses = append(spc.SubCourses, sbc)
        }
    }

    // Collect the Activities for each Course and SuperCourse, the list being
    // ordered with the longest durations first
    for _, l := range db.Activities {
        c := db.Elements[l.Course].(ActivityCourse)
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
        db.Elements[c.ClassGroup].(*Group).Class = c // Tag is empty.
        for _, d := range c.Divisions {
            for _, gref := range d.Groups {
                db.Elements[gref].(*Group).Class = c
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
    newtags(logger, "Subject", db.Subjects)
    newtags(logger, "Room", db.Rooms)
    newtags(logger, "Teacher", db.Teachers)

    // Check that the Rooms in RoomGroups and RoomChoiceGroups are valid.
    for _, rg := range db.RoomGroups {
        rlist := []NodeRef{}
        for _, r := range rg.Rooms {
            if _, ok := db.Elements[r].(*Room); ok {
                rlist = append(rlist, r)
            } else {
                logger.Error(
                    "Invalid Room (%s) in RoomGroup %s", r, rg.Tag)
            }
        }
        rg.Rooms = rlist
    }
    for _, rg := range db.RoomChoiceGroups {
        rlist := []NodeRef{}
        for _, r := range rg.Rooms {
            if _, ok := db.Elements[r].(*Room); ok {
                rlist = append(rlist, r)
            } else {
                logger.Error(
                    "Invalid Room (%s) in RoomChoiceGroup %s", r, rg.Tag)
            }
        }
        rg.Rooms = rlist
    }
}

func newtags[T Element](logger *Logger, etype string, elist []T) {
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
        logger.Error(
            "%s tag <%s> not unique: Element %s changed to <%s>",
            etype, tag0, e.GetRef(), tag)
    }
}

func (bd *BaseData) CheckDbBasics() bool {
    db := bd.Db
    logger := bd.Logger
    // This function is provided for use by code which needs the following
    // Elements to be provided.
    if len(db.Days) == 0 {
        logger.Error("No Days")
        return false
    }
    if len(db.Hours) == 0 {
        logger.Error("No Hours")
        return false
    }
    if len(db.Teachers) == 0 {
        logger.Error("No Teachers")
        return false
    }
    if len(db.Subjects) == 0 {
        logger.Error("No Subjects")
        return false
    }
    if len(db.Rooms) == 0 {
        logger.Error("No Rooms")
        return false
    }
    if len(db.Classes) == 0 {
        logger.Error("No Classes")
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
