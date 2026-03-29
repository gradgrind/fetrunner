package fet

import (
	"strconv"
)

// Convert "base" constraints to "FET" constraints.

/* Lunch-breaks

Lunch-breaks can be done using max-hours-in-interval constraint, but that
makes specification of max-gaps more difficult (because the lunch breaks
count as gaps).

The alternative is to add dummy lessons, clamped to the midday-break hours,
on the days where none of the midday-break hours are blocked. However, this
can also cause problems with gaps – the dummy lesson can itself create gaps,
for example when a teacher's lessons are earlier in the day.

All in all, I think the max-hours-in-interval constraint is probably better
for the teachers, but it may be necessary to adjust the maximum number of
gaps permitted to accommodate any lunch breaks. FET doesn't offer any perfect
solution.

*/

// ------------------------------------------------------------------------

func teacher_max_days(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintActivityPreferredStartingTime")
	c.CreateElement("Weight_Percentage").SetText(w1)
	tix := mapReadInt(constraint.Data, "Teacher")
	c.CreateElement("Teacher").SetText(fetbuild.TeacherList[tix])
	n := mapReadInt(constraint.Data, "MaxDays")
	c.CreateElement("Max_Days_Per_Week").SetText(strconv.Itoa(n))
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func teacher_min_hours_per_day(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintTeacherMinHoursDaily")
	c.CreateElement("Weight_Percentage").SetText(w1)
	tix := mapReadInt(constraint.Data, "Teacher")
	c.CreateElement("Teacher").SetText(fetbuild.TeacherList[tix])
	n := mapReadInt(constraint.Data, "nHours")
	c.CreateElement("Minimum_Hours_Daily").SetText(strconv.Itoa(n))
	c.CreateElement("Allow_Empty_Days").SetText("true")
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func teacher_max_hours_per_day(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintTeacherMaxHoursDaily")
	c.CreateElement("Weight_Percentage").SetText(w1)
	tix := mapReadInt(constraint.Data, "Teacher")
	c.CreateElement("Teacher").SetText(fetbuild.TeacherList[tix])
	n := mapReadInt(constraint.Data, "nHours")
	c.CreateElement("Maximum_Hours_Daily").SetText(strconv.Itoa(n))
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func teacher_max_afternoons(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintTeacherIntervalMaxDaysPerWeek")
	c.CreateElement("Weight_Percentage").SetText(w1)
	tix := mapReadInt(constraint.Data, "Teacher")
	c.CreateElement("Teacher").SetText(fetbuild.TeacherList[tix])
	n := mapReadInt(constraint.Data, "MaxAfternoons")
	h0 := mapReadInt(constraint.Data, "AfternoonStart")
	c.CreateElement("Interval_Start_Hour").SetText(fetbuild.HourList[h0])
	c.CreateElement("Interval_End_Hour").SetText("")
	c.CreateElement("Max_Days_Per_Week").SetText(strconv.Itoa(n))
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func teacher_lunch_breaks(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	// Generate the constraint unless all days have a blocked
	// lesson at lunchtime.
	//TODO: I also need to count the lunch-break days.
	mb0 := mapReadInt(constraint.Data, "Hour0")
	mb1 := mapReadInt(constraint.Data, "Hour1")
	tix := mapReadInt(constraint.Data, "Teacher")

nextday:
	for _, blist := range fetbuild.teacher_hard_blocked[tix] {
		for h := mb0; h <= mb1; h++ {
			if blist[h] {
				// A slot is blocked.
				continue nextday
			}
		}
		// This day has no blocked slots, generate the constraint.
		w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
		c := fetbuild.time_constraints_list.CreateElement("ConstraintTeacherMaxHoursDailyInInterval")
		c.CreateElement("Weight_Percentage").SetText(w1)
		c.CreateElement("Teacher").SetText(fetbuild.TeacherList[tix])
		c.CreateElement("Interval_Start_Hour").SetText(fetbuild.HourList[mb0])
		c.CreateElement("Interval_End_Hour").SetText(fetbuild.HourList[mb1+1])
		c.CreateElement("Maximum_Hours_Daily").SetText(strconv.Itoa(mb1 - mb0))
		c.CreateElement("Active").SetText("true")
		c.CreateElement("Comments").SetText(comment)

		fetbuild.ConstraintElements[i] = append(
			fetbuild.ConstraintElements[i], c)
		break
	}
}

func teacher_max_gaps_per_week(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintTeacherMaxGapsPerWeek")
	c.CreateElement("Weight_Percentage").SetText(w1)
	tix := mapReadInt(constraint.Data, "Teacher")
	c.CreateElement("Teacher").SetText(fetbuild.TeacherList[tix])
	n := mapReadInt(constraint.Data, "nHours")
	c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func teacher_max_gaps_per_day(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintTeacherMaxGapsPerDay")
	c.CreateElement("Weight_Percentage").SetText(w1)
	tix := mapReadInt(constraint.Data, "Teacher")
	c.CreateElement("Teacher").SetText(fetbuild.TeacherList[tix])
	n := mapReadInt(constraint.Data, "nHours")
	c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}
