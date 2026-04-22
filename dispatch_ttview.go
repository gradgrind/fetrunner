package fetrunner

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
)

func init() {
	OpHandlerMap["DAYS"] = get_days
	OpHandlerMap["HOURS"] = get_hours
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

func get_placements(op *DispatchOp) bool {
	if CheckArgs(op, 0) {
		for _, p := range autotimetable.AutoTt.GetPlacements() {
			base.LogResult("PLACEMENT", p)
		}
	}
	return true
}
