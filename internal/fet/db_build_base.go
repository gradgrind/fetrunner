package fet

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fmt"

	"github.com/beevik/etree"
)

const fet_version = "7.5.5"

type fet_build struct {
	real_soft           bool // if false, give soft constraints weight 100
	no_room_constraints bool // if true, no rooms will be allocated

	//TODO? basedata           *base.BaseData
	ttsource           autotimetable.TtSource
	source_constraints []*autotimetable.TtConstraint

	Doc                *etree.Document
	WeightTable        []float64
	ConstraintElements [][]*etree.Element // a source constraint can have multiple FET constraints
	TimeConstraints    []int              // indexes into `ConstraintElements`
	SpaceConstraints   []int              // indexes into `ConstraintElements`

	/*TODO--?
	Constraints        []constraint
	NConstraints       constraintIndex
	ConstraintTypes    []constraintType
	HardConstraintMap  map[constraintType][]constraintIndex
	SoftConstraintMap  map[constraintType][]constraintIndex
	*/

	fetroot                *etree.Element
	room_list              *etree.Element // needed for adding virtual rooms
	activity_tag_list      *etree.Element // in case these are needed
	time_constraints_list  *etree.Element
	space_constraints_list *etree.Element

	//TODO: What is this for?
	ActivityElementList []*etree.Element

	DayList      []string
	HourList     []string
	ClassList    []string
	TeacherList  []string
	SubjectList  []string
	RoomList     []string
	ActivityList []string

	hard_teacher_blocks [][]base.TimeSlot
	hard_class_blocks   [][]base.TimeSlot

	// Cache for FET virtual rooms, "hash" -> FET-virtual-room tag
	fet_virtual_rooms  map[string]string
	fet_virtual_room_n map[string]int // FET-virtual-room tag -> number of room sets

	//TODO--?
	//real_soft bool // if false, give soft constraints weight 100
}

var db_constraint_fet = map[constraintType]func(
	*fet_build,
	constraintIndex,
	*ttConstraint,
){
	base.C_RoomNotAvailable:      room_blocked,
	base.C_ClassNotAvailable:     class_blocked,
	base.C_TeacherNotAvailable:   teacher_blocked,
	base.C_ActivityStartTime:     activity_start,
	base.C_ActivityRooms:         activity_rooms,
	base.C_TeacherMaxDays:        teacher_max_days,
	base.C_TeacherMinHoursPerDay: teacher_min_hours_per_day,
	base.C_TeacherMaxHoursPerDay: teacher_max_hours_per_day,
	base.C_TeacherLunchBreak:     teacher_lunch_breaks,
	base.C_TeacherMaxGapsPerDay:  teacher_max_gaps_per_day,
	base.C_ClassMaxGapsPerWeek:   class_max_gaps_per_week,
	base.C_ClassMinHoursPerDay:   class_min_hours_per_day,
	base.C_ClassMaxHoursPerDay:   class_max_hours_per_day,
	base.C_ClassLunchBreak:       class_lunch_breaks,
	base.C_ClassMaxGapsPerDay:    class_max_gaps_per_day,
	base.C_ClassMaxGapsPerWeek:   class_max_gaps_per_week,
	base.C_ClassForceFirstHour:   class_force_first_hour,
}

func anyInt(a any) int {
	val, ok := a.(int)
	if ok {
		return val
	}
	panic("Expected int value")
}

func mapReadInt(m any, key string) int {
	mm, ok := m.(map[string]any)
	if ok {
		val, ok := mm[key].(int)
		if ok {
			return val
		}
	}
	panic("Expected map, key: " + key + ", int value")
}

func mapReadIndexList(m any, key string) []int {
	mm, ok := m.(map[string]any)
	if ok {
		val, ok := mm[key].([]int)
		if ok {
			return val
		}
	}
	panic("Expected map, key: " + key + ", int-list value")
}

func mapReadIndexListList(m any, key string) [][]int {
	mm, ok := m.(map[string]any)
	if ok {
		val, ok := mm[key].([][]int)
		if ok {
			return val
		}
	}
	panic("Expected map, key: " + key + ", int-list-list value")
}

func mapReadTimeSlot(m any) base.TimeSlot {
	mm, ok := m.(map[string]any)
	if ok {
		val, ok := mm["Time"].(base.TimeSlot)
		if ok {
			return val
		}
	}
	panic("Expected map, key: Time, a TimeSlot as value")
}

func mapReadTimeSlots(m any) []base.TimeSlot {
	mm, ok := m.(map[string]any)
	if ok {
		val, ok := mm["Times"].([]base.TimeSlot)
		if ok {
			return val
		}
	}
	panic("Expected map, key: Times, list of TimeSlots as value")
}

// Return FET weight and comment values.
func (fetbuild *fet_build) constraintWeight(i int, w int) (string, string) {
	w1 := "100"
	if w != base.MAXWEIGHT {
		if fetbuild.real_soft {
			w1 = fetbuild.DbWeight2Fet(w)
		}
		return w1, fmt.Sprintf("[%d:%d]", i, w)
	} else {
		return w1, fmt.Sprintf("[%d]", i)
	}
}
