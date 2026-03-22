package timetable

import (
	"fetrunner/internal/base"
	"slices"
	"strconv"
)

func (tt_data *TtData) teacher_constraints(constraint_map map[string][]*base.BaseConstraint) {
	tt_data.teacher_max_days(constraint_map)
	tt_data.teacher_min_activities_per_day(constraint_map)
	tt_data.teacher_max_activities_per_day(constraint_map)

	//TODO ...

}

func (tt_data *TtData) teacher_max_days(constraint_map map[string][]*base.BaseConstraint) {
	ndays := tt_data.ndays
	ctype := base.C_TeacherMaxDays
	for _, c0 := range constraint_map[ctype] {
		data := c0.Data.(base.ResourceN)
		n := data.N
		if n >= 0 && n < ndays {
			tix := tt_data.Teacher2Index[data.Resource]
			tt_data.constraints = append(tt_data.constraints, &constraint{
				Id:     string(c0.Id),
				CType:  ctype,
				Weight: c0.Weight,
				Data:   map[string]any{"Teacher": tix, "MaxDays": n},
			})
		}
	}
	delete(constraint_map, ctype)
}

func (tt_data *TtData) teacher_min_activities_per_day(constraint_map map[string][]*base.BaseConstraint) {
	ndays := tt_data.ndays
	ctype := base.C_TeacherMinActivitiesPerDay
	for _, c0 := range constraint_map[ctype] {
		data := c0.Data.(base.ResourceN)
		n := data.N
		if n >= 0 && n < ndays {
			tix := tt_data.Teacher2Index[data.Resource]
			tt_data.constraints = append(tt_data.constraints, &constraint{
				Id:     string(c0.Id),
				CType:  ctype,
				Weight: c0.Weight,
				Data:   map[string]any{"Teacher": tix, "MinActivitiesPerDay": n},
			})
		}
	}
	delete(constraint_map, ctype)
}

func (tt_data *TtData) teacher_max_activities_per_day(constraint_map map[string][]*base.BaseConstraint) {
	ndays := tt_data.ndays
	ctype := base.C_TeacherMaxActivitiesPerDay
	for _, c0 := range constraint_map[ctype] {
		data := c0.Data.(base.ResourceN)
		n := data.N
		if n >= 0 && n < ndays {
			tix := tt_data.Teacher2Index[data.Resource]
			tt_data.constraints = append(tt_data.constraints, &constraint{
				Id:     string(c0.Id),
				CType:  ctype,
				Weight: c0.Weight,
				Data:   map[string]any{"Teacher": tix, "MaxActivitiesPerDay": n},
			})
		}
	}
	delete(constraint_map, ctype)
}

func (tt_data *TtData) teacher_gaps_afternoons(constraint_map map[string][]*base.BaseConstraint) {
	ndays := tt_data.ndays

	// Gather the max afternoons constraints as they may influence the
	// max-gaps constraints.
	pmmap := map[teacherIndex]int{} // teacher index -> max number of afternoons
	h0 := tt_data.db.FirstAfternoonHour
	ctype := base.C_TeacherMaxAfternoons
	if h0 > 0 {
		for _, c0 := range constraint_map[ctype] {
			data := c0.Data.(base.ResourceN)
			n := data.N
			if n < ndays {
				tix := tt_data.Teacher2Index[data.Resource]
				pmmap[tix] = n
				tt_data.constraints = append(tt_data.constraints, &constraint{
					Id:     string(c0.Id),
					CType:  ctype,
					Weight: c0.Weight,
					Data:   map[string]any{"Teacher": tix, "MaxAfternoons": n},
				})
			}
		}
	}

	// Gather the lunch-break constraints as they may influence the
	// max-gaps constraints.
	lbmap := map[teacherIndex]int{} // teacher ref -> number of days with lunch break
	ctype = base.C_TeacherLunchBreak
	if mb0 := tt_data.db.MiddayBreak0; mb0 != 0 {
		mb1 := tt_data.db.MiddayBreak1
		for _, c0 := range constraint_map[ctype] {
			tref := c0.Data.(nodeRef)
			// Generate the constraint unless all days have a blocked
			// lesson at lunchtime.
			lbdmap := make([]bool, ndays)
			for _, ts := range namap[tref] {
				if slices.Contains(mbhours, ts.Hour) {
					lbdmap[ts.Day] = true
				}
			}
			lbdays := ndays
			for _, b := range lbdmap {
				if b {
					lbdays--
				}
			}
			if lbdays != 0 {
				// Add a lunch-break constraint.
			}
		}
	}

}

