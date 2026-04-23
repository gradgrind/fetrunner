package fetrunner

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"strconv"
	"strings"
)

func init() {
	OpHandlerMap["DAYS"] = get_days
	OpHandlerMap["HOURS"] = get_hours
	OpHandlerMap["CLASSES"] = get_classes
	OpHandlerMap["PLACEMENTS"] = get_placements
}

// The AutoTtData instance is available as `autotimetable.AutoTt`.

func get_days(op *DispatchOp) bool {
	if CheckArgs(op, 0) {
		for _, d := range autotimetable.AutoTt.Source.GetDays() {
			base.LogResult(op.Op, d.Tag)
		}
	}
	return true
}

func get_hours(op *DispatchOp) bool {
	if CheckArgs(op, 0) {
		for _, h := range autotimetable.AutoTt.Source.GetHours() {
			base.LogResult(op.Op, h.Tag)
		}
	}
	return true
}

func get_classes(op *DispatchOp) bool {
	if CheckArgs(op, 0) {
		for _, cls := range autotimetable.AutoTt.Source.GetClasses() {
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
			base.LogResult(op.Op, cls.Tag+":"+ais+":"+gs)
		}
	}
	return true
}

func get_placements(op *DispatchOp) bool {
	if CheckArgs(op, 0) {
		for _, p := range autotimetable.AutoTt.GetPlacements() {
			base.LogResult("PLACEMENT", p)
		}
	}
	return true
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
// class, so that a map can be built. In fact, because an activity can
// include more than one class, just the number of atomic groups is
// pretty useless!
