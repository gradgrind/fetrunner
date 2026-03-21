package timetable

import (
	"fetrunner/internal/base"
	"strings"
)

// TODO: Prepare the DB constraints for TtData.
func (tt_data *TtData) prepare_constraints(constraint_map map[string][]*base.BaseConstraint) {
	tt_data.activity_constraints(constraint_map)

	// ...

	//TODO

}

func (tt_data *TtData) prepare_days_between(
	bdata *base.BaseData,
	constraint_map map[string][]*base.BaseConstraint,
) {
	// If an "AutomaticDifferentDays" constraint is present (at most one is
	// permitted), the `auto_weight` and `auto_consec` variables will be set
	// accordingly, otherwise the default weight (`base.MAXWEIGHT`, i.e.
	// a hard constraint) will be used.

	auto_id := ""
	auto_weight := -1
	auto_consec := false
	noauto_ddays := map[nodeRef]struct{}{} // collect default overrides

	cadd := constraint_map[base.C_AutomaticDifferentDays]
	if len(cadd) == 1 {
		cadd0 := cadd[0]
		auto_id = string(cadd0.Id)
		auto_weight = cadd0.Weight
		auto_consec = cadd0.Data.(bool)
	} else if len(cadd) != 0 {
		// Multiple entries should not be possible.
		panic("Multiple constraint type " + base.C_AutomaticDifferentDays)
	}
	delete(constraint_map, base.C_AutomaticDifferentDays)

	for _, c := range constraint_map[base.C_DaysBetween] {
		data := c.Data.(base.DaysBetween)
		for _, course := range data.Courses {
			if data.DaysBetween == 1 || c.Weight == base.MAXWEIGHT {
				// Override default constraint
				noauto_ddays[course] = struct{}{}
			}
			tt_data.constraints = append(tt_data.constraints, &constraint{
				Id:     string(c.Id),
				CType:  base.C_DaysBetween,
				Weight: c.Weight,
				Data: map[string]any{
					"DaysBetween":          data.DaysBetween,
					"ConsecutiveIfSameDay": data.ConsecutiveIfSameDay,
					"ActivityLists": tt_data.days_between_activities(
						course, c.Weight, data.ConsecutiveIfSameDay, bdata),
				},
			})
		}
	}
	delete(constraint_map, base.C_DaysBetween)

	// Add automatic min-days-between constraints, where not overridden.
	if auto_weight < 0 {
		auto_weight = base.MAXWEIGHT
	}
	for _, cinfo := range tt_data.CourseInfoList {
		cref := cinfo.Id
		if _, ok := noauto_ddays[cref]; len(cinfo.Activities) > 1 && !ok {
			tt_data.constraints = append(tt_data.constraints, &constraint{
				Id:     auto_id,
				CType:  base.C_AutomaticDifferentDays,
				Weight: auto_weight,
				Data: map[string]any{
					"DaysBetween":          1,
					"ConsecutiveIfSameDay": auto_consec,
					"ActivityLists": tt_data.days_between_activities(
						cref, auto_weight, auto_consec, bdata),
				},
			})
		}
	}

	for _, c := range constraint_map[base.C_DaysBetweenJoin] {
		data := c.Data.(base.DaysBetweenJoin)

		tt_data.constraints = append(tt_data.constraints, &constraint{
			Id:     string(c.Id),
			CType:  base.C_DaysBetweenJoin,
			Weight: c.Weight,
			Data: map[string]any{
				"DaysBetween":          data.DaysBetween,
				"ConsecutiveIfSameDay": data.ConsecutiveIfSameDay,
				"ActivityLists": tt_data.days_between_join_activities(
					data.Course1, data.Course2),
			},
		})
	}
	delete(constraint_map, base.C_DaysBetweenJoin)
}

