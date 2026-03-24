package fet

import (
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

The code here uses max-hours-in-interval constraints. It may be necessary to
adjust the maximum number of gaps permitted to accommodate any lunch breaks.
FET doesn't offer any perfect solution.

*/

// ------------------------------------------------------------------------

func class_min_hours_per_day(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintStudentsSetMinHoursDaily")
	c.CreateElement("Weight_Percentage").SetText(w1)
	cix := mapReadInt(constraint.Data, "Class")
	c.CreateElement("Students").SetText(fetbuild.ClassList[cix])
	n := mapReadInt(constraint.Data, "nHours")
	c.CreateElement("Minimum_Hours_Daily").SetText(strconv.Itoa(n))
	c.CreateElement("Allow_Empty_Days").SetText("true")
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func class_max_hours_per_day(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintStudentsSetMaxHoursDaily")
	c.CreateElement("Weight_Percentage").SetText(w1)
	cix := mapReadInt(constraint.Data, "Class")
	c.CreateElement("Students").SetText(fetbuild.ClassList[cix])
	n := mapReadInt(constraint.Data, "nHours")
	c.CreateElement("Maximum_Hours_Daily").SetText(strconv.Itoa(n))
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func class_max_afternoons(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintStudentsSetIntervalMaxDaysPerWeek")
	c.CreateElement("Weight_Percentage").SetText(w1)
	cix := mapReadInt(constraint.Data, "Class")
	c.CreateElement("Students").SetText(fetbuild.ClassList[cix])
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

func class_lunch_breaks(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintStudentsSetMaxHoursDailyInInterval")
	c.CreateElement("Weight_Percentage").SetText(w1)
	cix := mapReadInt(constraint.Data, "Class")
	c.CreateElement("Students").SetText(fetbuild.ClassList[cix])
	mb0 := mapReadInt(constraint.Data, "Hour0")
	mb1 := mapReadInt(constraint.Data, "Hour1")
	c.CreateElement("Interval_Start_Hour").SetText(fetbuild.HourList[mb0])
	c.CreateElement("Interval_End_Hour").SetText(fetbuild.HourList[mb1+1])
	c.CreateElement("Maximum_Hours_Daily").SetText(strconv.Itoa(mb1 - mb0))
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func class_max_gaps_per_week(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintStudentsSetMaxGapsPerWeek")
	c.CreateElement("Weight_Percentage").SetText(w1)
	cix := mapReadInt(constraint.Data, "Class")
	c.CreateElement("Students").SetText(fetbuild.ClassList[cix])
	n := mapReadInt(constraint.Data, "nHours")
	c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func class_max_gaps_per_day(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintStudentsSetMaxGapsPerDay")
	c.CreateElement("Weight_Percentage").SetText(w1)
	cix := mapReadInt(constraint.Data, "Class")
	c.CreateElement("Students").SetText(fetbuild.ClassList[cix])
	n := mapReadInt(constraint.Data, "nHours")
	c.CreateElement("Max_Gaps").SetText(strconv.Itoa(n))
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func class_force_first_hour(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintStudentsSetMaxGapsPerDay")
	c.CreateElement("Weight_Percentage").SetText(w1)
	cix := constraint.Data.(int)
	c.CreateElement("Students").SetText(fetbuild.ClassList[cix])
	c.CreateElement("Max_Beginnings_At_Second_Hour").SetText("0")
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}
