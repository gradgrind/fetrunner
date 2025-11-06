package makefet

import (
	"fetrunner/db"
	"slices"
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

func (fetinfo *fetInfo) handle_teacher_constraints() {
	tt_data := fetinfo.tt_data
	db0 := tt_data.Db
	ndays := tt_data.NDays
	nhours := tt_data.NHours

	natimes := []teacherNotAvailable{}
	tnamap := map[db.NodeRef][]db.TimeSlot{}
	for _, c := range db0.Constraints[db.C_TeacherNotAvailable] {
		// The weight is assumed to be 100%.
		data := c.Data.(db.ResourceNotAvailable)
		tref := data.Resource
		tnamap[tref] = data.NotAvailable
		// `NotAvailable` is an ordered list of time-slots in which the
		// teacher is to be regarded as not available for the timetable.
		nats := []notAvailableTime{}
		for _, slot := range data.NotAvailable {
			nats = append(nats,
				notAvailableTime{
					Day:  db0.Days[slot.Day].GetTag(),
					Hour: db0.Hours[slot.Hour].GetTag()})
		}
		if len(nats) > 0 {
			natimes = append(natimes,
				teacherNotAvailable{
					Weight_Percentage:             100,
					Teacher:                       db0.Ref2Tag(tref),
					Number_of_Not_Available_Times: len(nats),
					Not_Available_Time:            nats,
					Active:                        true,
					Comments:                      string(c.Id),
				})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherNotAvailableTimes = natimes

	tmaxdpw := []maxDaysT{}
	for _, c := range db0.Constraints[db.C_TeacherMaxDays] {
		data := c.Data.(db.ResourceN)
		n := data.N
		if n >= 0 && n < ndays {
			tmaxdpw = append(tmaxdpw, maxDaysT{
				Weight_Percentage: 100,
				Teacher:           db0.Ref2Tag(data.Resource),
				Max_Days_Per_Week: n,
				Active:            true,
				Comments:          string(c.Id),
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxDaysPerWeek = tmaxdpw

	tminlpd := []minLessonsPerDayT{}
	for _, c := range db0.Constraints[db.C_TeacherMinActivitiesPerDay] {
		data := c.Data.(db.ResourceN)
		n := data.N
		if n >= 2 && n <= nhours {
			tminlpd = append(tminlpd, minLessonsPerDayT{
				Weight_Percentage:   100,
				Teacher:             db0.Ref2Tag(data.Resource),
				Minimum_Hours_Daily: n,
				Allow_Empty_Days:    true,
				Active:              true,
				Comments:            string(c.Id),
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMinHoursDaily = tminlpd

	tmaxlpd := []maxLessonsPerDayT{}
	for _, c := range db0.Constraints[db.C_TeacherMaxActivitiesPerDay] {
		data := c.Data.(db.ResourceN)
		n := data.N
		if n >= 2 && n <= nhours {
			tmaxlpd = append(tmaxlpd, maxLessonsPerDayT{
				Weight_Percentage:   100,
				Teacher:             db0.Ref2Tag(data.Resource),
				Maximum_Hours_Daily: n,
				Active:              true,
				Comments:            string(c.Id),
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxHoursDaily = tmaxlpd

	tmaxaft := []maxDaysinIntervalPerWeekT{}
	// Gather the max afternoons constraints as they may influence the
	// max-gaps constraints.
	//    teacher ref -> max number of afternoons
	pmmap := map[db.NodeRef]int{}
	h0 := db0.Info.FirstAfternoonHour
	if h0 > 0 {
		for _, c := range db0.Constraints[db.C_TeacherMaxAfternoons] {
			data := c.Data.(db.ResourceN)
			n := data.N
			if n < ndays {
				tmaxaft = append(tmaxaft, maxDaysinIntervalPerWeekT{
					Weight_Percentage:   100,
					Teacher:             db0.Ref2Tag(data.Resource),
					Interval_Start_Hour: db0.Hours[h0].GetTag(),
					Interval_End_Hour:   "", // end of day
					Max_Days_Per_Week:   n,
					Active:              true,
					Comments:            string(c.Id),
				})
				pmmap[data.Resource] = n
			}
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherIntervalMaxDaysPerWeek = tmaxaft

	tlblist := []lunchBreakT{}
	// Gather the lunch-break constraints as they may influence the
	// max-gaps constraints.
	//    teacher ref -> number of days with lunch break
	lbmap := map[db.NodeRef]int{}
	if mbhours := db0.Info.MiddayBreak; len(mbhours) != 0 {
		for _, c := range db0.Constraints[db.C_TeacherLunchBreak] {
			tref := c.Data.(db.NodeRef)
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
				tlblist = append(tlblist, lunchBreakT{
					Weight_Percentage:   100,
					Teacher:             db0.Ref2Tag(tref),
					Interval_Start_Hour: db0.Hours[mbhours[0]].GetTag(),
					Interval_End_Hour:   db0.Hours[mbhours[0]+len(mbhours)].GetTag(),
					Maximum_Hours_Daily: len(mbhours) - 1,
					Active:              true,
					Comments:            string(c.Id),
				})
				lbmap[tref] = lbdays
			}

		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxHoursDailyInInterval = tlblist

	tmaxgpd := []maxGapsPerDayT{}
	for _, c := range db0.Constraints[db.C_TeacherMaxGapsPerDay] {
		data := c.Data.(db.ResourceN)
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
			tmaxgpd = append(tmaxgpd, maxGapsPerDayT{
				Weight_Percentage: 100,
				Teacher:           db0.Ref2Tag(tref),
				Max_Gaps:          n,
				Active:            true,
				Comments:          string(c.Id),
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxGapsPerDay = tmaxgpd

	tmaxgpw := []maxGapsPerWeekT{}
	for _, c := range db0.Constraints[db.C_TeacherMaxGapsPerWeek] {
		data := c.Data.(db.ResourceN)
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
			tmaxgpw = append(tmaxgpw, maxGapsPerWeekT{
				Weight_Percentage: 100,
				Teacher:           db0.Ref2Tag(tref),
				Max_Gaps:          n,
				Active:            true,
			})
		}
	}
	fetinfo.fetdata.Time_Constraints_List.
		ConstraintTeacherMaxGapsPerWeek = tmaxgpw
}
