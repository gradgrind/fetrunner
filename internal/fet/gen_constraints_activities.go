package fet

import (
	"strconv"
)

type preferred_time struct {
	Preferred_Day  string
	Preferred_Hour string
}

func days_between(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintMinDaysBetweenActivities")
	c.CreateElement("Weight_Percentage").SetText(w1)
	cifsd := "false"
	if mapReadBool(constraint.Data, "ConsecutiveIfSameDay") {
		cifsd = "true"
	}
	alist := mapReadIndexList(constraint.Data, "Activities")
	n := mapReadInt(constraint.Data, "DaysBetween")
	c.CreateElement("Consecutive_If_Same_Day").SetText(cifsd)
	c.CreateElement("Number_of_Activities").SetText(strconv.Itoa(len(alist)))
	for _, ai := range alist {
		c.CreateElement("Activity_Id").SetText(fetbuild.ActivityList[ai])
	}
	c.CreateElement("MinDays").SetText(strconv.Itoa(n))
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func ends_day(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintActivityEndsStudentsDay")
	c.CreateElement("Weight_Percentage").SetText(w1)
	ai := anyInt(constraint.Data)
	c.CreateElement("Activity_Id").SetText(fetbuild.ActivityList[ai])
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func after_hour(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	h0 := mapReadInt(constraint.Data, "Hour") + 1
	ndays := len(fetbuild.DayList)
	nhours := len(fetbuild.HourList)
	timeslots := []preferred_time{}
	for d := range ndays {
		for h := h0; h < nhours; h++ {
			timeslots = append(timeslots, preferred_time{
				Preferred_Day:  fetbuild.DayList[d],
				Preferred_Hour: fetbuild.HourList[h],
			})
		}
	}
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintActivityPreferredTimeSlots")
	c.CreateElement("Weight_Percentage").SetText(w1)
	ai := mapReadInt(constraint.Data, "Activity")
	c.CreateElement("Activity_Id").SetText(fetbuild.ActivityList[ai])
	c.CreateElement("Number_of_Preferred_Time_Slots").
		SetText(strconv.Itoa(len(timeslots)))
	for _, t := range timeslots {
		pts := c.CreateElement("Preferred_Time_Slot")
		pts.CreateElement("Preferred_Day").SetText(t.Preferred_Day)
		pts.CreateElement("Preferred_Hour").SetText(t.Preferred_Hour)
	}
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func before_hour(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	h0 := mapReadInt(constraint.Data, "Hour") + 1
	ndays := len(fetbuild.DayList)
	timeslots := []preferred_time{}
	for d := range ndays {
		for h := 0; h < h0; h++ {
			timeslots = append(timeslots, preferred_time{
				Preferred_Day:  fetbuild.DayList[d],
				Preferred_Hour: fetbuild.HourList[h],
			})
		}
	}
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintActivityPreferredTimeSlots")
	c.CreateElement("Weight_Percentage").SetText(w1)
	ai := mapReadInt(constraint.Data, "Activity")
	c.CreateElement("Activity_Id").SetText(fetbuild.ActivityList[ai])
	c.CreateElement("Number_of_Preferred_Time_Slots").
		SetText(strconv.Itoa(len(timeslots)))
	for _, t := range timeslots {
		pts := c.CreateElement("Preferred_Time_Slot")
		pts.CreateElement("Preferred_Day").SetText(t.Preferred_Day)
		pts.CreateElement("Preferred_Hour").SetText(t.Preferred_Hour)
	}
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func parallel_activities(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintActivitiesSameStartingTime")
	c.CreateElement("Weight_Percentage").SetText(w1)
	alist := anyIntList(constraint.Data)
	c.CreateElement("Number_of_Activities").SetText(strconv.Itoa(len(alist)))
	for _, ai := range alist {
		c.CreateElement("Activity_Id").SetText(fetbuild.ActivityList[ai])
	}
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func double_no_break(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	nhours := len(fetbuild.HourList)
	ndays := len(fetbuild.DayList)
	timeslots := []preferred_time{}
	// Note that a double lesson can't start in the last slot of
	// the day.
	doubleBlocked := make([]bool, nhours-1)
	hlist := []int{}
	hours := mapReadIndexList(constraint.Data, "BreakHours")
	// Note that the breaks are immediately before the listed hours.
	for _, h := range hours {
		doubleBlocked[h-1] = true
		hlist = append(hlist, h)
	}
	for d := range ndays {
		for h, bl := range doubleBlocked {
			if !bl {
				timeslots = append(timeslots, preferred_time{
					Preferred_Day:  fetbuild.DayList[d],
					Preferred_Hour: fetbuild.HourList[h],
				})
			}
		}
	}
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintActivityPreferredStartingTimes")
	c.CreateElement("Weight_Percentage").SetText(w1)
	a := fetbuild.ActivityList[mapReadInt(constraint.Data, "Activity")]
	c.CreateElement("Activity_Id").SetText(a)
	c.CreateElement("Number_of_Preferred_Starting_Times").SetText(strconv.Itoa(len(timeslots)))
	for _, t := range timeslots {
		pts := c.CreateElement("Preferred_Starting_Time")
		pts.CreateElement("Day").SetText(t.Preferred_Day)
		pts.CreateElement("Hour").SetText(t.Preferred_Hour)
	}
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
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
