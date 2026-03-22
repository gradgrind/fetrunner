package timetable

import (
	"fetrunner/internal/base"
)

func (tt_data *TtData) activity_constraints(constraint_map map[string][]*base.BaseConstraint) {
	tt_data.before_after(constraint_map)
	tt_data.double_unbroken(constraint_map)
	tt_data.end_of_day(constraint_map)
}

func (tt_data *TtData) before_after(constraint_map map[string][]*base.BaseConstraint) {
	for _, ctype := range []string{base.C_AfterHour, base.C_BeforeHour} {
		for _, c0 := range constraint_map[ctype] {
			data := c0.Data.(base.BeforeAfterHour)
			for _, c := range data.Courses {
				cinfo, ok := tt_data.Ref2CourseInfo[c]
				if !ok {
					panic("Invalid course Id in constraint " + ctype + ": " + string(c))
				}
				for _, ai := range cinfo.Activities {
					tt_data.constraints = append(tt_data.constraints, &constraint{
						Id:     string(c0.Id),
						CType:  ctype,
						Weight: c0.Weight,
						Data:   map[string]any{"Activity": ai, "Hour": data.Hour},
					})
				}
			}
			delete(constraint_map, ctype)
		}
	}
}

// TODO: Should the break times be "global"?
func (tt_data *TtData) double_unbroken(constraint_map map[string][]*base.BaseConstraint) {
	ctype := base.C_DoubleActivityNotOverBreaks
	dulist := constraint_map[ctype]
	if len(dulist) == 1 {
		activities := tt_data.db.Activities // for access to the durations
		du := dulist[0]
		id := du.Id
		w := du.Weight
		break_hours := du.Data.([]int)
		for ai, a := range activities {
			if a.Duration == 2 {
				tt_data.constraints = append(tt_data.constraints, &constraint{
					Id:     string(id),
					CType:  ctype,
					Weight: w,
					Data:   map[string]any{"Activity": ai, "BreakHours": break_hours},
					// Note that the breaks are immediately before the listed hours.
				})
			}
		}
	} else {
		panic("Constraint type must be used once only: " + base.C_DoubleActivityNotOverBreaks)
	}
	delete(constraint_map, ctype)
}

func (tt_data *TtData) end_of_day(constraint_map map[string][]*base.BaseConstraint) {
	ctype := base.C_ActivitiesEndDay
	for _, c0 := range constraint_map[ctype] {
		course := c0.Data.(nodeRef)
		cinfo := tt_data.Ref2CourseInfo[course]
		for _, ai := range cinfo.Activities {
			tt_data.constraints = append(tt_data.constraints, &constraint{
				Id:     string(c0.Id),
				CType:  ctype,
				Weight: c0.Weight,
				Data:   ai})
		}
	}
	delete(constraint_map, ctype)
}
