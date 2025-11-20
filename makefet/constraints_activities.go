package makefet

import (
	"fetrunner/db"
	"strconv"
)

type preferred_time struct {
	Preferred_Day  string
	Preferred_Hour string
}

func (fetbuild *FetBuild) add_activity_constraints() {
	fetbuild.before_after_hour()
	fetbuild.double_no_break()
	fetbuild.days_between()
	fetbuild.ends_day()
	fetbuild.parallel_activities()
}

func (fetbuild *FetBuild) days_between() {
	tt_data := fetbuild.ttdata
	rundata := fetbuild.rundata
	tclist := fetbuild.time_constraints_list
	for _, c0 := range tt_data.MinDaysBetweenActivities {
		w := rundata.FetWeight(c0.Weight)
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

func (fetbuild *FetBuild) ends_day() {
	tt_data := fetbuild.ttdata
	db0 := tt_data.Db
	rundata := fetbuild.rundata
	tclist := fetbuild.time_constraints_list
	for _, c0 := range db0.Constraints[db.C_ActivitiesEndDay] {
		w := rundata.FetWeight(c0.Weight)
		course := c0.Data.(db.NodeRef)
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

func (fetbuild *FetBuild) parallel_activities() {
	tt_data := fetbuild.ttdata
	rundata := fetbuild.rundata
	tclist := fetbuild.time_constraints_list
	for _, c0 := range tt_data.ParallelActivities {
		w := rundata.FetWeight(c0.Weight)
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

func (fetbuild *FetBuild) before_after_hour() {
	tt_data := fetbuild.ttdata
	db0 := tt_data.Db
	rundata := fetbuild.rundata

	for _, c0 := range db0.Constraints[db.C_AfterHour] {
		data := c0.Data.(*db.BeforeAfterHour)
		timeslots := []preferred_time{}
		for d := 0; d < tt_data.NDays; d++ {
			for h := data.Hour + 1; h < tt_data.NHours; h++ {
				timeslots = append(timeslots, preferred_time{
					Preferred_Day:  rundata.DayIds[d].Backend,
					Preferred_Hour: rundata.HourIds[h].Backend,
				})
			}
		}
		fetbuild.make_before_after_hour(c0, timeslots)
	}

	for _, c0 := range db0.Constraints[db.C_BeforeHour] {
		data := c0.Data.(*db.BeforeAfterHour)
		timeslots := []preferred_time{}
		for d := 0; d < tt_data.NDays; d++ {
			for h := 0; h < data.Hour; h++ {
				timeslots = append(timeslots, preferred_time{
					Preferred_Day:  rundata.DayIds[d].Backend,
					Preferred_Hour: rundata.HourIds[h].Backend,
				})
			}
		}
		fetbuild.make_before_after_hour(c0, timeslots)
	}
}

func (fetbuild *FetBuild) make_before_after_hour(
	c0 *db.Constraint, timeslots []preferred_time,
) {
	tt_data := fetbuild.ttdata
	rundata := fetbuild.rundata
	tclist := fetbuild.time_constraints_list
	data := c0.Data.(*db.BeforeAfterHour)
	w := rundata.FetWeight(c0.Weight)
	for _, course := range data.Courses {
		cinfo, ok := tt_data.Ref2CourseInfo[course]
		if !ok {
			tt_data.Db.Logger.Bug("Invalid course: %s", course)
			continue
		}
		for _, ai := range cinfo.Activities {
			c := tclist.CreateElement("ConstraintActivityPreferredTimeSlots")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Activity_Id").SetText(fet_activity_index(ai))
			c.CreateElement("Number_of_Preferred_Time_Slots").
				SetText(strconv.Itoa(len(timeslots)))
			for _, t := range timeslots {
				pts := c.CreateElement("Preferred_Time_Slot")
				pts.CreateElement("Preferred_Day").SetText(t.Preferred_Day)
				pts.CreateElement("Preferred_Hour").SetText(t.Preferred_Hour)
			}
			c.CreateElement("Active").SetText("true")

			fetbuild.add_time_constraint(c, params_constraint(
				c0.CType, c0.Id, []int{ai, data.Hour}, c0.Weight))
		}
	}
}

func (fetbuild *FetBuild) double_no_break() {
	tt_data := fetbuild.ttdata
	rundata := fetbuild.rundata
	tclist := fetbuild.time_constraints_list

	var doubleBlocked []bool
	for _, c0 := range tt_data.Db.Constraints[db.C_DoubleActivityNotOverBreaks] {
		if len(doubleBlocked) != 0 {
			tt_data.Db.Logger.Bug("Constraint DoubleActivityNotOverBreaks" +
				" specified more than once")
			continue
		}
		w := rundata.FetWeight(c0.Weight)
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
						Preferred_Day:  rundata.DayIds[d].Backend,
						Preferred_Hour: rundata.HourIds[h].Backend,
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
for _, c0 := range db0.Constraints[db.C_MinHoursFollowing] {
	...Error("!!! Constraint not implemented:\n%+v\n", c0)
	//w := weight2fet(c0.Weight)
	//data := c0.Data.(*db.MinHoursFollowing)

	//MinHoursFollowing{
	//	Course1: course1,
	//	Course2: course2,
	//	Hours:   hours,
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
