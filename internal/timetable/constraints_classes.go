package timetable

import (
	"fetrunner/internal/base"
)

func (tt_data *TtData) class_constraints(constraint_map map[string][]*base.BaseConstraint) {
	tt_data.class_hours_per_day(constraint_map)
	tt_data.class_afternoons(constraint_map)
	tt_data.class_lunchbreak(constraint_map)
	tt_data.class_force_hour0(constraint_map)
	tt_data.class_max_gaps(constraint_map)
}

func (tt_data *TtData) class_hours_per_day(constraint_map map[string][]*base.BaseConstraint) {
	nhours := tt_data.nhours
	for _, ctype := range []string{base.C_ClassMinHoursPerDay, base.C_ClassMaxHoursPerDay} {
		for _, c0 := range constraint_map[ctype] {
			data := c0.Data.(base.ResourceN)
			n := data.N
			if n >= 2 && n < nhours {
				cix := tt_data.class2Index[data.Resource]
				tt_data.constraints = append(tt_data.constraints, &ttConstraint{
					Id:     string(c0.Id),
					CType:  ctype,
					Weight: c0.Weight,
					Data:   map[string]any{"Class": cix, "nHours": n},
				})
			}
		}
		delete(constraint_map, ctype)
	}
}

func (tt_data *TtData) class_afternoons(constraint_map map[string][]*base.BaseConstraint) {
	ndays := tt_data.ndays
	h0 := tt_data.db.FirstAfternoonHour
	ctype := base.C_ClassMaxAfternoons
	if h0 > 0 {
		for _, c0 := range constraint_map[ctype] {
			data := c0.Data.(base.ResourceN)
			n := data.N
			if n < ndays {
				cix := tt_data.class2Index[data.Resource]
				tt_data.constraints = append(tt_data.constraints, &ttConstraint{
					Id:     string(c0.Id),
					CType:  ctype,
					Weight: c0.Weight,
					Data: map[string]any{
						"Class": cix, "MaxAfternoons": n, "AfternoonStart": h0},
				})
			}
		}
	}
	delete(constraint_map, ctype)
}

func (tt_data *TtData) class_lunchbreak(constraint_map map[string][]*base.BaseConstraint) {
	ctype := base.C_ClassLunchBreak
	if mb0 := tt_data.db.MiddayBreak0; mb0 != 0 {
		mb1 := tt_data.db.MiddayBreak1
		for _, c0 := range constraint_map[ctype] {
			cref := c0.Data.(nodeRef)
			cix := tt_data.class2Index[cref]
			tt_data.constraints = append(tt_data.constraints, &ttConstraint{
				Id:     string(c0.Id),
				CType:  ctype,
				Weight: c0.Weight,
				Data:   map[string]any{"Class": cix, "Hour0": mb0, "Hour1": mb1},
			})
		}
	}
	delete(constraint_map, ctype)
}

func (tt_data *TtData) class_force_hour0(constraint_map map[string][]*base.BaseConstraint) {
	ctype := base.C_ClassForceFirstHour
	for _, c0 := range constraint_map[ctype] {
		cref := c0.Data.(nodeRef)
		cix := tt_data.class2Index[cref]
		tt_data.constraints = append(tt_data.constraints, &ttConstraint{
			Id:     string(c0.Id),
			CType:  ctype,
			Weight: c0.Weight,
			Data:   cix,
		})
	}
	delete(constraint_map, ctype)
}

func (tt_data *TtData) class_max_gaps(constraint_map map[string][]*base.BaseConstraint) {
	for _, ctype := range []string{base.C_ClassMaxGapsPerDay, base.C_ClassMaxGapsPerWeek} {
		for _, c0 := range constraint_map[ctype] {
			data := c0.Data.(base.ResourceN)
			n := data.N
			if n >= 0 {
				cix := tt_data.class2Index[data.Resource]
				tt_data.constraints = append(tt_data.constraints, &ttConstraint{
					Id:     string(c0.Id),
					CType:  ctype,
					Weight: c0.Weight,
					Data:   map[string]any{"Class": cix, "nHours": n},
				})
			}
		}
		delete(constraint_map, ctype)
	}
}
