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

func a2r(r any) NodeRef {
	return NodeRef(r.(string))
}

func a2i(i any) int {
	return int(i.(float64))
}

func a2rr(rr any) []NodeRef {
	rlist := []NodeRef{}
	for _, r := range rr.([]any) {
		rlist = append(rlist, a2r(r))
	}
	return rlist
}

func a2ii(ii any) []int {
	ilist := []int{}
	for _, i := range ii.([]any) {
		ilist = append(ilist, a2i(i))
	}
	return ilist
}

// Read the constraints read from a W365 JSON file into the equivalent
// internal constraints.
func (db0 *W365TopLevel) readConstraints(newdb *db.DbTopLevel) {
	cmap := map[string]string{}
	for _, pair := range ConstraintMap {
		cmap[pair.W365] = pair.Db
	}
	for _, e := range db0.Constraints {
		cw365 := e["Constraint"].(string)
		switch cmap[cw365] {
		case db.C_ActivitiesEndDay:
			newdb.NewActivitiesEndDay(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2r(e["Course"]))
		case db.C_AfterHour:
			newdb.NewAfterHour(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2rr(e["Courses"]),
				a2i(e["Hour"]))
		case db.C_BeforeHour:
			newdb.NewBeforeHour(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2rr(e["Courses"]),
				a2i(e["Hour"]))
		case db.C_AutomaticDifferentDays:
			newdb.NewAutomaticDifferentDays(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				e["ConsecutiveIfSameDay"].(bool))
		case db.C_DaysBetween:
			newdb.NewDaysBetween(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2rr(e["Courses"]),
				a2i(e["DaysBetween"]),
				e["ConsecutiveIfSameDay"].(bool))
		case db.C_DaysBetweenJoin:
			newdb.NewDaysBetweenJoin(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2r(e["Course1"]),
				a2r(e["Course2"]),
				a2i(e["DaysBetween"]),
				e["ConsecutiveIfSameDay"].(bool))
		case db.C_MinHoursFollowing:
			newdb.NewMinHoursFollowing(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2r(e["Course1"]),
				a2r(e["Course2"]),
				a2i(e["Hours"]))
		case db.C_DoubleActivityNotOverBreaks:
			newdb.NewDoubleActivityNotOverBreaks(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2ii(e["Hours"]))
		case db.C_ParallelCourses:
			newdb.NewParallelCourses(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2rr(e["Courses"]))
		default:
			base.Error.Printf(" @W365 ConstraintInvalid: %s\n", cw365)
		}
	}
}
