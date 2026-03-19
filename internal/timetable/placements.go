package timetable

import (
	"fetrunner/internal/base"
)

func (tt_data *TtData) placement_constraints(constraint_map map[string][]*base.Constraint) {
	for _, c0 := range constraint_map[base.C_ActivityStartTime] {
		data := c0.Data.(base.ActivityStartTime)
		i := len(tt_data.constraints)
		ai := tt_data.Ref2ActivityIndex[data.Activity]
		tt_data.constraints = append(tt_data.constraints, &constraint{
			Id:     string(c0.Id),
			CType:  c0.CType,
			Weight: c0.Weight,
			Data: map[string]any{
				"Activity": ai, "Time": base.TimeSlot{Day: data.Day, Hour: data.Hour},
			},
		})
	}
	delete(constraint_map, base.C_ActivityStartTime)
}
