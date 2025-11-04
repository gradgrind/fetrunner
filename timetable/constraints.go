package timetable

import (
	"fetrunner/base"
	"fetrunner/db"
	"strings"
)

/* The constraints AutomaticDifferentDays, DaysBetween are processed and
 * combined to be replaced by TtDaysBetween constraints. These also gain
 * activity lists to assist in the implementation of the constraint.
 */
type TtDaysBetween struct {
	Id                   NodeRef
	Weight               int
	DaysBetween          int
	ConsecutiveIfSameDay bool
	ActivityLists        [][]ActivityIndex
}

func (c *TtDaysBetween) IsHard() bool {
	// Note that "ConsecutiveIfSameDay" is hard regardless of
	// the weight.
	return c.Weight == db.MAXWEIGHT || c.ConsecutiveIfSameDay
}

/* The DaysBetweenJoin constraints are converted to TtDaysBetweenJoin and
 * gain activity lists to assist in the implementation of the constraint.
 */
type TtDaysBetweenJoin struct {
	Constraint           string
	Weight               int
	Course1              NodeRef // Course or SuperCourse
	Course2              NodeRef // Course or SuperCourse
	DaysBetween          int
	ConsecutiveIfSameDay bool
	ActivityLists        [][]ActivityIndex
}

func (c *TtDaysBetweenJoin) IsHard() bool {
	// Note that "ConsecutiveIfSameDay" is hard regardless of
	// the weight.
	return c.Weight == db.MAXWEIGHT || c.ConsecutiveIfSameDay
}

/* The ParallelCourses constraints are transformed to TtParallelActivities
 * constraints.
 */
type TtParallelActivities struct {
	Id         NodeRef
	Weight     int
	Activities []ActivityIndex
}

func (c *TtParallelActivities) IsHard() bool {
	return c.Weight == db.MAXWEIGHT
}

// `preprocessConstraints` produces the new constraints.
// The result is a map, constraint-type -> list of constraints.
func (tt_data *TtData) preprocessConstraints() {
	db0 := tt_data.Db

	// If an "AutomaticDifferentDays" constraint is present (at most one is
	// permitted), the `auto_weight` and `auto_consec` variables will be set
	// accordingly, otherwise the default weight (`db.MAXWEIGHT`, i.e.
	// a hard constraint) will be used.

	var auto_id NodeRef = ""
	auto_weight := -1
	auto_consec := false
	noauto_ddays := map[NodeRef]struct{}{} // collect default overrides

	cadd := db0.Constraints[db.C_AutomaticDifferentDays]
	if len(cadd) != 0 {
		if len(cadd) == 1 {
			cadd0 := cadd[0]
			auto_id = cadd0.Id
			auto_weight = cadd0.Weight
			auto_consec = cadd0.Data.(bool)
		} else {
			base.Error.Printf("!!! %s * %d – only one is permitted\n",
				//TODO: use the source name instead?
				db.C_AutomaticDifferentDays, len(cadd))
		}
	}

	//TODO ... ??? Need "fixed" status for activities ... which is probably
	// loaded from the placements when the TtData is built?

	for _, c := range db0.Constraints[db.C_DaysBetween] {
		data := c.Data.(db.DaysBetween)
		for _, course := range data.Courses {
			if data.DaysBetween == 1 || c.Weight == db.MAXWEIGHT {
				// Override default constraint
				noauto_ddays[course] = struct{}{}
			}
			tt_data.MinDaysBetweenActivities = append(
				tt_data.MinDaysBetweenActivities, &TtDaysBetween{
					Id:                   c.Id,
					Weight:               c.Weight,
					DaysBetween:          data.DaysBetween,
					ConsecutiveIfSameDay: data.ConsecutiveIfSameDay,
					ActivityLists: tt_data.days_between_activities(
						course, c.Weight, data.ConsecutiveIfSameDay)})
		}
	}

	// Add automatic min-days-between constraints, where not overridden.
	if auto_weight < 0 {
		auto_weight = db.MAXWEIGHT
	}
	for _, cinfo := range tt_data.CourseInfoList {
		cref := cinfo.Id

		if _, ok := noauto_ddays[cref]; len(cinfo.TtActivities) > 1 && !ok {
			tt_data.MinDaysBetweenActivities = append(
				tt_data.MinDaysBetweenActivities, &TtDaysBetween{
					Id:                   auto_id,
					Weight:               auto_weight,
					DaysBetween:          1,
					ConsecutiveIfSameDay: auto_consec,
					ActivityLists: tt_data.days_between_activities(
						cref, auto_weight, auto_consec)})
		}
	}

	for _, c := range db0.Constraints[db.C_DaysBetweenJoin] {
		data := c.Data.(db.DaysBetweenJoin)
		tt_data.MinDaysBetweenActivities = append(
			tt_data.MinDaysBetweenActivities, &TtDaysBetween{
				Id:                   c.Id,
				Weight:               c.Weight,
				DaysBetween:          data.DaysBetween,
				ConsecutiveIfSameDay: data.ConsecutiveIfSameDay,
				ActivityLists: tt_data.days_between_join_activities(
					data.Course1, data.Course2,
					c.Weight, data.ConsecutiveIfSameDay)})
	}

	for _, c := range db0.Constraints[db.C_DaysBetweenJoin] {
		courses := c.Data.([]NodeRef)
		// The courses must have the same number of activities and the
		// lengths of the corresponding activities must also be the same.

		// Check activity lengths
		footprint := []int{}         // activity durations
		var alen int = 0             // number of activities in each course
		var alists [][]ActivityIndex // collect the parallel activities
		for i, cref := range courses {
			cinfo := tt_data.Ref2CourseInfo[cref]
			if i == 0 {
				alen = len(cinfo.TtActivities)
				alists = make([][]ActivityIndex, alen)
			} else if len(cinfo.TtActivities) != alen {
				//TODO: This is a data error
				clist := []string{}
				for _, cr := range courses {
					clist = append(clist, string(cr))
				}
				base.Error.Fatalf("Parallel courses have different"+
					" activities: %s\n",
					strings.Join(clist, ","))
			}
			for j, ai := range cinfo.TtActivities {
				a := tt_data.Activities[ai]
				if i == 0 {
					footprint = append(footprint, a.Activity.Duration)
				} else if a.Activity.Duration != footprint[j] {
					//TODO: This is a data error
					clist := []string{}
					for _, cr := range courses {
						clist = append(clist, string(cr))
					}
					base.Error.Fatalf("Parallel courses have activity"+
						" mismatch: %s\n",
						strings.Join(clist, ","))
				}
				alists[j] = append(alists[j], cinfo.TtActivities[j])
			}
		}
		// `alists` is now a list of lists of parallel activity indexes.
		for _, alist := range alists {
			tt_data.ParallelActivities = append(
				tt_data.ParallelActivities, &TtParallelActivities{
					Id:         c.Id,
					Weight:     c.Weight,
					Activities: alist,
				})
		}
	}
}

