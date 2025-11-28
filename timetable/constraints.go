package timetable

import (
	"fetrunner/base"
	"strings"
)

/* The constraints AutomaticDifferentDays, DaysBetween are processed and
 * combined to be replaced by TtDaysBetween constraints. These also gain
 * activity lists to assist in the implementation of the constraint.
 */
type TtDaysBetween struct {
	Id                   NodeRef
	Weight               int
	CType                string
	DaysBetween          int
	ConsecutiveIfSameDay bool
	ActivityLists        [][]ActivityIndex
}

func (c *TtDaysBetween) IsHard() bool {
	// Note that "ConsecutiveIfSameDay" is hard regardless of
	// the weight.
	return c.Weight == base.MAXWEIGHT || c.ConsecutiveIfSameDay
}

/* The ParallelCourses constraints are transformed to TtParallelActivities
 * constraints.
 */
type TtParallelActivities struct {
	Id            NodeRef
	Weight        int
	CType         string
	ActivityLists [][]ActivityIndex
}

func (c *TtParallelActivities) IsHard() bool {
	return c.Weight == base.MAXWEIGHT
}

// `preprocessConstraints` produces the new constraints, which are then
// accessible in `TtData`. It also adds the fixed activity placements to
// the `TtAcivity` items.
func (tt_data *TtData) preprocessConstraints() {
	bdata := tt_data.BaseData
	db := bdata.Db
	logger := bdata.Logger

	// Deal with the fixed activity placements.
	for _, c := range db.Constraints[base.C_ActivityStartTime] {
		if c.IsHard() {
			data := c.Data.(base.ActivityStartTime)
			aix := tt_data.Ref2ActivityIndex[data.Activity]
			tt_data.Activities[aix].FixedStartTime = &base.TimeSlot{
				Day: data.Day, Hour: data.Hour}
		}
	}

	// If an "AutomaticDifferentDays" constraint is present (at most one is
	// permitted), the `auto_weight` and `auto_consec` variables will be set
	// accordingly, otherwise the default weight (`base.MAXWEIGHT`, i.e.
	// a hard constraint) will be used.

	var auto_id NodeRef = ""
	auto_weight := -1
	auto_consec := false
	noauto_ddays := map[NodeRef]struct{}{} // collect default overrides

	cadd := db.Constraints[base.C_AutomaticDifferentDays]
	if len(cadd) != 0 {
		if len(cadd) == 1 {
			cadd0 := cadd[0]
			auto_id = cadd0.Id
			auto_weight = cadd0.Weight
			auto_consec = cadd0.Data.(bool)
		} else {
			logger.Error("!!! %s * %d – only one is permitted\n",
				//TODO: use the source name instead?
				base.C_AutomaticDifferentDays, len(cadd))
		}
	}

	//TODO ... ??? Need "fixed" status for activities ... which is probably
	// loaded from the placements when the TtData is built?

	for _, c := range db.Constraints[base.C_DaysBetween] {
		data := c.Data.(base.DaysBetween)
		for _, course := range data.Courses {
			if data.DaysBetween == 1 || c.Weight == base.MAXWEIGHT {
				// Override default constraint
				noauto_ddays[course] = struct{}{}
			}
			tt_data.MinDaysBetweenActivities = append(
				tt_data.MinDaysBetweenActivities, &TtDaysBetween{
					Id:                   c.Id,
					Weight:               c.Weight,
					CType:                base.C_DaysBetween,
					DaysBetween:          data.DaysBetween,
					ConsecutiveIfSameDay: data.ConsecutiveIfSameDay,
					ActivityLists: tt_data.days_between_activities(
						course, c.Weight, data.ConsecutiveIfSameDay)})
		}
	}

	// Add automatic min-days-between constraints, where not overridden.
	if auto_weight < 0 {
		auto_weight = base.MAXWEIGHT
	}
	for _, cinfo := range tt_data.CourseInfoList {
		cref := cinfo.Id

		if _, ok := noauto_ddays[cref]; len(cinfo.Activities) > 1 && !ok {
			tt_data.MinDaysBetweenActivities = append(
				tt_data.MinDaysBetweenActivities, &TtDaysBetween{
					Id:                   auto_id,
					Weight:               auto_weight,
					CType:                base.C_AutomaticDifferentDays,
					DaysBetween:          1,
					ConsecutiveIfSameDay: auto_consec,
					ActivityLists: tt_data.days_between_activities(
						cref, auto_weight, auto_consec)})
		}
	}

	for _, c := range db.Constraints[base.C_DaysBetweenJoin] {
		data := c.Data.(base.DaysBetweenJoin)
		tt_data.MinDaysBetweenActivities = append(
			tt_data.MinDaysBetweenActivities, &TtDaysBetween{
				Id:                   c.Id,
				Weight:               c.Weight,
				CType:                base.C_DaysBetweenJoin,
				DaysBetween:          data.DaysBetween,
				ConsecutiveIfSameDay: data.ConsecutiveIfSameDay,
				ActivityLists: tt_data.days_between_join_activities(
					data.Course1, data.Course2)})
	}

cloop:
	for _, c := range db.Constraints[base.C_ParallelCourses] {
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
				alen = len(cinfo.Activities)
				alists = make([][]ActivityIndex, alen)
			} else if len(cinfo.Activities) != alen {
				// This is a data error
				clist := []string{}
				for _, cr := range courses {
					clist = append(clist, string(cr))
				}
				logger.Error("Parallel courses have different"+
					" activities: %s\n",
					strings.Join(clist, ","))
				continue cloop
			}
			for j, ai := range cinfo.Activities {
				a := db.Activities[ai]
				if i == 0 {
					footprint = append(footprint, a.Duration)
				} else if a.Duration != footprint[j] {
					// This is a data error
					clist := []string{}
					for _, cr := range courses {
						clist = append(clist, string(cr))
					}
					logger.Error("Parallel courses have activity"+
						" mismatch: %s\n",
						strings.Join(clist, ","))
					continue cloop
				}
				alists[j] = append(alists[j], cinfo.Activities[j])
			}
		}
		// `alists` is now a list of lists of parallel activity indexes.
		tt_data.ParallelActivities = append(
			tt_data.ParallelActivities, &TtParallelActivities{
				Id:            c.Id,
				Weight:        c.Weight,
				CType:         base.C_ParallelCourses,
				ActivityLists: alists,
			})
	}
}

