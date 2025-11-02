// Package db provides data structures for management of a school's data,
// together with some supporting functions and methods.
//
// Initially designed around the data connected with timetabling, it is
// readily extendable. The root element is the [DbTopLevel] struct, which
// is sometimes referred to as the "database".
// No particular method of persistent storage is specified, the data
// structures can be assembled from and saved to any format for which a
// supporting package is defined.
// However, there is basic in-built support for reading from and saving to
// JSON.
// TODO: Currently dealing only with the elements needed for the timetable
package db

import (
	"fetrunner/base"
)

var ErrorMessages = map[string]string{}

// A NodeRef is used to identify the constituent elements of the database.
type NodeRef string // Element Id

// A TimeSlot specifies an activity time period within the school week. The
// school week is divided into days, which are divided into "hours" (activity
// periods), which are usually not 60 minutes in length. Each day has the
// same number of activities.
type TimeSlot struct {
	Day  int // index to [DbTopLevel.Days]
	Hour int // index to [DbTopLevel.Hours]
}

// A Division specifies a particular splitting of a school "class" (the
// students, not the activities) into a number of groups (say, "A" and "B").
//
// In principle, a class may have any number of divisions, each of which
// may have any number of groups, though keeping them to a minimum is
// generally advisable.
//
// Group names must be unique within a class and groups from different
// divisions may not have activities at the same time.
type Division struct {
	Name   string
	Groups []NodeRef
}

// An Info (of which there will only be one instance) collects general
// information which doesn't have its own structure.
type Info struct {
	// Institution can be the name of the school. It may be used in printed
	// output, for example.
	Institution string
	// FirstAfternoonHour is the first "hour" (0-based index) which is to
	// be regarded as "afternoon".
	FirstAfternoonHour int
	// MiddayBreak specifies the "hours" (0-based indexes) which are to be
	// regarded as possible lunch breaks. They should be contiguous.
	MiddayBreak []int
	// Reference can be used to distinguish this particular data set from
	// others. It is not used in the code.
	Reference string
}

type ElementBase struct {
	Id NodeRef
	// Not all elements use the Tag field
	Tag string // abbreviation/acronym
}

type Element interface {
	GetRef() NodeRef
	GetTag() string
	setTag(string)
}

func (e *ElementBase) GetRef() NodeRef {
	return e.Id
}

func (e *ElementBase) GetTag() string {
	return e.Tag
}

func (e *ElementBase) setTag(tag string) {
	e.Tag = tag
}

// A Day represents a day of the timetable's week
type Day struct {
	ElementBase
	Name string
}

// An Hour represents an activity period ("hour") of a timetable's day
type Hour struct {
	ElementBase
	Name  string
	Start string // start time, format hour:mins, e.g. "13:45"
	End   string // end time, format hour:mins, e.g. "14:30"
}

// A Teacher represents a member of staff, including various constraint
// information relevant for the timetable.
// It can be specified as a recourse for an activity.
type Teacher struct {
	ResourceBase
	Name      string
	Firstname string
}

// A Subject represents a taught subject, used for labelling an activitiy, but
// it can also be used for any other activities which are timetabled (say,
// conferences).
type Subject struct {
	ElementBase
	Name string
}

// A Room is a resource which can be specified for an activity.
type Room struct {
	ResourceBase
	Name        string
	Constraints []Constraint
}

// IsReal reports whether r is an actual [Room], rather than a [RoomGroup] or
// [RoomChoiceGroup].
func (r *Room) IsReal() bool {
	return true
}

// A RoomGroup is a collection of [Room] items, all of which are "required".
type RoomGroup struct {
	ElementBase
	Name  string
	Rooms []NodeRef
}

func (r *RoomGroup) IsReal() bool {
	return false
}

// A RoomChoiceGroup is a collection of [Room] items, one of which is
// "required".
type RoomChoiceGroup struct {
	ElementBase
	Name  string
	Rooms []NodeRef
}

func (r *RoomChoiceGroup) IsReal() bool {
	return false
}

// A Class represents a collection of students and will generally correspond
// to a school class (not "lesson"). It includes various constraint
// information relevant for the timetable.
// See type [Group] (representing a subgroup of a class) for the student
// groups which can be specified as a resourse for an activity.
// A class often has a name which consists of a number and a letter or two.
// The number (Year field) represents the class's "year" (A.E. "grade"), the
// Letter field the text part (it can be more than one letter). The Tag field
// is the combination, e.g. "11A". The Name field can be used for a longer
// description of the class.
type Class struct {
	ResourceBase
	Name       string
	Year       int
	Letter     string
	Divisions  []Division
	ClassGroup NodeRef // the Group representing the whole class
}

type Group struct {
	ElementBase
	// These fields do not belong in the JSON object:
	Class *Class `json:"-"`
}

// A Course specifies a collection of resources needed for a set of
// activities ([Activity] elements). The [Subject] field is a sort of label.
type Course struct {
	ElementBase
	Subject  NodeRef
	Groups   []NodeRef // always `Group`: class references use the ClassGroup
	Teachers []NodeRef
	Room     NodeRef // [Room], [RoomGroup] or [RoomChoiceGroup] element
	// These fields do not belong in the JSON object:
	Activities []*Activity `json:"-"`
}

func (c *Course) GetActivityList() []*Activity {
	return c.Activities
}

func (c *Course) SetActivityList(ll []*Activity) {
	c.Activities = ll
}