// Construct the activity relationships for a `DaysBetween` constraint.
func (tt_data *TtData) days_between_activities(
	course NodeRef, weight int, consecutiveIfSameDay bool,
) [][]ActivityIndex {
	allist := [][]ActivityIndex{}
	cinfo := tt_data.Ref2CourseInfo[course]
	fixeds := []ActivityIndex{}
	unfixeds := []ActivityIndex{}
	for _, ai := range cinfo.TtActivities {
		if tt_data.Activities[ai].Fixed {
			fixeds = append(fixeds, ai)
		} else {
			unfixeds = append(unfixeds, ai)
		}
	}

	if len(unfixeds) == 0 || (len(fixeds) == 0 && len(unfixeds) == 1) {
		// No constraints necessary
		//TODO
		base.Warning.Printf("Ignoring superfluous DaysBetween constraint on"+
			" course:\n  -- %s", tt_data.View(cinfo))
		return allist
	}
	// Collect the activity groups to which the constraint is to be applied
	aidlists := [][]ActivityIndex{}
	if len(fixeds) <= 1 {
		// At most 1 fixed activity, so all activities are relevant
		aidlists = append(aidlists, cinfo.TtActivities)
	} else {
		// Multiple fixed activities, at least one unfixed one:
		for _, aidf := range fixeds {
			for _, aidu := range unfixeds {
				aidlists = append(aidlists, []ActivityIndex{aidf, aidu})
			}
		}
		if len(unfixeds) > 1 {
			aidlists = append(aidlists, unfixeds)
		}
	}
	// Add the constraints as `MinDaysBetweenActivities`
	if weight != 0 || consecutiveIfSameDay {
		// Add constraint
		for _, alist := range aidlists {
			if len(alist) > tt_data.NDays {
				//TODO
				base.Warning.Printf("Course has too many activities for"+
					"DifferentDays constraint:\n  -- %s\n",
					tt_data.View(cinfo))
				continue
			}
			allist = append(allist, alist)
		}
	}
	return allist
}

// Construct the activity relationships for a `DaysBetweenJoin` constraint.
func (tt_data *TtData) days_between_join_activities(
	course1 NodeRef, course2 NodeRef, weight int, consecutiveIfSameDay bool,
) [][]ActivityIndex {
	allist := [][]ActivityIndex{}
	cinfo1 := tt_data.Ref2CourseInfo[course1]
	cinfo2 := tt_data.Ref2CourseInfo[course2]
	for _, ai1 := range cinfo1.TtActivities {
		f1 := tt_data.Activities[ai1].Fixed
		for _, ai2 := range cinfo2.TtActivities {
			f2 := tt_data.Activities[ai2].Fixed
			if f1 && f2 {
				// both fixed => no constraint
				continue
			}
			allist = append(allist, []ActivityIndex{ai1, ai2})
		}
	}
	return allist
}