// Construct the activity relationships for a `DaysBetween` constraint.
func (tt_data *TtData) days_between_activities(
	course NodeRef, weight int, consecutiveIfSameDay bool,
) [][]ActivityIndex {
	logger := tt_data.BaseData.Logger
	allist := [][]ActivityIndex{}
	cinfo := tt_data.Ref2CourseInfo[course]
	fixeds := []ActivityIndex{}
	unfixeds := []ActivityIndex{}
	for _, ai := range cinfo.Activities {
		if tt_data.Activities[ai].FixedStartTime != nil {
			fixeds = append(fixeds, ai)
		} else {
			unfixeds = append(unfixeds, ai)
		}
	}

	if len(unfixeds) == 0 || (len(fixeds) == 0 && len(unfixeds) == 1) {
		// No constraints necessary
		//TODO?
		logger.Warning("Ignoring superfluous DaysBetween constraint on"+
			" course:\n  -- %s", tt_data.View(cinfo))
		return allist
	}
	// Collect the activity groups to which the constraint is to be applied
	aidlists := [][]ActivityIndex{}
	if len(fixeds) <= 1 {
		// At most 1 fixed activity, so all activities are relevant
		aidlists = append(aidlists, cinfo.Activities)
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
				//TODO?
				logger.Warning("Course has too many activities for"+
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
	course1 NodeRef, course2 NodeRef,
) [][]ActivityIndex {
	allist := [][]ActivityIndex{}
	cinfo1 := tt_data.Ref2CourseInfo[course1]
	cinfo2 := tt_data.Ref2CourseInfo[course2]
	for _, ai1 := range cinfo1.Activities {
		f1 := tt_data.Activities[ai1].FixedStartTime != nil
		for _, ai2 := range cinfo2.Activities {
			f2 := tt_data.Activities[ai2].FixedStartTime != nil
			if f1 && f2 {
				// both fixed => no constraint
				continue
			}
			allist = append(allist, []ActivityIndex{ai1, ai2})
		}
	}
	return allist
}