func (c *Course) IsSuperCourse() bool {
	return false
}

// A SuperCourse specifies a collection of [SubCourse] elements which are
// associated with a set of activities ([Activity] elements). The [Subject]
// field is a sort of label.
type SuperCourse struct {
	ElementBase
	Subject NodeRef
	// These fields do not belong in the JSON object:
	SubCourses []*SubCourse `json:"-"`
	Activities []*Activity  `json:"-"`
}

func (c *SuperCourse) IsSuperCourse() bool {
	return true
}

func (c *SuperCourse) GetActivityList() []*Activity {
	return c.Activities
}

func (c *SuperCourse) SetActivityList(ll []*Activity) {
	c.Activities = ll
}

// A SubCourse has no activities of its own, but shares those of its parent
// [SuperCourse] elements. A SubCourse may blong to more than one
// [SuperCourse]. Otherwise it is much like a [Course], bundling the
// necessary resources.
type SubCourse struct {
	ElementBase
	SuperCourses []NodeRef
	Subject      NodeRef
	Groups       []NodeRef // always `Group`: class references use the ClassGroup
	Teachers     []NodeRef
	Room         NodeRef //  [Room], [RoomGroup] or [RoomChoiceGroup] element
}

// A GeneralRoom covers  [Room], [RoomGroup] and [RoomChoiceGroup].
type GeneralRoom interface {
	IsReal() bool
}

// A Activity is an activity which needs placing in the timetable.
// Its resources are determined by the course ([Course] or [SuperCourse]) to
// which it belongs.
type Activity struct {
	ElementBase
	Course   NodeRef   // [Course] or [SuperCourse] elements
	Duration int       // number of "hours" covered
	Day      int       // 0-based index, -1 for "unplaced"
	Hour     int       // 0-based index
	Fixed    bool      // whether the Activity is unmovable
	Rooms    []NodeRef // actually allocated Room elements
	//Background string // colour, as "#RRGGBB"
	//Footnote   string
}

// ActivityCourse is a type of course which can have activities, i.e. a
// [Course] or a [SuperCourse].
type ActivityCourse interface {
	IsSuperCourse() bool // whether this is a SuperCourse

	// When the data is initially loaded the courses have no attached
	// activities.
	// The activity list is built from the course references in the Activity
	// elements. The individual activities are inserted such that they are
	// ordered with the longest (duration) first. The following functions
	// are used in the building of these lists.
	GetActivityList() []*Activity
	SetActivityList([]*Activity)
}

// Constraint is a rule used in the construction of a timetable.
//
// These can be very varied and they may have very little in common. Each
// implementation must have a distinguishing CType.
type Constraint struct {
	CType    string  // constraint type
	Id       NodeRef // reference to external source
	Weight   int     // range 0 (inactive) - 100 (hard)
	Data     any     // contents depend on CType
	Disabled bool
}

func (c *Constraint) IsHard() bool {
	return c.Weight == MAXWEIGHT
}

// A `DbTopLevel` is the root of a data set.
// In general, the list fields should be ordered, where this is relevant.
type DbTopLevel struct {
	Info Info
	// ModuleData is for data supplied and managed by other packages
	ModuleData       map[string]any
	Days             []*Day
	Hours            []*Hour
	Teachers         []*Teacher
	Subjects         []*Subject
	Rooms            []*Room
	RoomGroups       []*RoomGroup       `json:",omitempty"`
	RoomChoiceGroups []*RoomChoiceGroup `json:",omitempty"`
	Classes          []*Class
	Groups           []*Group       `json:",omitempty"`
	Courses          []*Course      `json:",omitempty"`
	SuperCourses     []*SuperCourse `json:",omitempty"`
	SubCourses       []*SubCourse   `json:",omitempty"`
	Activities       []*Activity    `json:",omitempty"`
	Constraints      []*Constraint  `json:",omitempty"`

	// This field is a convenience structure built from other elements
	// of the `DbTopLevel`. It should not be saved with the rest of the
	// structure.
	Elements map[NodeRef]Element `json:"-"`
}

func (db *DbTopLevel) GetElement(ref NodeRef) Element {
	e, ok := db.Elements[ref]
	if !ok {
		panic("GetElement, unknown Ref: " + ref)
	}
	return e
}

func (db *DbTopLevel) Ref2Tag(ref NodeRef) string {
	e, ok := db.Elements[ref]
	if !ok {
		base.Bug.Fatalf("No Ref2Tag for %s\n", ref)
	}
	return e.GetTag()
}

type ResourceBase struct {
	ElementBase
	Constraints []*Constraint
}

type Resource interface {
	GetResourceTag() string
	addConstraint(*Constraint)
}

func (r *ResourceBase) addConstraint(c *Constraint) {
	r.Constraints = append(r.Constraints, c)
}

func (r *ResourceBase) GetResourceTag() string {
	return r.Tag
}

func (t *Teacher) GetResourceTag() string {
	return t.Tag
}

func (r *Room) GetResourceTag() string {
	return r.Tag
}

var CLASS_GROUP_SEPARATOR string = "."

// The tag for a group within a class is constructed from the class tag,
// CLASS_GROUP_SEPARATOR and the group tag. However, if the group is the
// whole class, just the class tag is used.
func GroupTag(g *Group) string {
	gt := g.Class.Tag
	if g.Tag != "" {
		gt += CLASS_GROUP_SEPARATOR + g.Tag
	}
	return gt
}
