package makefet

import (
	"encoding/xml"
)

type preferredSlots struct {
	XMLName                        xml.Name `xml:"ConstraintActivitiesPreferredTimeSlots"`
	Weight_Percentage              string
	Teacher                        string
	Students                       string
	Subject                        string
	Activity_Tag                   string
	Duration                       string
	Number_of_Preferred_Time_Slots int
	Preferred_Time_Slot            []preferredTime
	Active                         bool
	Comments                       string
}

type preferredTime struct {
	//XMLName                       xml.Name `xml:"Preferred_Time_Slot"`
	Preferred_Day  string
	Preferred_Hour string
}

type preferredStarts struct {
	XMLName                            xml.Name `xml:"ConstraintActivitiesPreferredStartingTimes"`
	Weight_Percentage                  string
	Teacher                            string
	Students                           string
	Subject                            string
	Activity_Tag                       string
	Duration                           string
	Number_of_Preferred_Starting_Times int
	Preferred_Starting_Time            []preferredStart
	Active                             bool
	Comments                           string
}

type preferredStart struct {
	//XMLName                       xml.Name `xml:"Preferred_Starting_Time"`
	Preferred_Starting_Day  string
	Preferred_Starting_Hour string
}

type lessonEndsDay struct {
	XMLName           xml.Name `xml:"ConstraintActivityEndsStudentsDay"`
	Weight_Percentage string
	Activity_Id       int
	Active            bool
	Comments          string
}

type activityPreferredTimes struct {
	XMLName                        xml.Name `xml:"ConstraintActivityPreferredTimeSlots"`
	Weight_Percentage              string
	Activity_Id                    int
	Number_of_Preferred_Time_Slots int
	Preferred_Time_Slot            []preferredTime
	Active                         bool
	Comments                       string
}

type sameStartingTime struct {
	XMLName              xml.Name `xml:"ConstraintActivitiesSameStartingTime"`
	Weight_Percentage    string
	Number_of_Activities int
	Activity_Id          []int
	Active               bool
	Comments             string
}

