package timetable

import (
	"fetrunner/internal/base"
	"strconv"
)

func (tt_data *TtData) activity_constraints(constraint_map map[string][]*base.BaseConstraint) {
	tt_data.before_after(constraint_map)
	tt_data.double_unbroken(constraint_map)
	tt_data.ndays_between(constraint_map)

	//

	/*TODO--
	for _, c0 := range constraint_map[base.C_ActivityStartTime] {
		data := c0.Data.(base.ActivityStartTime)
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
	*/
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
	} else if len(dulist) != 0 {
		// Multiple entries should not be possible.
		panic("Multiple constraint type " + base.C_DoubleActivityNotOverBreaks)
	}
}

func (tt_data *TtData) ndays_between(constraint_map map[string][]*base.BaseConstraint) {
	ctype := base.C_AfterHour
	for _, c0 := range tt_data.minDaysBetweenActivities {
		//	for _, c0 := range constraint_map[ctype] {
	}
}

//--

func (fetbuild *fet_build) activity_constraints() {
	//fetbuild.before_after_hour()
	//fetbuild.double_no_break()
	fetbuild.days_between()
	fetbuild.ends_day()
	fetbuild.parallel_activities()
}

func (fetbuild *fet_build) days_between() {
	tt_data := fetbuild.ttdata
	//rundata := fetbuild.rundata
	tclist := fetbuild.time_constraints_list
	for _, c0 := range tt_data.MinDaysBetweenActivities {
		w := fetbuild.DbWeight2Fet(c0.Weight)
		for _, alist := range c0.ActivityLists {
			cifsd := "false"
			if c0.ConsecutiveIfSameDay {
				cifsd = "true"
			}
			c := tclist.CreateElement("ConstraintMinDaysBetweenActivities")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Consecutive_If_Same_Day").SetText(cifsd)
			c.CreateElement("Number_of_Activities").SetText(strconv.Itoa(len(alist)))
			for _, ai := range alist {
				c.CreateElement("Activity_Id").SetText(fet_activity_index((ai)))
			}
			c.CreateElement("MinDays").SetText(strconv.Itoa(c0.DaysBetween))
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, params_constraint(
				c0.CType, c0.Id, alist, c0.Weight))
		}
	}
}

func (fetbuild *fet_build) ends_day() {
	tt_data := fetbuild.ttdata
	db := fetbuild.basedata.Db
	//rundata := fetbuild.rundata
	tclist := fetbuild.time_constraints_list
	for _, c0 := range db.Constraints[base.C_ActivitiesEndDay] {
		w := fetbuild.DbWeight2Fet(c0.Weight)
		course := c0.Data.(NodeRef)
		cinfo := tt_data.Ref2CourseInfo[course]
		for _, ai := range cinfo.Activities {
			c := tclist.CreateElement("ConstraintActivityEndsStudentsDay")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Activity_Id").SetText(fet_activity_index(ai))
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, param_constraint(
				c0.CType, c0.Id, ai, c0.Weight))
		}
	}
}

func (fetbuild *fet_build) parallel_activities() {
	tt_data := fetbuild.ttdata
	//rundata := fetbuild.rundata
	tclist := fetbuild.time_constraints_list
	for _, c0 := range tt_data.ParallelActivities {
		w := fetbuild.DbWeight2Fet(c0.Weight)
		for _, alist := range c0.ActivityLists {
			c := tclist.CreateElement("ConstraintActivitiesSameStartingTime")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Number_of_Activities").SetText(strconv.Itoa(len(alist)))
			for _, ai := range alist {
				c.CreateElement("Activity_Id").SetText(fet_activity_index((ai)))
			}
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, params_constraint(
				c0.CType, c0.Id, alist, c0.Weight))
		}
	}
}

func (fetbuild *fet_build) double_no_break() {
	tt_data := fetbuild.ttdata
	db := fetbuild.basedata.Db
	//rundata := fetbuild.rundata
	tclist := fetbuild.time_constraints_list

	var doubleBlocked []bool
	for _, c0 := range db.Constraints[base.C_DoubleActivityNotOverBreaks] {
		if len(doubleBlocked) != 0 {
			fetbuild.basedata.Logger.Bug(
				"Constraint DoubleActivityNotOverBreaks" +
					" specified more than once")
			continue
		}
		w := fetbuild.DbWeight2Fet(c0.Weight)
		timeslots := []preferred_time{}
		// Note that a double lesson can't start in the last slot of
		// the day.
		doubleBlocked = make([]bool, tt_data.NHours-1)
		hlist := []int{}
		for _, h := range c0.Data.([]int) {
			doubleBlocked[h-1] = true
			hlist = append(hlist, h)
		}
		for d := 0; d < tt_data.NDays; d++ {
			for h, bl := range doubleBlocked {
				if !bl {
					timeslots = append(timeslots, preferred_time{
						Preferred_Day:  fetbuild.DayList[d],
						Preferred_Hour: fetbuild.HourList[h],
					})
				}
			}
		}

		c := tclist.CreateElement("ConstraintActivitiesPreferredStartingTimes")
		c.CreateElement("Weight_Percentage").SetText(w)
		c.CreateElement("Duration").SetText("2")
		c.CreateElement("Number_of_Preferred_Starting_Times").
			SetText(strconv.Itoa(len(timeslots)))
		for _, t := range timeslots {
			pts := c.CreateElement("Preferred_Starting_Time")
			pts.CreateElement("Preferred_Starting_Day").SetText(t.Preferred_Day)
			pts.CreateElement("Preferred_Starting_Hour").SetText(t.Preferred_Hour)
		}
		c.CreateElement("Active").SetText("true")

		fetbuild.add_time_constraint(c, params_constraint(
			c0.CType, c0.Id, hlist, c0.Weight))
	}
}

/* TODO
for _, c0 := range db0.Constraints[base.C_MinHoursFollowing] {
    ...Error("!!! Constraint not implemented:\n%+v", c0)
    //w := weight2fet(c0.Weight)
    //data := c0.Data.(*base.MinHoursFollowing)

    //MinHoursFollowing{
    //  Course1: course1,
    //  Course2: course2,
    //  Hours:   hours,
    //}

    // It may be better to specify these constraints in a better way!

       <ConstraintStudentsSetMinGapsBetweenOrderedPairOfActivityTags>
         <Weight_Percentage>100</Weight_Percentage>
         <Students>12G</Students>
         <First_Activity_Tag>tag2</First_Activity_Tag>
         <Second_Activity_Tag>tag1</Second_Activity_Tag>
         <MinGaps>1</MinGaps>
         <Active>true</Active>
         <Comments></Comments>
       </ConstraintStudentsSetMinGapsBetweenOrderedPairOfActivityTags>
}
*/
