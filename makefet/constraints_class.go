package makefet

import (
	"fetrunner/db"
	"slices"
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

func (fetinfo *fetInfo) handle_class_constraints() {
	tt_data := fetinfo.tt_data
	db0 := tt_data.Db
	ndays := tt_data.NDays
	nhours := tt_data.NHours

	natimes := []studentsNotAvailable{}
	cnamap := map[db.NodeRef][]db.TimeSlot{}
	for _, c := range db0.Constraints[db.C_ClassNotAvailable] {
		// The weight is assumed to be 100%.
		data := c.Data.(db.ResourceNotAvailable)
		cref := data.Resource
		cnamap[cref] = data.NotAvailable
		// `NotAvailable` is an ordered list of time-slots in which the
		// class is to be regarded as not available for the timetable.
		nats := []notAvailableTime{}
		for _, slot := range data.NotAvailable {
			nats = append(nats,
				notAvailableTime{
					Day:  db0.Days[slot.Day].GetTag(),
					Hour: db0.Hours[slot.Hour].GetTag()})
		}
		if len(nats) > 0 {
			natimes = append(natimes,
				studentsNotAvailable{
					Weight_Percentage:             100,
					Students:                      db0.Ref2Tag(cref),
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
					Comments: resource_constraint(
						c.Id, cref, db.C_ClassNotAvailable),
				})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetNotAvailableTimes = natimes

	cminlpd := []minLessonsPerDay{}
	for _, c := range db0.Constraints[db.C_ClassMinActivitiesPerDay] {
		data := c.Data.(db.ResourceN)
		cref := data.Resource
		n := data.N
		if n >= 2 && n <= nhours {
			cminlpd = append(cminlpd, minLessonsPerDay{
				Weight_Percentage:   100,
				Students:            db0.Ref2Tag(cref),
				Minimum_Hours_Daily: n,
				Allow_Empty_Days:    true,
				Active:              true,
				Comments: resource_constraint(
					c.Id, cref, db.C_ClassMinActivitiesPerDay),
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMinHoursDaily = cminlpd

	cmaxlpd := []maxLessonsPerDay{}
	for _, c := range db0.Constraints[db.C_ClassMaxActivitiesPerDay] {
		data := c.Data.(db.ResourceN)
		cref := data.Resource
		n := data.N
		if n >= 2 && n <= nhours {
			cmaxlpd = append(cmaxlpd, maxLessonsPerDay{
				Weight_Percentage:   100,
				Students:            db0.Ref2Tag(cref),
				Maximum_Hours_Daily: n,
				Active:              true,
				Comments: resource_constraint(
					c.Id, cref, db.C_ClassMaxActivitiesPerDay),
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxHoursDaily = cmaxlpd

	cmaxaft := []maxDaysinIntervalPerWeek{}
	// Gather the max afternoons constraints as they may influence the
	// max-gaps constraints.
	//    class ref -> max number of afternoons
	pmmap := map[db.NodeRef]int{}
	h0 := db0.Info.FirstAfternoonHour
	if h0 > 0 {
		for _, c := range db0.Constraints[db.C_ClassMaxAfternoons] {
			data := c.Data.(db.ResourceN)
			cref := data.Resource
			n := data.N
			if n < ndays {
				cmaxaft = append(cmaxaft, maxDaysinIntervalPerWeek{
					Weight_Percentage:   100,
					Students:            db0.Ref2Tag(cref),
					Interval_Start_Hour: db0.Hours[h0].GetTag(),
					Interval_End_Hour:   "", // end of day
					Max_Days_Per_Week:   n,
					Active:              true,
					Comments: resource_constraint(
						c.Id, cref, db.C_ClassMaxAfternoons),
				})
				pmmap[data.Resource] = n
			}
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetIntervalMaxDaysPerWeek = cmaxaft

	clblist := []lunchBreak{}
	// Gather the lunch-break constraints as they may influence the
	// max-gaps constraints.
	//    class ref -> number of days with lunch break
	lbmap := map[db.NodeRef]int{}
	if mbhours := db0.Info.MiddayBreak; len(mbhours) != 0 {
		for _, c := range db0.Constraints[db.C_ClassLunchBreak] {
			cref := c.Data.(db.NodeRef)
			// Generate the constraint unless all days have a blocked
			// lesson at lunchtime.
			lbdmap := make([]bool, ndays)
			for _, ts := range cnamap[cref] {
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
				clblist = append(clblist, lunchBreak{
					Weight_Percentage:   100,
					Students:            db0.Ref2Tag(cref),
					Interval_Start_Hour: db0.Hours[mbhours[0]].GetTag(),
					Interval_End_Hour:   db0.Hours[mbhours[0]+len(mbhours)].GetTag(),
					Maximum_Hours_Daily: len(mbhours) - 1,
					Active:              true,
					Comments: resource_constraint(
						c.Id, cref, db.C_ClassLunchBreak),
				})
				lbmap[cref] = lbdays
			}

		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxHoursDailyInInterval = clblist

	cmaxgpd := []maxGapsPerDay{}
	for _, c := range db0.Constraints[db.C_ClassMaxGapsPerDay] {
		data := c.Data.(db.ResourceN)
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
			cmaxgpd = append(cmaxgpd, maxGapsPerDay{
				Weight_Percentage: 100,
				Students:          db0.Ref2Tag(cref),
				Max_Gaps:          n,
				Active:            true,
				Comments: resource_constraint(
					c.Id, cref, db.C_ClassMaxGapsPerDay),
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxGapsPerDay = cmaxgpd

	cmaxgpw := []maxGapsPerWeek{}
	for _, c := range db0.Constraints[db.C_ClassMaxGapsPerWeek] {
		data := c.Data.(db.ResourceN)
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
			cmaxgpw = append(cmaxgpw, maxGapsPerWeek{
				Weight_Percentage: 100,
				Students:          db0.Ref2Tag(cref),
				Max_Gaps:          n,
				Active:            true,
				Comments: resource_constraint(
					c.Id, cref, db.C_ClassMaxGapsPerWeek),
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetMaxGapsPerWeek = cmaxgpw

	cmaxls := []maxLateStarts{}
	for _, c := range db0.Constraints[db.C_ClassForceFirstHour] {
		cref := c.Data.(db.NodeRef)
		cmaxls = append(cmaxls, maxLateStarts{
			Weight_Percentage:             100,
			Max_Beginnings_At_Second_Hour: 0,
			Students:                      db0.Ref2Tag(cref),
			Active:                        true,
			Comments: resource_constraint(
				c.Id, cref, db.C_ClassForceFirstHour),
		})

	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintStudentsSetEarlyMaxBeginningsAtSecondHour = cmaxls

}
