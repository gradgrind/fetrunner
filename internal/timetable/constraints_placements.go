package timetable

import (
	"fetrunner/internal/base"
)

func (tt_data *TtData) placement_constraints(constraint_map map[string][]*base.BaseConstraint) {
	for _, c0 := range constraint_map[base.C_ActivityStartTime] {
		data := c0.Data.(base.ActivityStartTime)
		ai := tt_data.Ref2ActivityIndex[data.Activity]
		tt_data.constraints = append(tt_data.constraints, &constraint{
			Id:     string(c0.Id),
			CType:  c0.CType,
			Weight: c0.Weight,
			Data: map[string]any{
				"Activity": ai, "Time": base.TimeSlot{Day: data.Day, Hour: data.Hour},
			},
		})

		// Collect the hard fixed activity placements as these are needed in
		// the generation of the days-between constraints.
		if c0.IsHard() {
			tt_data.TtActivities[ai].fixedStartTime = &base.TimeSlot{
				Day: data.Day, Hour: data.Hour}
		}
	}
	delete(constraint_map, base.C_ActivityStartTime)
}
