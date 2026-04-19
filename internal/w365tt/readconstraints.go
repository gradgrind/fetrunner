package w365tt

import (
	"fetrunner/internal/base"
)

type DbW365Pair struct {
	Db   string
	W365 string
}

// List providing a mapping between Db constraint names and
// W365 constraint names:
var ConstraintMap []DbW365Pair = []DbW365Pair{
	{base.C_ActivitiesEndDay, "MARGIN_HOUR"},
	{base.C_AfterHour, "AFTER_HOUR"},
	{base.C_BeforeHour, "BEFORE_HOUR"},
	{base.C_AutomaticDifferentDays, "AUTOMATIC_DIFFERENT_DAYS"},
	{base.C_DaysBetween, "DAYS_BETWEEN"},
	{base.C_DaysBetweenJoin, "DAYS_BETWEEN_JOIN"},
	{base.C_MinHoursFollowing, "MIN_HOURS_FOLLOWING"},
	{base.C_DoubleActivityNotOverBreaks, "DOUBLE_LESSON_NOT_OVER_BREAKS"},
	{base.C_ParallelCourses, "PARALLEL_COURSES"},
}

// Parameter-reading functions for the constraints

func a2r(logger *base.Logger, r any) NodeRef {
	rr, ok := r.(string)
	if ok {
		return NodeRef(rr)
	}
	if r != nil {
		base.LogError("Invalid NodeRef in Constraint: %+v", r)
	}
	return ""
}

func a2i(logger *base.Logger, i any) int {
	ii, ok := i.(float64)
	if !ok {
		base.LogError("Invalid number in Constraint: %+v", i)
		return 0
	}
	return int(ii)
}

func a2rr(logger *base.Logger, rr any) []NodeRef {
	rlist := []NodeRef{}
	rrr, ok := rr.([]any)
	if ok {
		for _, r := range rrr {
			rlist = append(rlist, a2r(logger, r))
		}
	} else if rr != nil {
		base.LogError("Invalid NodeRef list in Constraint: %+v", rr)
	}
	return rlist
}

func a2ii(logger *base.Logger, ii any) []int {
	ilist := []int{}
	iii, ok := ii.([]any)
	if ok {
		for _, i := range iii {
			ilist = append(ilist, a2i(logger, i))
		}
	} else if ii != nil {
		base.LogError("Invalid number list in Constraint: %+v", ii)
	}
	return ilist
}

// Read the constraints read from a W365 JSON file into the equivalent
// internal constraints.
func (db0 *W365TopLevel) readConstraints(newdb *base.BaseData) {
	logger := newdb.Logger
	ndb := newdb.Db
	cmap := map[string]string{}
	for _, pair := range ConstraintMap {
		cmap[pair.W365] = pair.Db
	}

	// The following constraints may only appear once:
	automatic_different_days := false
	double_activity_not_over_breaks := false

	for _, e := range db0.Constraints {
		cw365 := e["Constraint"].(string)
		switch cmap[cw365] {
		case base.C_ActivitiesEndDay:
			ndb.NewActivitiesEndDay(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2r(logger, e["Course"]))
		case base.C_AfterHour:
			ndb.NewAfterHour(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2rr(logger, e["Courses"]),
				a2i(logger, e["Hour"]))
		case base.C_BeforeHour:
			ndb.NewBeforeHour(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2rr(logger, e["Courses"]),
				a2i(logger, e["Hour"]))
		case base.C_AutomaticDifferentDays:
			if automatic_different_days {
				base.LogError("Multiple constraints of type %s", base.C_AutomaticDifferentDays)
			} else {
				automatic_different_days = true
				ndb.NewAutomaticDifferentDays(
					a2r(logger, e["Id"]),
					a2i(logger, e["Weight"]),
					e["ConsecutiveIfSameDay"].(bool))
			}
		case base.C_DaysBetween:
			ndb.NewDaysBetween(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2rr(logger, e["Courses"]),
				a2i(logger, e["DaysBetween"]),
				e["ConsecutiveIfSameDay"].(bool))
		case base.C_DaysBetweenJoin:
			ndb.NewDaysBetweenJoin(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2r(logger, e["Course1"]),
				a2r(logger, e["Course2"]),
				a2i(logger, e["DaysBetween"]),
				e["ConsecutiveIfSameDay"].(bool))
		case base.C_MinHoursFollowing:
			ndb.NewMinHoursFollowing(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2r(logger, e["Course1"]),
				a2r(logger, e["Course2"]),
				a2i(logger, e["Hours"]))
		case base.C_DoubleActivityNotOverBreaks:
			if double_activity_not_over_breaks {
				base.LogError("Multiple constraints of type %s", base.C_DoubleActivityNotOverBreaks)
			} else {
				double_activity_not_over_breaks = true
				ndb.NewDoubleActivityNotOverBreaks(
					a2r(logger, e["Id"]),
					a2i(logger, e["Weight"]),
					a2ii(logger, e["Hours"]))
			}
		case base.C_ParallelCourses:
			ndb.NewParallelCourses(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2rr(logger, e["Courses"]))
		default:
			base.LogError(" @W365 ConstraintInvalid: %s", cw365)
		}
	}
}
