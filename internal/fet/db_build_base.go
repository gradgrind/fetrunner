package fet

import (
	"fetrunner/internal/base"
	"fetrunner/internal/timetable"

	"github.com/beevik/etree"
)

const fet_version = "7.5.5"

type fet_build struct {
	basedata *base.BaseData
	ttdata   *timetable.TtData

	Doc                *etree.Document
	WeightTable        []float64
	ConstraintElements []*etree.Element
	TimeConstraints    []int // indexes into `ConstraintElements`
	SpaceConstraints   []int // indexes into `ConstraintElements`
	Constraints        []constraint
	NConstraints       constraintIndex
	ConstraintTypes    []constraintType
	HardConstraintMap  map[constraintType][]constraintIndex
	SoftConstraintMap  map[constraintType][]constraintIndex

	fetroot                *etree.Element
	room_list              *etree.Element // needed for adding virtual rooms
	activity_tag_list      *etree.Element // in case these are needed
	time_constraints_list  *etree.Element
	space_constraints_list *etree.Element

	//TODO: What is this for?
	ActivityElementList []*etree.Element

	constraint_counter int // for tagging constraints

	DayList      []string
	HourList     []string
	TeacherList  []string
	SubjectList  []string
	RoomList     []string
	ActivityList []string

	// Cache for FET virtual rooms, "hash" -> FET-virtual-room tag
	fet_virtual_rooms  map[string]string
	fet_virtual_room_n map[string]int // FET-virtual-room tag -> number of room sets

	real_soft bool // if false, give soft constraints weight 100
}
