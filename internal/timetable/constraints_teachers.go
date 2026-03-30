package timetable

import (
	"fetrunner/internal/base"
)

func (tt_data *TtData) teacher_constraints(constraint_map map[string][]*base.BaseConstraint) {
	tt_data.teacher_max_days(constraint_map) // before teacher_lunchbreak()
	tt_data.teacher_hours_per_day(constraint_map)
	tt_data.teacher_afternoons(constraint_map) // before teacher_lunchbreak()
	tt_data.teacher_lunchbreak(constraint_map) // before teacher_max_gaps()
	tt_data.teacher_max_gaps(constraint_map)
}

func (tt_data *TtData) teacher_max_days(constraint_map map[string][]*base.BaseConstraint) {
	ndays := tt_data.ndays
	ctype := base.C_TeacherMaxDays
	for _, c0 := range constraint_map[ctype] {
		data := c0.Data.(base.ResourceN)
		n := data.N
		if n >= 0 && n < ndays {
			tix := tt_data.teacher2Index[data.Resource]
			tt_data.constraints = append(tt_data.constraints, &ttConstraint{
				Id:     string(c0.Id),
				CType:  ctype,
				Weight: c0.Weight,
				Data:   map[string]any{"Teacher": tix, "MaxDays": n},
			})
		}
	}
	delete(constraint_map, ctype)
}

func (tt_data *TtData) teacher_hours_per_day(constraint_map map[string][]*base.BaseConstraint) {
	nhours := tt_data.nhours
	for _, ctype := range []string{base.C_TeacherMinHoursPerDay, base.C_TeacherMaxHoursPerDay} {
		for _, c0 := range constraint_map[ctype] {
			data := c0.Data.(base.ResourceN)
			n := data.N
			if n >= 2 && n < nhours {
				tix := tt_data.teacher2Index[data.Resource]
				tt_data.constraints = append(tt_data.constraints, &ttConstraint{
					Id:     string(c0.Id),
					CType:  ctype,
					Weight: c0.Weight,
					Data:   map[string]any{"Teacher": tix, "nHours": n},
				})
			}
		}
		delete(constraint_map, ctype)
	}
}

func (tt_data *TtData) teacher_afternoons(constraint_map map[string][]*base.BaseConstraint) {
	ndays := tt_data.ndays
	h0 := tt_data.db.FirstAfternoonHour
	ctype := base.C_TeacherMaxAfternoons
	if h0 > 0 {
		for _, c0 := range constraint_map[ctype] {
			data := c0.Data.(base.ResourceN)
			n := data.N
			if n < ndays {
				tix := tt_data.teacher2Index[data.Resource]
				tt_data.constraints = append(tt_data.constraints, &ttConstraint{
					Id:     string(c0.Id),
					CType:  ctype,
					Weight: c0.Weight,
					Data: map[string]any{
						"Teacher": tix, "MaxAfternoons": n, "AfternoonStart": h0},
				})
			}
		}
	}
	delete(constraint_map, ctype)
}

func (tt_data *TtData) teacher_lunchbreak(constraint_map map[string][]*base.BaseConstraint) {
	ctype := base.C_TeacherLunchBreak
	if mb0 := tt_data.db.MiddayBreak0; mb0 != 0 {
		mb1 := tt_data.db.MiddayBreak1
		for _, c0 := range constraint_map[ctype] {
			tref := c0.Data.(nodeRef)
			tix := tt_data.teacher2Index[tref]
			lbdays := []int{} // collect days needing lunch-break
			blocked := tt_data.teacher_hard_blocked[tix]
		nextday:
			for d := range tt_data.ndays {
				if len(blocked) != 0 {
					blist := blocked[d]
					for h := mb0; h <= mb1; h++ {
						if blist[h] {
							// A slot is blocked.
							continue nextday
						}
					}
				}
				lbdays = append(lbdays, d) // this day has no blocked lunch-break slots
			}
			if len(lbdays) != 0 {
				tt_data.constraints = append(tt_data.constraints, &ttConstraint{
					Id:     string(c0.Id),
					CType:  ctype,
					Weight: c0.Weight,
					Data: map[string]any{"Teacher": tix, "Hour0": mb0, "Hour1": mb1,
						"Days": lbdays},
				})
			}
		}
	}
	delete(constraint_map, ctype)
}

func (tt_data *TtData) teacher_max_gaps(constraint_map map[string][]*base.BaseConstraint) {
	for _, ctype := range []string{base.C_TeacherMaxGapsPerDay, base.C_TeacherMaxGapsPerWeek} {
		for _, c0 := range constraint_map[ctype] {
			data := c0.Data.(base.ResourceN)
			n := data.N
			if n >= 0 {
				tix := tt_data.teacher2Index[data.Resource]
				tt_data.constraints = append(tt_data.constraints, &ttConstraint{
					Id:     string(c0.Id),
					CType:  ctype,
					Weight: c0.Weight,
					Data:   map[string]any{"Teacher": tix, "nHours": n},
				})
			}
		}
		delete(constraint_map, ctype)
	}
}
