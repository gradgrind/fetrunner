package w365tt

import (
	"fetrunner/internal/base"
	"strconv"
)

func (dbi *W365TopLevel) readClasses(newdb *base.BaseData) {
	logger := newdb.Logger
	// Every Class-Group must be within one – and only one – Class-Division.
	// To handle that, the Group references are first gathered here. Then,
	// when a Group is "used" it is flagged. At the end, any unused Groups
	// can be found and reported.
	pregroups := map[NodeRef]bool{}
	for _, n := range dbi.Groups {
		pregroups[n.Id] = false
	}

	dbi.GroupRefMap = map[NodeRef]NodeRef{}
	for _, e := range dbi.Classes {
		// Get the divisions and flag their Groups.
		divs := []base.Division{}
	dloop:
		for i, wdiv := range e.Divisions {
			dname := wdiv.Name
			if dname == "" {
				dname = "#div" + strconv.Itoa(i+1)
			}
			glist := []NodeRef{}
			for _, g := range wdiv.Groups {
				// get Tag
				flag, ok := pregroups[g]
				if ok {
					if flag {
						logger.Error(
							"Group Defined in multiple Divisions:\n  -- %s", g)
						continue dloop
					}
					// Flag Group and add to division's group list
					pregroups[g] = true
					glist = append(glist, g)
				} else {
					logger.Error(
						"Unknown Group in Class %s, Division %s:\n  -- %s",
						e.Tag, wdiv.Name, g)
				}
			}
			// Accept Divisions which have too few Groups at this stage.
			if len(glist) < 2 {
				logger.Warning(
					"In Class %s, not enough valid Groups (>1) in Division %s",
					e.Tag, wdiv.Name)
			}
			divs = append(divs, base.Division{
				Name:   dname,
				Groups: glist,
			})
		}

		// Add a Group for the whole class (not provided by W365).
		classGroup := newdb.NewGroup("")
		classGroup.Tag = ""
		dbi.GroupRefMap[e.Id] = classGroup.Id

		n := newdb.NewClass(e.Id)
		n.Tag = e.Tag
		n.Year = e.Year
		n.Letter = e.Letter
		n.Name = e.Name
		n.Divisions = divs
		n.ClassGroup = classGroup.Id

		// +++ Add constraints ...
		ndb := newdb.Db

		// MaxAfternoons = 0 has a special meaning (all blocked), so the
		// corresponding constraint is not needed, see `handleZeroAfternoons`.
		amax := e.MaxAfternoons
		if amax > 0 {
			ndb.NewClassMaxAfternoons(
				"", base.MAXWEIGHT, n.Id, amax)
		}
		// Not available times – add all afternoons if amax == 0
		tsl := dbi.handleZeroAfternoons(e.NotAvailable, amax)
		if len(tsl) != 0 {
			// Add a constraint
			ndb.NewClassNotAvailable("", base.MAXWEIGHT, n.Id, tsl)
		}

		// MinActivitiesPerDay
		if e.MinLessonsPerDay > 0 {
			ndb.NewClassMinActivitiesPerDay(
				"", base.MAXWEIGHT, n.Id, e.MinLessonsPerDay)
		}
		// MaxActivitiesPerDay
		if e.MaxLessonsPerDay > 0 {
			ndb.NewClassMaxActivitiesPerDay(
				"", base.MAXWEIGHT, n.Id, e.MaxLessonsPerDay)
		}
		// MaxGapsPerDay
		if e.MaxGapsPerDay >= 0 {
			ndb.NewClassMaxGapsPerDay(
				"", base.MAXWEIGHT, n.Id, e.MaxGapsPerDay)
		}
		// MaxGapsPerWeek
		if e.MaxGapsPerWeek >= 0 {
			ndb.NewClassMaxGapsPerWeek(
				"", base.MAXWEIGHT, n.Id, e.MaxGapsPerWeek)
		}
		// LunchBreak
		if e.LunchBreak {
			ndb.NewClassLunchBreak(
				"", base.MAXWEIGHT, n.Id)
		}
		// ForceFirstHour
		if e.ForceFirstHour {
			ndb.NewClassForceFirstHour(
				"", base.MAXWEIGHT, n.Id)
		}
	}

	// Copy Groups.
	for _, n := range dbi.Groups {
		if pregroups[n.Id] {
			g := newdb.NewGroup(n.Id)
			g.Tag = n.Tag
			dbi.GroupRefMap[n.Id] = n.Id // mapping to itself is correct!
		} else {
			logger.Error("Group not in Division, removing: %s", n.Id)
		}
	}
}