// Parallel courses.
func (tt_data *TtData) prepare_parallels(
	bdata *base.BaseData,
	constraint_map map[string][]*base.BaseConstraint,
) {
	logger := bdata.Logger
	db := bdata.Db

cloop:
	for _, c := range constraint_map[base.C_ParallelCourses] {
		courses := c.Data.([]nodeRef)
		// The courses must have the same number of activities and the
		// lengths of the corresponding activities must also be the same.

		// Check activity lengths
		footprint := []int{}         // activity durations
		var alen int = 0             // number of activities in each course
		var alists [][]activityIndex // collect the parallel activities
		for i, cref := range courses {
			cinfo := tt_data.Ref2CourseInfo[cref]
			if i == 0 {
				alen = len(cinfo.Activities)
				alists = make([][]activityIndex, alen)
			} else if len(cinfo.Activities) != alen {
				// This is a data error
				clist := []string{}
				for _, cr := range courses {
					clist = append(clist, string(cr))
				}
				logger.Error(
					"Parallel courses have different activities: %s",
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
					logger.Error(
						"Parallel courses have activity mismatch: %s",
						strings.Join(clist, ","))
					continue cloop
				}
				alists[j] = append(alists[j], cinfo.Activities[j])
			}
		}
		// `alists` is now a list of lists of parallel activity indexes.

		//TODO: Separate constraint for each activity list?

		tt_data.constraints = append(tt_data.constraints, &constraint{
			Id:     string(c.Id),
			CType:  base.C_ParallelCourses,
			Weight: c.Weight,
			Data: map[string]any{
				"ActivityLists": alists,
			},
		})
	}
	delete(constraint_map, base.C_ParallelCourses)
}

// Construct the activity relationships for a `DaysBetween` constraint.
func (tt_data *TtData) days_between_activities(
	course nodeRef, weight int, consecutiveIfSameDay bool, bdata *base.BaseData,
) [][]activityIndex {
	logger := bdata.Logger
	allist := [][]activityIndex{}
	cinfo := tt_data.Ref2CourseInfo[course]
	fixeds := []activityIndex{}
	unfixeds := []activityIndex{}
	for _, ai := range cinfo.Activities {
		if tt_data.TtActivities[ai].fixedStartTime != nil {
			fixeds = append(fixeds, ai)
		} else {
			unfixeds = append(unfixeds, ai)
		}
	}

	if len(unfixeds) == 0 || (len(fixeds) == 0 && len(unfixeds) == 1) {
		// No constraints necessary
		//TODO?
		logger.Warning(
			"Ignoring superfluous DaysBetween constraint on course:\n  -- %s",
			tt_data.View(cinfo, bdata.Db))
		return allist
	}
	// Collect the activity groups to which the constraint is to be applied
	aidlists := [][]activityIndex{}
	if len(fixeds) <= 1 {
		// At most 1 fixed activity, so all activities are relevant
		aidlists = append(aidlists, cinfo.Activities)
	} else {
		// Multiple fixed activities, at least one unfixed one:
		for _, aidf := range fixeds {
			for _, aidu := range unfixeds {
				aidlists = append(aidlists, []activityIndex{aidf, aidu})
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
			if len(alist) > tt_data.ndays {
				//TODO?
				logger.Warning(
					"Course has too many activities for DifferentDays constraint:\n  -- %s",
					tt_data.View(cinfo, bdata.Db))
				continue
			}
			allist = append(allist, alist)
		}
	}
	return allist
}

// Construct the activity relationships for a `DaysBetweenJoin` constraint.
func (tt_data *TtData) days_between_join_activities(
	course1 nodeRef, course2 nodeRef,
) [][]activityIndex {
	allist := [][]activityIndex{}
	cinfo1 := tt_data.Ref2CourseInfo[course1]
	cinfo2 := tt_data.Ref2CourseInfo[course2]
	for _, ai1 := range cinfo1.Activities {
		f1 := tt_data.TtActivities[ai1].fixedStartTime != nil
		for _, ai2 := range cinfo2.Activities {
			f2 := tt_data.TtActivities[ai2].fixedStartTime != nil
			if f1 && f2 {
				// both fixed => no constraint
				continue
			}
			allist = append(allist, []activityIndex{ai1, ai2})
		}
	}
	return allist
}