/*
func getExtraConstraints(fetinfo *fetInfo) {
	tclist := &fetinfo.fetdata.Time_Constraints_List
	tt_data := fetinfo.tt_data
	db0 := tt_data.Db
	db0 := tt_data.Db

	fetinfo.handle_teacher_constraints()
	fetinfo.handle_class_constraints()
	fetinfo.handle_room_constraints()


	for _, c := range tt_data.ParallelActivities {
		w := weight2fet(c.Weight)
		for _, alist := range c.ActivityLists {
			aidlist := make([]int, len(alist))
			for i, ai := range alist {
				aidlist[i] = activityIndex2fet(tt_data, ai)
			}
			tclist.ConstraintActivitiesSameStartingTime = append(
				tclist.ConstraintActivitiesSameStartingTime,
				sameStartingTime{
					Weight_Percentage:    w,
					Number_of_Activities: len(aidlist),
					Activity_Id:          aidlist,
					Active:               true,

					Comments: fetinfo.activities_constraint(
						db.C_ParallelCourses, c.Id, alist),
				})
		}
	}

	for _, c := range db0.Constraints[db.C_ActivitiesEndDay] {
		w := weight2fet(c.Weight)
		course := c.Data.(db.NodeRef)
		cinfo := tt_data.Ref2CourseInfo[course]
		for _, ai := range cinfo.Activities {
			tclist.ConstraintActivityEndsStudentsDay = append(
				tclist.ConstraintActivityEndsStudentsDay,
				lessonEndsDay{
					Weight_Percentage: w,
					Activity_Id:       activityIndex2fet(tt_data, ai),
					Active:            true,

					Comments: fetinfo.param_constraint(
						db.C_ActivitiesEndDay,
						c.Id, strconv.Itoa(int(ai))),
				})
		}
	}

	var doubleBlocked []bool
	for _, c := range db0.Constraints[db.C_DoubleActivityNotOverBreaks] {
		if len(doubleBlocked) != 0 {
			base.Error.Fatalln("Constraint DoubleActivityNotOverBreaks" +
				" specified more than once")
		}

		timeslots := []preferredStart{}
		// Note that a double lesson can't start in the last slot of
		// the day.
		doubleBlocked = make([]bool, tt_data.NHours-1)
		hlist := []string{}
		for _, h := range c.Data.([]int) {
			doubleBlocked[h-1] = true
			hlist = append(hlist, strconv.Itoa(h))
		}
		for d := 0; d < tt_data.NDays; d++ {
			for h, bl := range doubleBlocked {
				if !bl {
					timeslots = append(timeslots, preferredStart{
						Preferred_Starting_Day:  day2Tag(db0, d),
						Preferred_Starting_Hour: hour2Tag(db0, h),
					})
				}
			}
		}

		tclist.ConstraintActivitiesPreferredStartingTimes = append(
			tclist.ConstraintActivitiesPreferredStartingTimes,
			preferredStarts{
				Weight_Percentage:                  weight2fet(c.Weight),
				Duration:                           "2",
				Number_of_Preferred_Starting_Times: len(timeslots),
				Preferred_Starting_Time:            timeslots,
				Active:                             true,

				Comments: fetinfo.param_constraint(
					db.C_DoubleActivityNotOverBreaks,
					c.Id, strings.Join(hlist, ",")),
			})
	}

	for _, c := range db0.Constraints[db.C_BeforeAfterHour] {
		w := weight2fet(c.Weight)
		data := c.Data.(*db.BeforeAfterHour)
		timeslots := []preferredTime{}
		if data.After {
			for d := 0; d < tt_data.NDays; d++ {
				for h := data.Hour + 1; h < tt_data.NHours; h++ {
					timeslots = append(timeslots, preferredTime{
						Preferred_Day:  day2Tag(db0, d),
						Preferred_Hour: hour2Tag(db0, h),
					})
				}
			}
		} else {
			for d := 0; d < tt_data.NDays; d++ {
				for h := 0; h < data.Hour; h++ {
					timeslots = append(timeslots, preferredTime{
						Preferred_Day:  day2Tag(db0, d),
						Preferred_Hour: hour2Tag(db0, h),
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
				tclist.ConstraintActivityPreferredTimeSlots = append(
					tclist.ConstraintActivityPreferredTimeSlots,
					activityPreferredTimes{
						Weight_Percentage:              w,
						Activity_Id:                    activityIndex2fet(tt_data, ai),
						Number_of_Preferred_Time_Slots: len(timeslots),
						Preferred_Time_Slot:            timeslots,
						Active:                         true,

						Comments: fetinfo.param_constraint(
							db.C_BeforeAfterHour, c.Id, arg),
					})
			}
		}
	}

	//TODO
	for _, c := range db0.Constraints[db.C_MinHoursFollowing] {
		base.Error.Printf("!!! Constraint not implemented:\n%+v\n", c)
		//w := weight2fet(c.Weight)
		//data := c.Data.(*db.MinHoursFollowing)

		//MinHoursFollowing{
		//	Course1: course1,
		//	Course2: course2,
		//	Hours:   hours,
		//}

		/* It may be better to specify these constraints in a better way!

		   <ConstraintStudentsSetMinGapsBetweenOrderedPairOfActivityTags>
		     <Weight_Percentage>100</Weight_Percentage>
		     <Students>12G</Students>
		     <First_Activity_Tag>tag2</First_Activity_Tag>
		     <Second_Activity_Tag>tag1</Second_Activity_Tag>
		     <MinGaps>1</MinGaps>
		     <Active>true</Active>
		     <Comments></Comments>
		   </ConstraintStudentsSetMinGapsBetweenOrderedPairOfActivityTags>

*/ /*
	}

}
*/
