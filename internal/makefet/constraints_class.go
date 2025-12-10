package makefet

import (
	"fetrunner/internal/base"
	"slices"
	"strconv"
)

/* Lunch-breaks

Lunch-breaks can be done using max-hours-in-interval constraint, but that
makes specification of max-gaps more difficult (because the lunch breaks
count as gaps).

The alternative is to add dummy lessons, clamped to the midday-break hours,
on the days where none of the midday-break hours are blocked. However, this
can also cause problems with gaps – the dummy lesson can itself create gaps,
for example when a class only has lessons earlier in the day.

Tests with the dummy lessons approach suggest that it is difficult to get the
number of these lessons and their placement on the correct days right.

The code here uses max-hours-in-interval constraints and tries to compensate
for the gaps that are thus created by adjusting the max-gaps constraints.

*/

// ------------------------------------------------------------------------

func (fetbuild *FetBuild) add_class_constraints(
	namap map[NodeRef][]base.TimeSlot,
) {
	tt_data := fetbuild.ttdata
	db := fetbuild.basedata.Db
	rundata := fetbuild.rundata
	ndays := tt_data.NDays
	nhours := tt_data.NHours
	tclist := fetbuild.time_constraints_list

	for _, c0 := range db.Constraints[base.C_ClassMinActivitiesPerDay] {
		data := c0.Data.(base.ResourceN)
		cref := data.Resource
		w := rundata.FetWeight(c0.Weight)
		n := data.N
		if n >= 2 && n <= nhours {
			c := tclist.CreateElement("ConstraintStudentsSetMinHoursDaily")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Students").SetText(db.Ref2Tag(cref))
			c.CreateElement("Minimum_Hours_Daily").SetText(strconv.Itoa(n))
			c.CreateElement("Allow_Empty_Days").SetText("true")
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, tt_data.ClassIndex[cref], c0.Weight))
		}
	}

	for _, c0 := range db.Constraints[base.C_ClassMaxActivitiesPerDay] {
		data := c0.Data.(base.ResourceN)
		cref := data.Resource
		w := rundata.FetWeight(c0.Weight)
		n := data.N
		if n >= 2 && n <= nhours {
			c := tclist.CreateElement("ConstraintStudentsSetMaxHoursDaily")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Students").SetText(db.Ref2Tag(cref))
			c.CreateElement("Maximum_Hours_Daily").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, tt_data.ClassIndex[cref], c0.Weight))
		}
	}

	// Gather the max afternoons constraints as they may influence the
	// max-gaps constraints.
	//    class ref -> max number of afternoons
	pmmap := map[NodeRef]int{}
	h0 := db.Info.FirstAfternoonHour
	if h0 > 0 {
		for _, c0 := range db.Constraints[base.C_ClassMaxAfternoons] {
			data := c0.Data.(base.ResourceN)
			cref := data.Resource
			w := rundata.FetWeight(c0.Weight)
			n := data.N
			if n < ndays {
				c := tclist.CreateElement("ConstraintStudentsSetIntervalMaxDaysPerWeek")
				c.CreateElement("Weight_Percentage").SetText(w)
				c.CreateElement("Students").SetText(db.Ref2Tag(cref))
				c.CreateElement("Interval_Start_Hour").SetText(rundata.HourIds[h0].Backend)
				c.CreateElement("Interval_End_Hour").SetText("")
				c.CreateElement("Max_Days_Per_Week").SetText(strconv.Itoa(n))
				c.CreateElement("Active").SetText("true")

				fetbuild.add_time_constraint(c, param_constraint(
					c0.CType, c0.Id, tt_data.ClassIndex[cref], c0.Weight))
				pmmap[data.Resource] = n
			}
		}
	}

	for _, c0 := range db.Constraints[base.C_ClassForceFirstHour] {
		cref := c0.Data.(NodeRef)
		w := rundata.FetWeight(c0.Weight)
		c := tclist.CreateElement("ConstraintStudentsSetEarlyMaxBeginningsAtSecondHour")
		c.CreateElement("Weight_Percentage").SetText(w)
		c.CreateElement("Students").SetText(db.Ref2Tag(cref))
		c.CreateElement("Max_Beginnings_At_Second_Hour").SetText("0")
		c.CreateElement("Active").SetText("true")

		fetbuild.add_time_constraint(c, param_constraint(
			c0.CType, c0.Id, tt_data.ClassIndex[cref], c0.Weight))
	}

	// Gather the lunch-break constraints as they may influence the
	// max-gaps constraints.
	//    class ref -> number of days with lunch break
	lbmap := map[NodeRef]int{}
	if mbhours := db.Info.MiddayBreak; len(mbhours) != 0 {
		for _, c0 := range db.Constraints[base.C_ClassLunchBreak] {
			cref := c0.Data.(NodeRef)
			w := rundata.FetWeight(c0.Weight)
			// Generate the constraint unless all days have a blocked
			// lesson at lunchtime.
			lbdmap := make([]bool, ndays)
			for _, ts := range namap[cref] {
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
				c := tclist.CreateElement("ConstraintStudentsSetMaxHoursDailyInInterval")
				c.CreateElement("Weight_Percentage").SetText(w)
				c.CreateElement("Students").SetText(db.Ref2Tag(cref))
				c.CreateElement("Interval_Start_Hour").
					SetText(rundata.HourIds[mbhours[0]].Backend)
				c.CreateElement("Interval_End_Hour").
					SetText(rundata.HourIds[mbhours[0]+len(mbhours)].Backend)
				c.CreateElement("Maximum_Hours_Daily").
					SetText(strconv.Itoa(len(mbhours) - 1))
				c.CreateElement("Active").SetText("true")

				fetbuild.add_time_constraint(c, param_constraint(
					c0.CType, c0.Id, tt_data.ClassIndex[cref], c0.Weight))
				lbmap[cref] = lbdays
			}
		}
	}

	for _, c0 := range db.Constraints[base.C_ClassMaxGapsPerDay] {
		data := c0.Data.(base.ResourceN)
		w := rundata.FetWeight(c0.Weight)
		n := data.N
		cref := data.Resource
		// Ensure that a gap is allowed if there are lunch breaks.
		if n == 0 {
			_, ok := lbmap[cref]
			if ok {
				// lbdays > 0
				maxpm, ok := pmmap[cref]
				if !ok || maxpm != 0 {
					n = 1
				}
			}
		}
		if n >= 0 {
			c := tclist.CreateElement("ConstraintStudentsSetMaxGapsPerDay")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Students").SetText(db.Ref2Tag(cref))
			c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, tt_data.ClassIndex[cref], c0.Weight))
		}
	}

	for _, c0 := range db.Constraints[base.C_ClassMaxGapsPerWeek] {
		data := c0.Data.(base.ResourceN)
		w := rundata.FetWeight(c0.Weight)
		n := data.N
		cref := data.Resource
		if n >= 0 {
			// Adjust to accommodate lunch breaks
			lbdays, ok := lbmap[cref]
			if ok {
				// lbdays > 0
				maxpm, ok := pmmap[cref]
				if ok && maxpm < lbdays {
					lbdays = maxpm
				}
				n += lbdays
			}
			c := tclist.CreateElement("ConstraintStudentsSetMaxGapsPerWeek")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Students").SetText(db.Ref2Tag(cref))
			c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, tt_data.ClassIndex[cref], c0.Weight))
		}
	}
}