// TODO--
func (fetbuild *fet_build) add_teacher_constraints(
	namap map[NodeRef][]base.TimeSlot,
) {
	tt_data := fetbuild.ttdata
	db := fetbuild.basedata.Db
	//rundata := fetbuild.rundata
	ndays := tt_data.NDays
	nhours := tt_data.NHours
	tclist := fetbuild.time_constraints_list

	// Gather the max afternoons constraints as they may influence the
	// max-gaps constraints.
	//    teacher ref -> max number of afternoons
	pmmap := map[NodeRef]int{}
	h0 := db.Info.FirstAfternoonHour
	if h0 > 0 {
		for _, c0 := range db.Constraints[base.C_TeacherMaxAfternoons] {
			data := c0.Data.(base.ResourceN)
			w := fetbuild.DbWeight2Fet(c0.Weight)
			n := data.N
			if n < ndays {
				tref := data.Resource
				c := tclist.CreateElement("ConstraintTeacherIntervalMaxDaysPerWeek")
				c.CreateElement("Weight_Percentage").SetText(w)
				c.CreateElement("Teacher").SetText(db.Ref2Tag(tref))
				c.CreateElement("Interval_Start_Hour").SetText(fetbuild.HourList[h0])
				c.CreateElement("Interval_End_Hour").SetText("")
				c.CreateElement("Max_Days_Per_Week").SetText(strconv.Itoa(n))
				c.CreateElement("Active").SetText("true")

				fetbuild.add_time_constraint(c, param_constraint(
					c0.CType, c0.Id, tt_data.TeacherIndex[tref], c0.Weight))
				pmmap[data.Resource] = n
			}
		}
	}

	// Gather the lunch-break constraints as they may influence the
	// max-gaps constraints.
	//    teacher ref -> number of days with lunch break
	lbmap := map[NodeRef]int{}
	if mbhours := db.Info.MiddayBreak; len(mbhours) != 0 {
		for _, c0 := range db.Constraints[base.C_TeacherLunchBreak] {
			w := fetbuild.DbWeight2Fet(c0.Weight)
			tref := c0.Data.(NodeRef)
			// Generate the constraint unless all days have a blocked
			// lesson at lunchtime.
			lbdmap := make([]bool, ndays)
			for _, ts := range namap[tref] {
				if slices.Contains(mbhours, ts.Hour) {
					lbdmap[ts.Day] = true
				}
			}
			lbdays := ndays
			for _, b := range lbdmap {
				if b {
					lbdays--
				}
			}
			if lbdays != 0 {
				// Add a lunch-break constraint.
				c := tclist.CreateElement("ConstraintTeacherMaxHoursDailyInInterval")
				c.CreateElement("Weight_Percentage").SetText(w)
				c.CreateElement("Teacher").SetText(db.Ref2Tag(tref))
				c.CreateElement("Interval_Start_Hour").
					SetText(fetbuild.HourList[mbhours[0]])
				c.CreateElement("Interval_End_Hour").
					SetText(fetbuild.HourList[mbhours[0]+len(mbhours)])
				c.CreateElement("Maximum_Hours_Daily").
					SetText(strconv.Itoa(len(mbhours) - 1))
				c.CreateElement("Active").SetText("true")

				fetbuild.add_time_constraint(c, param_constraint(
					c0.CType, c0.Id, tt_data.TeacherIndex[tref], c0.Weight))
				lbmap[tref] = lbdays
			}
		}
	}

	for _, c0 := range db.Constraints[base.C_TeacherMaxGapsPerDay] {
		data := c0.Data.(base.ResourceN)
		w := fetbuild.DbWeight2Fet(c0.Weight)
		n := data.N
		tref := data.Resource
		// Ensure that a gap is allowed if there are lunch breaks.
		if n == 0 {
			_, ok := lbmap[tref]
			if ok {
				// lbdays > 0
				maxpm, ok := pmmap[tref]
				if !ok || maxpm != 0 {
					n = 1
				}
			}
		}
		if n >= 0 {
			c := tclist.CreateElement("ConstraintTeacherMaxGapsPerDay")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Teacher").SetText(db.Ref2Tag(tref))
			c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, tt_data.TeacherIndex[tref], c0.Weight))
		}
	}

	for _, c0 := range db.Constraints[base.C_TeacherMaxGapsPerWeek] {
		data := c0.Data.(base.ResourceN)
		w := fetbuild.DbWeight2Fet(c0.Weight)
		n := data.N
		tref := data.Resource
		if n >= 0 {
			// Adjust to accommodate lunch breaks
			lbdays, ok := lbmap[tref]
			if ok {
				// lbdays > 0
				maxpm, ok := pmmap[tref]
				if ok && maxpm < lbdays {
					lbdays = maxpm
				}
				n += lbdays
			}
			c := tclist.CreateElement("ConstraintTeacherMaxGapsPerWeek")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Teacher").SetText(db.Ref2Tag(tref))
			c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, tt_data.TeacherIndex[tref], c0.Weight))
		}
	}
}
