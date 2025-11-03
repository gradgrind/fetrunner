package w365tt

import (
	"encoding/json"
	"fetrunner/db"
)

// The structures used for the "database", adapted to read from W365

type NodeRef = db.NodeRef // Element reference

type Info struct {
	Institution        string
	FirstAfternoonHour int
	MiddayBreak        []int
	Reference          string `json:"Scenario"`
}

type Day struct {
	Id   NodeRef
	Type string
	Name string
	Tag  string `json:"Shortcut"`
}

type Hour struct {
	Id    NodeRef
	Type  string
	Name  string
	Tag   string `json:"Shortcut"`
	Start string
	End   string
}

type TimeSlot struct {
	Day  int
	Hour int
}

type Teacher struct {
	Id               NodeRef
	Type             string
	Name             string
	Tag              string `json:"Shortcut"`
	Firstname        string
	NotAvailable     []TimeSlot `json:"Absences"`
	MinLessonsPerDay int
	MaxLessonsPerDay int
	MaxDays          int
	MaxGapsPerDay    int
	MaxGapsPerWeek   int
	MaxAfternoons    int
	LunchBreak       bool
}

func (t *Teacher) UnmarshalJSON(data []byte) error {
	// Customize defaults for Teacher
	t.MinLessonsPerDay = -1
	t.MaxLessonsPerDay = -1
	t.MaxDays = -1
	t.MaxGapsPerDay = -1
	t.MaxGapsPerWeek = -1
	t.MaxAfternoons = -1

	type tempT Teacher
	return json.Unmarshal(data, (*tempT)(t))
}

type Subject struct {
	Id   NodeRef
	Type string
	Name string
	Tag  string `json:"Shortcut"`
}

type Room struct {
	Id           NodeRef
	Type         string
	Name         string
	Tag          string     `json:"Shortcut"`
	NotAvailable []TimeSlot `json:"Absences"`
}

type RoomGroup struct {
	Id    NodeRef
	Type  string
	Name  string
	Tag   string `json:"Shortcut"`
	Rooms []NodeRef
}

type Class struct {
	Id               NodeRef
	Type             string
	Name             string
	Tag              string `json:"Shortcut"`
	Year             int    `json:"Level"`
	Letter           string
	NotAvailable     []TimeSlot `json:"Absences"`
	Divisions        []Division
	MinLessonsPerDay int
	MaxLessonsPerDay int
	MaxGapsPerDay    int
	MaxGapsPerWeek   int
	MaxAfternoons    int
	LunchBreak       bool
	ForceFirstHour   bool
}

func (c *Class) UnmarshalJSON(data []byte) error {
	// Customize defaults for Class
	c.MinLessonsPerDay = -1
	c.MaxLessonsPerDay = -1
	c.MaxGapsPerDay = -1
	c.MaxGapsPerWeek = -1
	c.MaxAfternoons = -1

	type tempC Class
	return json.Unmarshal(data, (*tempC)(c))
}

type Group struct {
	Id   NodeRef
	Type string
	Tag  string `json:"Shortcut"`
}

type Division struct {
	Id     NodeRef
	Type   string
	Name   string
	Groups []NodeRef
}

type Course struct {
	Id             NodeRef
	Type           string
	Subjects       []NodeRef
	Groups         []NodeRef // can be `Class` or `Group`
	Teachers       []NodeRef
	PreferredRooms []NodeRef
}

type SuperCourse struct {
	Id         NodeRef
	Type       string
	EpochPlan  NodeRef
	SubCourses []SubCourse
}

type SubCourse struct {
	Id             NodeRef
	Type           string
	Subjects       []NodeRef
	Groups         []NodeRef // can be `Class` or `Group`
	Teachers       []NodeRef
	PreferredRooms []NodeRef
}

type Lesson struct {
	Id       NodeRef
	Type     string
	Course   NodeRef // Course or SuperCourse Elements
	Duration int
	Day      int
	Hour     int
	Fixed    bool
	Rooms    []NodeRef `json:"LocalRooms"` // only Room Elements
	//Flags      []string
	//Background string
	//Footnote   string
}

type EpochPlan struct {
	Id   NodeRef
	Type string
	Tag  string `json:"Shortcut"`
	Name string
}

type W365TopLevel struct {
	Info Info `json:"W365TT"`
	//PrintTables  []*ttprint.PrintTable
	FetData      map[string]string
	Days         []*Day
	Hours        []*Hour
	Teachers     []*Teacher
	Subjects     []*Subject
	Rooms        []*Room
	RoomGroups   []*RoomGroup
	Classes      []*Class
	Groups       []*Group
	Courses      []*Course
	SuperCourses []*SuperCourse
	Lessons      []*Lesson
	EpochPlans   []*EpochPlan
	Constraints  []map[string]any

	// These fields do not belong in the JSON object.
	RealRooms       map[NodeRef]*db.Room      `json:"-"`
	RoomGroupMap    map[NodeRef]*db.RoomGroup `json:"-"`
	SubjectMap      map[NodeRef]*db.Subject   `json:"-"`
	GroupRefMap     map[NodeRef]NodeRef       `json:"-"`
	TeacherMap      map[NodeRef]struct{}      `json:"-"`
	CourseMap       map[NodeRef]struct{}      `json:"-"`
	SubjectTags     map[string]NodeRef        `json:"-"`
	RoomTags        map[string]NodeRef        `json:"-"`
	RoomChoiceNames map[string]NodeRef        `json:"-"`
}

// Block all afternoons if nAfternnons == 0.
func (dbp *W365TopLevel) handleZeroAfternoons(
	notAvailable []TimeSlot,
	nAfternoons int,
) []db.TimeSlot {
	// Make a bool array and fill this in two passes, then remake list
	namap := make([][]bool, len(dbp.Days))
	nhours := len(dbp.Hours)
	// In the first pass, conditionally block afternoons
	for i := range namap {
		namap[i] = make([]bool, nhours)
		if nAfternoons == 0 {
			for h := dbp.Info.FirstAfternoonHour; h < nhours; h++ {
				namap[i][h] = true
			}
		}
	}
	// In the second pass, include existing blocked hours.
	for _, ts := range notAvailable {
		if ts.Hour < len(dbp.Hours) {
			// Exclude invalid hours
			namap[ts.Day][ts.Hour] = true
		}
		//TODO: else an error message?
	}
	// Build a new base.TimeSlot list
	na := []db.TimeSlot{}
	for d, naday := range namap {
		for h, nahour := range naday {
			if nahour {
				na = append(na, db.TimeSlot{Day: d, Hour: h})
			}
		}
	}
	return na
}
