package w365tt

import (
	"fetrunner/base"
	"fetrunner/db"
)

type DbW365Pair struct {
	Db   string
	W365 string
}

// List providing a mapping between Db constraint names and
// W365 constraint names:
var ConstraintMap []DbW365Pair = []DbW365Pair{
	{db.C_ActivitiesEndDay, "MARGIN_HOUR"},
	{db.C_AfterHour, "AFTER_HOUR"},
	{db.C_BeforeHour, "BEFORE_HOUR"},
	{db.C_AutomaticDifferentDays, "AUTOMATIC_DIFFERENT_DAYS"},
	{db.C_DaysBetween, "DAYS_BETWEEN"},
	{db.C_DaysBetweenJoin, "DAYS_BETWEEN_JOIN"},
	{db.C_MinHoursFollowing, "MIN_HOURS_FOLLOWING"},
	{db.C_DoubleActivityNotOverBreaks, "DOUBLE_LESSON_NOT_OVER_BREAKS"},
	{db.C_ParallelCourses, "PARALLEL_COURSES"},
}

// Parameter-reading functions for the constraints

func a2r(logger *base.LogInstance, r any) NodeRef {
	rr, ok := r.(string)
	if ok {
		return NodeRef(rr)
	}
	if r != nil {
		logger.Error("Invalid NodeRef in Constraint: %+v\n", r)
	}
	return ""
}

func a2i(logger *base.LogInstance, i any) int {
	ii, ok := i.(float64)
	if !ok {
		logger.Error("Invalid number in Constraint: %+v\n", i)
		return 0
	}
	return int(ii)
}

func a2rr(logger *base.LogInstance, rr any) []NodeRef {
	rlist := []NodeRef{}
	rrr, ok := rr.([]any)
	if ok {
		for _, r := range rrr {
			rlist = append(rlist, a2r(logger, r))
		}
	} else if rr != nil {
		logger.Error("Invalid NodeRef list in Constraint: %+v\n", rr)
	}
	return rlist
}

func a2ii(logger *base.LogInstance, ii any) []int {
	ilist := []int{}
	iii, ok := ii.([]any)
	if ok {
		for _, i := range iii {
			ilist = append(ilist, a2i(logger, i))
		}
	} else if ii != nil {
		logger.Error("Invalid number list in Constraint: %+v\n", ii)
	}
	return ilist
}

// Read the constraints read from a W365 JSON file into the equivalent
// internal constraints.
func (db0 *W365TopLevel) readConstraints(newdb *db.DbTopLevel) {
	logger := newdb.Logger
	cmap := map[string]string{}
	for _, pair := range ConstraintMap {
		cmap[pair.W365] = pair.Db
	}
	for _, e := range db0.Constraints {
		cw365 := e["Constraint"].(string)
		switch cmap[cw365] {
		case db.C_ActivitiesEndDay:
			newdb.NewActivitiesEndDay(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2r(logger, e["Course"]))
		case db.C_AfterHour:
			newdb.NewAfterHour(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2rr(logger, e["Courses"]),
				a2i(logger, e["Hour"]))
		case db.C_BeforeHour:
			newdb.NewBeforeHour(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2rr(logger, e["Courses"]),
				a2i(logger, e["Hour"]))
		case db.C_AutomaticDifferentDays:
			newdb.NewAutomaticDifferentDays(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				e["ConsecutiveIfSameDay"].(bool))
		case db.C_DaysBetween:
			newdb.NewDaysBetween(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2rr(logger, e["Courses"]),
				a2i(logger, e["DaysBetween"]),
				e["ConsecutiveIfSameDay"].(bool))
		case db.C_DaysBetweenJoin:
			newdb.NewDaysBetweenJoin(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2r(logger, e["Course1"]),
				a2r(logger, e["Course2"]),
				a2i(logger, e["DaysBetween"]),
				e["ConsecutiveIfSameDay"].(bool))
		case db.C_MinHoursFollowing:
			newdb.NewMinHoursFollowing(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2r(logger, e["Course1"]),
				a2r(logger, e["Course2"]),
				a2i(logger, e["Hours"]))
		case db.C_DoubleActivityNotOverBreaks:
			newdb.NewDoubleActivityNotOverBreaks(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2ii(logger, e["Hours"]))
		case db.C_ParallelCourses:
			newdb.NewParallelCourses(
				a2r(logger, e["Id"]),
				a2i(logger, e["Weight"]),
				a2rr(logger, e["Courses"]))
		default:
			logger.Error(" @W365 ConstraintInvalid: %s\n", cw365)
		}
	}
}
