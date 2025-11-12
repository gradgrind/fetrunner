package makefet

import (
	"fetrunner/timetable"
	"strconv"
)

func days_between(tt_data *timetable.TtData) {
	tclist := tt_data.BackendData.(*FetData).time_constraints_list

	for _, c0 := range tt_data.MinDaysBetweenActivities {
		w := weight2fet(c0.Weight)
		for _, alist := range c0.ActivityLists {
			aidlist := make([]int, len(alist))

			for i, ai := range alist {
				aidlist[i] = int(ai) + 1
			}
			cifsd := "false"
			if c0.ConsecutiveIfSameDay {
				cifsd = "true"
			}

			c := tclist.CreateElement("ConstraintMinDaysBetweenActivities")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Consecutive_If_Same_Day").SetText(cifsd)
			c.CreateElement("Number_of_Activities").SetText(strconv.Itoa(len(alist)))
			for _, ai := range alist {
				c.CreateElement("Activity_Id").SetText(strconv.Itoa(int(ai) + 1))
			}
			c.CreateElement("MinDays").SetText(strconv.Itoa(c0.DaysBetween))
			c.CreateElement("Active").SetText("true")
			c.CreateElement("Comments").SetText(activities_constraint(
				c0.Constraint, c0.Id, alist))

		}
	}
}
