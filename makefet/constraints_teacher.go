package makefet

import (
	"fetrunner/db"
	"slices"
	"strconv"
)

/* Lunch-breaks

Lunch-breaks can be done using max-hours-in-interval constraint, but that
makes specification of max-gaps more difficult (becuase the lunch breaks
count as gaps).

The alternative is to add dummy lessons, clamped to the midday-break hours,
on the days where none of the midday-break hours are blocked. However, this
can also cause problems with gaps – the dummy lesson can itself create gaps,
for example when a teacher's lessons are earlier in the day.

All in all, I think the max-hours-in-interval constraint is probably better
for the teachers. The code here tries to compensate for the gaps that are
thus created by adjusting the max-gaps constraints.

*/

// ------------------------------------------------------------------------

func (fetbuild *FetBuild) add_teacher_constraints(
	namap map[db.NodeRef][]db.TimeSlot,
) {
	tt_data := fetbuild.ttdata
	db0 := tt_data.Db
	rundata := fetbuild.rundata
	ndays := tt_data.NDays
	nhours := tt_data.NHours
	tclist := fetbuild.time_constraints_list

	for _, c0 := range db0.Constraints[db.C_TeacherMaxDays] {
		data := c0.Data.(db.ResourceN)
		w := rundata.FetWeight(c0.Weight)
		n := data.N
		if n >= 0 && n < ndays {
			tref := data.Resource
			c := tclist.CreateElement("ConstraintTeacherMaxDaysPerWeek")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			c.CreateElement("Max_Days_Per_Week").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, tt_data.TeacherIndex[tref], c0.Weight))
		}
	}

	for _, c0 := range db0.Constraints[db.C_TeacherMinActivitiesPerDay] {
		data := c0.Data.(db.ResourceN)
		w := rundata.FetWeight(c0.Weight)
		n := data.N
		if n >= 2 && n <= nhours {
			tref := data.Resource
			c := tclist.CreateElement("ConstraintTeacherMinHoursDaily")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			c.CreateElement("Minimum_Hours_Daily").SetText(strconv.Itoa(n))
			c.CreateElement("Allow_Empty_Days").SetText("true")
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, tt_data.TeacherIndex[tref], c0.Weight))
		}
	}

	for _, c0 := range db0.Constraints[db.C_TeacherMaxActivitiesPerDay] {
		data := c0.Data.(db.ResourceN)
		w := rundata.FetWeight(c0.Weight)
		n := data.N
		if n >= 2 && n <= nhours {
			tref := data.Resource
			c := tclist.CreateElement("ConstraintTeacherMaxHoursDaily")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			c.CreateElement("Maximum_Hours_Daily").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, tt_data.TeacherIndex[tref], c0.Weight))
		}
	}

	// Gather the max afternoons constraints as they may influence the
	// max-gaps constraints.
	//    teacher ref -> max number of afternoons
	pmmap := map[db.NodeRef]int{}
	h0 := db0.Info.FirstAfternoonHour
	if h0 > 0 {
		for _, c0 := range db0.Constraints[db.C_TeacherMaxAfternoons] {
			data := c0.Data.(db.ResourceN)
			w := rundata.FetWeight(c0.Weight)
			n := data.N
			if n < ndays {
				tref := data.Resource
				c := tclist.CreateElement("ConstraintTeacherIntervalMaxDaysPerWeek")
				c.CreateElement("Weight_Percentage").SetText(w)
				c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
				c.CreateElement("Interval_Start_Hour").SetText(rundata.HourIds[h0].Backend)
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
	lbmap := map[db.NodeRef]int{}
	if mbhours := db0.Info.MiddayBreak; len(mbhours) != 0 {
		for _, c0 := range db0.Constraints[db.C_TeacherLunchBreak] {
			w := rundata.FetWeight(c0.Weight)
			tref := c0.Data.(db.NodeRef)
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
				c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
				c.CreateElement("Interval_Start_Hour").
					SetText(rundata.HourIds[mbhours[0]].Backend)
				c.CreateElement("Interval_End_Hour").
					SetText(rundata.HourIds[mbhours[0]+len(mbhours)].Backend)
				c.CreateElement("Maximum_Hours_Daily").
					SetText(strconv.Itoa(len(mbhours) - 1))
				c.CreateElement("Active").SetText("true")

				fetbuild.add_time_constraint(c, param_constraint(
					c0.CType, c0.Id, tt_data.TeacherIndex[tref], c0.Weight))
				lbmap[tref] = lbdays
			}
		}
	}

	for _, c0 := range db0.Constraints[db.C_TeacherMaxGapsPerDay] {
		data := c0.Data.(db.ResourceN)
		w := rundata.FetWeight(c0.Weight)
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
			c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, tt_data.TeacherIndex[tref], c0.Weight))
		}
	}

	for _, c0 := range db0.Constraints[db.C_TeacherMaxGapsPerWeek] {
		data := c0.Data.(db.ResourceN)
		w := rundata.FetWeight(c0.Weight)
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
			c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, tt_data.TeacherIndex[tref], c0.Weight))
		}
	}
}
