package fetrunner

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fmt"
	"strconv"
	"strings"
)

func init() {
	OpHandlerMap["TT_DAYS"] = get_days
	OpHandlerMap["TT_HOURS"] = get_hours
	OpHandlerMap["TT_CLASSES"] = get_classes
	OpHandlerMap["TT_TEACHERS"] = get_teachers
	OpHandlerMap["TT_ROOMS"] = get_rooms
	OpHandlerMap["TT_ACTIVITIES"] = get_activities
	OpHandlerMap["TT_CLASS_PLACEMENTS"] = get_class_placements
	OpHandlerMap["TT_TEACHER_PLACEMENTS"] = get_teacher_placements
	OpHandlerMap["TT_ROOM_PLACEMENTS"] = get_room_placements
}

// The AutoTtData instance is available as `autotimetable.AutoTt`.

func get_days(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	for _, d := range lres.Days {
		base.LogResult(op.Op, d.Tag+":")
	}
}

func get_hours(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	for _, h := range lres.Hours {
		base.LogResult(op.Op, h.Tag+":")
	}
}

func get_classes(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	for _, cls := range lres.Classes {
		//TODO? If I could ensure that the atomic groups of a class are consecutive,
		// an alternative would be to include just start and end index here.
		ailist := []string{}
		for _, ai := range cls.AtomicIndexes {
			ailist = append(ailist, strconv.Itoa(ai))
		}
		glist := []string{}
		for _, g := range cls.Groups {
			glist = append(glist, g.Tag)
		}
		ais := strings.Join(ailist, ",")
		gs := strings.Join(glist, ",")
		base.LogResult(op.Op, cls.Tag+"::"+ais+":"+gs)
	}
}

// TODO: (long) names
func get_teachers(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	for _, t := range lres.Teachers {
		base.LogResult(op.Op, t.Tag+":")
	}
}

// TODO: (long) names
func get_rooms(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	for _, r := range lres.Rooms {
		base.LogResult(op.Op, r.Tag+":")
	}
}

func get_activities(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	for _, a := range lres.Activities {
		tlist := []string{}
		for _, tix := range a.Teachers {
			tlist = append(tlist, strconv.Itoa(tix))
		}
		aglist := []string{}
		for _, agix := range a.AtomicGroupIndexes {
			aglist = append(aglist, strconv.Itoa(agix))
		}
		glist := []string{}
		for _, g := range a.Groups {
			glist = append(glist, g.Tag)
		}
		base.LogResult(op.Op, fmt.Sprintf("%d:%s:%s:%s:%s",
			a.Duration, a.Subject,
			strings.Join(tlist, ","),
			strings.Join(aglist, ","),
			strings.Join(glist, ",")))
	}
}

func get_class_placements(op *DispatchOp) {
	cix, err := strconv.Atoi(op.Arg)
	if err != nil {
		panic(err)
	}
	lres := autotimetable.AutoTt.GetLastResult()
	for _, p := range autotimetable.ClassPlacements(lres, cix) {
		base.LogResult("PLACEMENT", autotimetable.SerializePlacement(p))
	}
}

func get_teacher_placements(op *DispatchOp) {
	tix, err := strconv.Atoi(op.Arg)
	if err != nil {
		panic(err)
	}
	lres := autotimetable.AutoTt.GetLastResult()
	for _, p := range autotimetable.TeacherPlacements(lres, tix) {
		base.LogResult("PLACEMENT", autotimetable.SerializePlacement(p))
	}
}

func get_room_placements(op *DispatchOp) {
	rix, err := strconv.Atoi(op.Arg)
	if err != nil {
		panic(err)
	}
	lres := autotimetable.AutoTt.GetLastResult()
	for _, p := range autotimetable.RoomPlacements(lres, rix) {
		base.LogResult("PLACEMENT", autotimetable.SerializePlacement(p))
	}
}

//TODO: Consider extending BaseElement to include a "long name". Alternatively,
// these could be fetched by further "get_xxx()" calls.

//TODO: Somehow it will be necessary to recognise which classes are
// represented by the groups. If there is always a clear separator
// in the group name between class part and group part, it could be
// done as a string split with the existing data. Another possibility
// would be to provide a mapping, e.g. via a list of groups for each
// class. Another possibility would be to pass the full atomic group
// list for each activity and also (separately) the lists for each
// class, so that a map can be built.
