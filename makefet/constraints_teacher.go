package makefet

import (
	"fetrunner/db"
	"fetrunner/timetable"
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

func add_teacher_constraints(tt_data *timetable.TtData) {
	db0 := tt_data.Db
	ndays := tt_data.NDays
	nhours := tt_data.NHours
	tclist := tt_data.BackendData.(*FetData).time_constraints_list

	tnamap := map[db.NodeRef][]db.TimeSlot{} // needed for lunch-break constraints

	for _, c0 := range db0.Constraints[db.C_TeacherNotAvailable] {
		// The weight is assumed to be 100%.
		data := c0.Data.(db.ResourceNotAvailable)
		tref := data.Resource
		tnamap[tref] = data.NotAvailable
		// `NotAvailable` is an ordered list of time-slots in which the
		// teacher is to be regarded as not available for the timetable.
		if len(data.NotAvailable) != 0 {
			cna := tclist.CreateElement("ConstraintTeacherNotAvailableTimes")
			cna.CreateElement("Weight_Percentage").SetText("100")
			cna.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			cna.CreateElement("Number_of_Not_Available_Times").
				SetText(strconv.Itoa(len(data.NotAvailable)))
			for _, slot := range data.NotAvailable {
				nat := cna.CreateElement("Not_Available_Time")
				nat.CreateElement("Day").SetText(db0.Days[slot.Day].GetTag())
				nat.CreateElement("Hour").SetText(db0.Hours[slot.Hour].GetTag())
			}
			cna.CreateElement("Active").SetText("true")
			cna.CreateElement("Comments").SetText(resource_constraint(
				db.C_TeacherNotAvailable, c0.Id, tref))
		}
	}

	for _, c0 := range db0.Constraints[db.C_TeacherMaxDays] {
		data := c0.Data.(db.ResourceN)
		n := data.N
		if n >= 0 && n < ndays {
			tref := data.Resource
			cna := tclist.CreateElement("ConstraintTeacherMaxDaysPerWeek")
			cna.CreateElement("Weight_Percentage").SetText("100")
			cna.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			cna.CreateElement("Max_Days_Per_Week").SetText(strconv.Itoa(n))
			cna.CreateElement("Active").SetText("true")
			cna.CreateElement("Comments").SetText(resource_constraint(
				db.C_TeacherMaxDays, c0.Id, tref))
		}
	}

	for _, c0 := range db0.Constraints[db.C_TeacherMinActivitiesPerDay] {
		data := c0.Data.(db.ResourceN)
		n := data.N
		if n >= 2 && n <= nhours {
			tref := data.Resource
			c := tclist.CreateElement("ConstraintTeacherMinHoursDaily")
			c.CreateElement("Weight_Percentage").SetText("100")
			c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			c.CreateElement("Minimum_Hours_Daily").SetText(strconv.Itoa(n))
			c.CreateElement("Allow_Empty_Days").SetText("true")
			c.CreateElement("Active").SetText("true")
			c.CreateElement("Comments").SetText(resource_constraint(
				db.C_TeacherMinActivitiesPerDay, c0.Id, tref))
		}
	}

	for _, c0 := range db0.Constraints[db.C_TeacherMaxActivitiesPerDay] {
		data := c0.Data.(db.ResourceN)
		n := data.N
		if n >= 2 && n <= nhours {
			tref := data.Resource
			c := tclist.CreateElement("ConstraintTeacherMaxHoursDaily")
			c.CreateElement("Weight_Percentage").SetText("100")
			c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			c.CreateElement("Maximum_Hours_Daily").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")
			c.CreateElement("Comments").SetText(resource_constraint(
				db.C_TeacherMaxActivitiesPerDay, c0.Id, tref))
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
			n := data.N
			if n < ndays {
				tref := data.Resource
				c := tclist.CreateElement("ConstraintTeacherIntervalMaxDaysPerWeek")
				c.CreateElement("Weight_Percentage").SetText("100")
				c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
				c.CreateElement("Interval_Start_Hour").SetText(db0.Hours[h0].GetTag())
				c.CreateElement("Interval_End_Hour").SetText("")
				c.CreateElement("Max_Days_Per_Week").SetText(strconv.Itoa(n))
				c.CreateElement("Active").SetText("true")
				c.CreateElement("Comments").SetText(resource_constraint(
					db.C_TeacherMaxAfternoons, c0.Id, tref))
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
			tref := c0.Data.(db.NodeRef)
			// Generate the constraint unless all days have a blocked
			// lesson at lunchtime.
			lbdmap := make([]bool, ndays)
			for _, ts := range tnamap[tref] {
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
				c.CreateElement("Weight_Percentage").SetText("100")
				c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
				c.CreateElement("Interval_Start_Hour").
					SetText(db0.Hours[mbhours[0]].GetTag())
				c.CreateElement("Interval_End_Hour").
					SetText(db0.Hours[mbhours[0]+len(mbhours)].GetTag())
				c.CreateElement("Maximum_Hours_Daily").
					SetText(strconv.Itoa(len(mbhours) - 1))
				c.CreateElement("Active").SetText("true")
				c.CreateElement("Comments").SetText(resource_constraint(
					db.C_TeacherLunchBreak, c0.Id, tref))
				lbmap[tref] = lbdays
			}
		}
	}

	for _, c0 := range db0.Constraints[db.C_TeacherMaxGapsPerDay] {
		data := c0.Data.(db.ResourceN)
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
			c.CreateElement("Weight_Percentage").SetText("100")
			c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")
			c.CreateElement("Comments").SetText(resource_constraint(
				db.C_TeacherMaxGapsPerDay, c0.Id, tref))
		}
	}

	for _, c0 := range db0.Constraints[db.C_TeacherMaxGapsPerWeek] {
		data := c0.Data.(db.ResourceN)
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
			c.CreateElement("Weight_Percentage").SetText("100")
			c.CreateElement("Teacher").SetText(db0.Ref2Tag(tref))
			c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")
			c.CreateElement("Comments").SetText(resource_constraint(
				db.C_TeacherMaxGapsPerWeek, c0.Id, tref))
		}
	}
}
