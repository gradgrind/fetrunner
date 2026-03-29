package fet

import (
    "fetrunner/internal/autotimetable"
    "fetrunner/internal/base"
    "math"
    "slices"
    "strconv"

    "github.com/beevik/etree"
)

// In FET the IDs and "tags" (short names) are generally the same, and only
// unique within the repective category (teacher, room, etc.).

type element = base.ElementBase

type ttConstraint = autotimetable.TtConstraint
type constraintIndex = autotimetable.ConstraintIndex
type autoTtData = autotimetable.AutoTtData
type constraintType = autotimetable.ConstraintType
type ttActivity = autotimetable.TtActivity

type softWeight struct {
    Index  constraintIndex
    Weight string
}

type TtSourceFet struct {
    doc                *etree.Document
    weightTable        []float64
    constraintElements []*etree.Element // ordered constraint elements
    constraints        []*ttConstraint  // ordered constraint info for "autotimetable"
    t_constraints      []int            // (source) indexes of active time constraints
    s_constraints      []int            // (source) indexes of active space constraints
    activities         []*ttActivity

    softWeights []softWeight // used for reconstructing original soft weights

    //activityElements []*etree.Element

    // FET has time and space constraints separate. It might be useful in
    // some way to have that information here.
    //timeConstraints  []int // indexes into `ConstraintElements`
    //spaceConstraints []int // indexes into `ConstraintElements`

    //nConstraints      constraintIndex
    constraintTypes   []constraintType
    hardConstraintMap map[constraintType][]constraintIndex
    softConstraintMap map[constraintType][]constraintIndex

    days     []element
    hours    []element
    subjects []element
    teachers []element
    classes  []*autotimetable.TtClass
    rooms    []element
}

func (sourcefet *TtSourceFet) GetDays() []element {
    return sourcefet.days
}

func (sourcefet *TtSourceFet) GetHours() []element {
    return sourcefet.hours
}

func (sourcefet *TtSourceFet) GetTeachers() []element {
    return sourcefet.teachers
}

func (sourcefet *TtSourceFet) GetSubjects() []element {
    return sourcefet.subjects
}

func (sourcefet *TtSourceFet) GetRooms() []element {
    return sourcefet.rooms
}

func (sourcefet *TtSourceFet) GetClasses() []*autotimetable.TtClass {
    return sourcefet.classes
}

// TODO?
// I guess it should be possible to implement this properly (didn't I do it
// somewhere already?), but it might not be necessary ...
func (sourcefet *TtSourceFet) GetAtomicGroups() []string {
    return nil
}

func (sourcefet *TtSourceFet) GetActivities() []*ttActivity {
    return sourcefet.activities
}

func (sourcefet *TtSourceFet) GetConstraints() []*ttConstraint { return sourcefet.constraints }

func (sourcefet *TtSourceFet) GetConstraintTypes() []constraintType {
    return sourcefet.constraintTypes
}

func (sourcefet *TtSourceFet) GetResourceUnavailableConstraintTypes() []constraintType {
    return []constraintType{
        "ConstraintStudentsSetNotAvailableTimes",
        "ConstraintTeacherNotAvailableTimes",
        "ConstraintRoomNotAvailableTimes",
    }
}

func (sourcefet *TtSourceFet) GetConstraintMaps() (
    map[constraintType][]constraintIndex,
    map[constraintType][]constraintIndex,
) {
    return sourcefet.hardConstraintMap, sourcefet.softConstraintMap
}

func MakeFetWeights() []float64 {
    wtable := make([]float64, 101)
    wtable[0] = 0.0
    wtable[100] = 100.0
    for w := 1; w < 100; w++ {
        wf := float64(w + 1)
        denom := wf + math.Pow(2, (wf-50.0)*0.2)
        wtable[w] = 100.0 - 100.0/denom
    }
    return wtable
}

func FetWeight2Db(w string, weightTable []float64) int {
    wf, err := strconv.ParseFloat(w, 64)
    if err != nil {
        panic(err)
    }
    wdb, _ := slices.BinarySearch(weightTable, wf)
    return wdb
}
