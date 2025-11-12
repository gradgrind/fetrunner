package makefet

import (
	"fetrunner/base"
	"fetrunner/db"
	"fetrunner/timetable"
	"fmt"
	"strconv"
)

type preferred_time struct {
	Preferred_Day  string
	Preferred_Hour string
}

type preferred_start struct {
	Preferred_Starting_Day  string
	Preferred_Starting_Hour string
}

func add_activity_constraints(tt_data *timetable.TtData) {
	before_after_hour(tt_data)

	/*
		db0 := tt_data.Db
		tclist := tt_data.BackendData.(*FetData).time_constraints_list

		//ConstraintActivityPreferredStartingTime    []startingTime
			ConstraintActivityPreferredTimeSlots       []activityPreferredTimes
			ConstraintActivitiesPreferredTimeSlots     []preferredSlots
			ConstraintActivitiesPreferredStartingTimes []preferredStarts
			ConstraintMinDaysBetweenActivities         []minDaysBetweenActivities
			ConstraintActivityEndsStudentsDay          []lessonEndsDay
			ConstraintActivitiesSameStartingTime       []sameStartingTime
	*/

	days_between(tt_data)
	parallel_activities(tt_data)
}

func days_between(tt_data *timetable.TtData) {
	tclist := tt_data.BackendData.(*FetData).time_constraints_list
	for _, c0 := range tt_data.MinDaysBetweenActivities {
		w := weight2fet(c0.Weight)
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
				c.CreateElement("Activity_Id").SetText(fet_activity_index((int(ai))))
			}
			c.CreateElement("MinDays").SetText(strconv.Itoa(c0.DaysBetween))
			c.CreateElement("Active").SetText("true")
			c.CreateElement("Comments").SetText(activities_constraint(
				c0.Constraint, c0.Id, alist))
		}
	}
}

func parallel_activities(tt_data *timetable.TtData) {
	tclist := tt_data.BackendData.(*FetData).time_constraints_list
	for _, c0 := range tt_data.ParallelActivities {
		w := weight2fet(c0.Weight)
		for _, alist := range c0.ActivityLists {
			c := tclist.CreateElement("ConstraintActivitiesSameStartingTime")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Number_of_Activities").SetText(strconv.Itoa(len(alist)))
			for _, ai := range alist {
				c.CreateElement("Activity_Id").SetText(fet_activity_index((int(ai))))
			}
			c.CreateElement("Active").SetText("true")
			c.CreateElement("Comments").SetText(activities_constraint(
				db.C_ParallelCourses, c0.Id, alist))
		}
	}
}

func before_after_hour(tt_data *timetable.TtData) {
	db0 := tt_data.Db
	tclist := tt_data.BackendData.(*FetData).time_constraints_list

	for _, c0 := range db0.Constraints[db.C_BeforeAfterHour] {
		w := weight2fet(c0.Weight)
		data := c0.Data.(*db.BeforeAfterHour)
		timeslots := []preferred_time{}
		if data.After {
			for d := 0; d < tt_data.NDays; d++ {
				for h := data.Hour + 1; h < tt_data.NHours; h++ {
					timeslots = append(timeslots, preferred_time{
						Preferred_Day:  db0.Days[d].GetTag(),
						Preferred_Hour: db0.Hours[h].GetTag(),
					})
				}
			}
		} else {
			for d := 0; d < tt_data.NDays; d++ {
				for h := 0; h < data.Hour; h++ {
					timeslots = append(timeslots, preferred_time{
						Preferred_Day:  db0.Days[d].GetTag(),
						Preferred_Hour: db0.Hours[h].GetTag(),
					})
				}
			}
		}
		for _, course := range data.Courses {
			cinfo, ok := tt_data.Ref2CourseInfo[course]
			if !ok {
				base.Bug.Fatalf("Invalid course: %s\n", course)
			}
			for _, ai := range cinfo.Activities {
				var after string
				if data.After {
					after = "+"
				} else {
					after = "-"
				}
				arg := fmt.Sprintf("%d/%s/%d", ai, after, data.Hour)

				c := tclist.CreateElement("ConstraintActivityPreferredTimeSlots")
				c.CreateElement("Weight_Percentage").SetText(w)
				c.CreateElement("Activity_Id").SetText(fet_activity_index(int(ai)))
				c.CreateElement("Number_of_Preferred_Time_Slots").
					SetText(strconv.Itoa(len(timeslots)))
				for _, t := range timeslots {
					pts := c.CreateElement("Preferred_Time_Slot")
					pts.CreateElement("Preferred_Day").SetText(t.Preferred_Day)
					pts.CreateElement("Preferred_Hour").SetText(t.Preferred_Hour)
				}
				c.CreateElement("Active").SetText("true")
				c.CreateElement("Comments").SetText(param_constraint(
					db.C_BeforeAfterHour, c0.Id, arg))
			}
		}
	}
}
