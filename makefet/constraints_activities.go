package makefet

import (
	"fetrunner/base"
	"fetrunner/db"
	"fmt"
	"strconv"
	"strings"
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
	tclist := fetbuild.time_constraints_list
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

func (fetbuild *FetBuild) ends_day() {
	tt_data := fetbuild.ttdata
	db0 := tt_data.Db
	tclist := fetbuild.time_constraints_list
	for _, c0 := range db0.Constraints[db.C_ActivitiesEndDay] {
		w := weight2fet(c0.Weight)
		course := c0.Data.(db.NodeRef)
		cinfo := tt_data.Ref2CourseInfo[course]
		for _, ai := range cinfo.Activities {
			c := tclist.CreateElement("ConstraintActivityEndsStudentsDay")
			c.CreateElement("Weight_Percentage").SetText(w)
			c.CreateElement("Activity_Id").SetText(fet_activity_index(int(ai)))
			c.CreateElement("Active").SetText("true")
			c.CreateElement("Comments").SetText(param_constraint(
				db.C_ActivitiesEndDay, c0.Id, strconv.Itoa(int(ai))))
		}
	}
}

func (fetbuild *FetBuild) parallel_activities() {
	tt_data := fetbuild.ttdata
	tclist := fetbuild.time_constraints_list
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

func (fetbuild *FetBuild) before_after_hour() {
	tt_data := fetbuild.ttdata
	db0 := tt_data.Db
	tclist := fetbuild.time_constraints_list

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

func (fetbuild *FetBuild) double_no_break() {
	tt_data := fetbuild.ttdata
	db0 := tt_data.Db
	tclist := fetbuild.time_constraints_list

	var doubleBlocked []bool
	for _, c0 := range db0.Constraints[db.C_DoubleActivityNotOverBreaks] {
		if len(doubleBlocked) != 0 {
			base.Error.Fatalln("Constraint DoubleActivityNotOverBreaks" +
				" specified more than once")
		}
		w := weight2fet(c0.Weight)
		timeslots := []preferred_time{}
		// Note that a double lesson can't start in the last slot of
		// the day.
		doubleBlocked = make([]bool, tt_data.NHours-1)
		hlist := []string{}
		for _, h := range c0.Data.([]int) {
			doubleBlocked[h-1] = true
			hlist = append(hlist, strconv.Itoa(h))
		}
		for d := 0; d < tt_data.NDays; d++ {
			for h, bl := range doubleBlocked {
				if !bl {
					timeslots = append(timeslots, preferred_time{
						Preferred_Day:  db0.Days[d].GetTag(),
						Preferred_Hour: db0.Hours[h].GetTag(),
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
		c.CreateElement("Comments").SetText(param_constraint(
			db.C_DoubleActivityNotOverBreaks, c0.Id, strings.Join(hlist, ",")))
	}
}

/* TODO
for _, c0 := range db0.Constraints[db.C_MinHoursFollowing] {
	base.Error.Printf("!!! Constraint not implemented:\n%+v\n", c0)
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
