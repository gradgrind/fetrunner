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
	{"ActivitiesEndDay", "MARGIN_HOUR"},
	{"BeforeAfterHour", "BEFORE_AFTER_HOUR"},
	{"AutomaticDifferentDays", "AUTOMATIC_DIFFERENT_DAYS"},
	{"DaysBetween", "DAYS_BETWEEN"},
	{"DaysBetweenJoin", "DAYS_BETWEEN_JOIN"},
	{"MinHoursFollowing", "MIN_HOURS_FOLLOWING"},
	{"DoubleActivityNotOverBreaks", "DOUBLE_LESSON_NOT_OVER_BREAKS"},
	{"ParallelCourses", "PARALLEL_COURSES"},
}

// Parameter-reading functions for the constraints

func a2r(r any) Ref {
	return Ref(r.(string))
}

func a2i(i any) int {
	return int(i.(float64))
}

func a2rr(rr any) []Ref {
	rlist := []Ref{}
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
func (db *DbTopLevel) readConstraints(newdb *db.DbTopLevel) {
	cmap := map[string]string{}
	for _, pair := range ConstraintMap {
		cmap[pair.W365] = pair.Db
	}
	for _, e := range db.Constraints {
		cw365 := e["Constraint"].(string)
		switch cmap[cw365] {
		case "ActivitiesEndDay":
			newdb.NewActivitiesEndDay(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2r(e["Course"]))
		case "BeforeAfterHour":
			newdb.NewBeforeAfterHour(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2rr(e["Courses"]),
				e["After"].(bool),
				a2i(e["Hour"]))
		case "AutomaticDifferentDays":
			newdb.NewAutomaticDifferentDays(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				e["ConsecutiveIfSameDay"].(bool))
		case "DaysBetween":
			newdb.NewDaysBetween(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2rr(e["Courses"]),
				a2i(e["DaysBetween"]),
				e["ConsecutiveIfSameDay"].(bool))
		case "DaysBetweenJoin":
			newdb.NewDaysBetweenJoin(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2r(e["Course1"]),
				a2r(e["Course2"]),
				a2i(e["DaysBetween"]),
				e["ConsecutiveIfSameDay"].(bool))
		case "MinHoursFollowing":
			newdb.NewMinHoursFollowing(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2r(e["Course1"]),
				a2r(e["Course2"]),
				a2i(e["Hours"]))
		case "DoubleActivityNotOverBreaks":
			newdb.NewDoubleActivityNotOverBreaks(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2ii(e["Hours"]))
		case "ParallelCourses":
			newdb.NewParallelCourses(
				a2r(e["Id"]),
				a2i(e["Weight"]),
				a2rr(e["Courses"]))
		default:
			base.Error.Printf(" @W365 ConstraintInvalid: %s\n", cw365)
		}
	}
}
