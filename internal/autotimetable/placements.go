package autotimetable

import (
	"cmp"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

/*

Each placement represents an activity, which will be placed in one or more
consecutive timetable slots. For the classes the matter is complicated by the
possible division of a class into groups, perhaps in more than one way. This
can make the placement of the corresponding tiles within a timetable slot rather
complicated. Depending on the size allocated to each slot, there will be a
limit to the number of tiles that can be placed legibly within a slot.

TODO ...

At first I will concentrate on timetable displays for single classes, teachers
and rooms. A timetable for a single group is probably not that helpful in many
cases, because a group can contain students from several other groups (from other
class divisions), so it probably wouldn't really be clearer than a view of the
whole class.

At first I should probably develop functions to select the tiles (placements)
to appear in a single timetable. For classes with groups a further function
would be needed to determine order and sizing of the tiles.

Whether to build indexable data structures from the placements, or just search
the list each time?

*/

func SerializePlacement(p *TtActivityPlacement) string {
	rlist := []string{}
	for _, r := range p.Rooms {
		rlist = append(rlist, strconv.Itoa(r))
	}
	return fmt.Sprintf("%d:%d:%d:%s",
		p.Activity, p.Day, p.Hour, strings.Join(rlist, ","))
}

func TeacherPlacements(last_result *Result, tix int) []*TtActivityPlacement {
	activities := last_result.Activities
	plist := []*TtActivityPlacement{}
	for _, p := range last_result.Placements {
		ai := p.Activity
		a := activities[ai]
		if slices.Contains(a.Teachers, tix) {
			plist = append(plist, p)
		}
	}
	return plist
}

func RoomPlacements(last_result *Result, rix int) []*TtActivityPlacement {
	plist := []*TtActivityPlacement{}
	for _, p := range last_result.Placements {
		if slices.Contains(p.Rooms, rix) {
			plist = append(plist, p)
		}
	}
	return plist
}

// Whether a placement is relevant for a class can be determined by the
// atomic groups. This is probably safer, more general, than an attempt to
// extract the class from a group name. However, the group lists could
// be used in a similar way ... if they were provided by all input readers
// (currently not the case for FET).
func ClassPlacements(last_result *Result, cix int) []*TtActivityPlacement {
	plist := []*TtActivityPlacement{}
	clist := last_result.Classes
	cdata := clist[cix]
	caglist := cdata.AtomicIndexes
	activities := last_result.Activities
	for _, p := range last_result.Placements {
		ai := p.Activity
		a := activities[ai]
		for _, agi := range a.AtomicGroupIndexes {
			if slices.Contains(caglist, agi) {
				plist = append(plist, p)
				break
			}
		}
	}
	return plist
}

// TODO: Buffer the class view placements, so that fractional tiles can be
// constructed and placed. It would be useful to have ordered lists of
// divisions, if it is possible to derive these from the atomic groups of
// each student group.
func weekbuffer(last_result *Result) {
	ndays := len(last_result.Days)
	nhours := len(last_result.Hours)
	type ap struct {
		p *TtActivityPlacement
		n int // index, for length > 1 (first part has n == 0)
	}
	//TODO: Maybe make space for the division lists in the array slots?
	type period struct {
		divs [][]string
		aps  []ap
	}
	week := make([][]period, ndays)
	for d := range ndays {
		week[d] = make([]period, nhours)
	}

	activities := last_result.Activities
	plist := []*TtActivityPlacement{}
	for _, p := range plist {
		a := activities[p.Activity]
		for i := range a.Duration {
			week[p.Day][p.Hour].aps = append(week[p.Day][p.Hour+i].aps, ap{p, i})
		}
	}

	//TODO: Discover division and allocate the activities to it?
	// Deal with long activities.
	for _, dvec := range week {
		h := 0

		//TODO: I need to handle length > 1!
		glist := []string{}
		for _, apx := range dvec[h] {
			a := activities[apx.p.Activity]
			aglist := []string{}
			for _, g := range a.Groups {
				//TODO: Special handling for full class?
				//TODO: Filter for just the present class ...

				aglist = append(aglist, g.Tag)
			}
			glist = append(glist, aglist...)
		}
		//TODO: Seek matching division
		divs := [][]string{}
		thisdiv := [][]string{}
		for _, div := range divs {
			if slices.Equal(glist, div) {
				thisdiv = append(thisdiv, div)
			} else if subset(glist, div) {
				thisdiv = append(thisdiv, div)
			}
		}
		if len(thisdiv) == 0 {
			panic("Found no division")
		}
		//TODO ...
	}
}

func ClassDivisions(last_result *Result, cix int) [][]string {
	c := last_result.Classes[cix]
	glist := c.Groups
	if len(glist) == 0 {
		return nil
	}
	// Convert atomic groups to indexes
	ags := c.AtomicIndexes
	ag2agix := map[AtomicIndex]int{}
	for i, ag := range ags {
		ag2agix[ag] = i
	}

	agix2g := make([]string, len(ags)) // map ag indexes to single-ag groups
	// Collect single-ag groups and multi-ag groups
	// Build first (potentially incomplete) division, containing all single-ag groups
	gmulti := []*TtGroup{}
	dglist := []string{}
	for _, g := range glist {
		agl := g.AtomicIndexes
		if len(agl) == 1 {
			agix2g[ag2agix[agl[0]]] = g.Tag
			dglist = append(dglist, g.Tag)
		} else {
			gmulti = append(gmulti, g)
		}
	}
	dglists := [][]string{dglist}

	// Substitute multi-ag groups where possible to build further divisions

	// The div structure contains the blocked agixs and the list of multi-ag groups
	type div struct {
		blocked []bool
		glist   []string
	}

	div0 := &div{blocked: make([]bool, len(ags))}
	dlists := []*div{div0}
	for _, g := range gmulti {
		gagixs := []int{}
		for _, ag := range g.AtomicIndexes {
			gagixs = append(gagixs, ag2agix[ag])
		}
		newdivs := []*div{}
	next_div:
		for _, divx := range dlists {
			// Check compatibility, if OK add a new division
			for _, agix := range gagixs {
				if divx.blocked[agix] {
					continue next_div
				}
			}
			newdiv := &div{
				blocked: slices.Clone(divx.blocked),
				glist:   slices.Clone(divx.glist)}
			for _, agix := range gagixs {
				newdiv.blocked[agix] = true
			}
			newdiv.glist = append(newdiv.glist, g.Tag)
			newdivs = append(newdivs, newdiv)

			// Build group list
			glist := []string{}
			for agix, blk := range newdiv.blocked {
				if !blk {
					gg := agix2g[agix]
					if gg != "" {
						//TODO: flag this division as incomplete? (It might be
						// desirable to show this in a timetable display.)
						glist = append(glist, gg)
					}
				}
			}
			glist = append(glist, newdiv.glist...)
			dglists = append(dglists, glist)
		}
		dlists = append(dlists, newdivs...)
	}

	// `dglists` list can contain elements which are subsets of other elements These
	// should be removed.
	// Sort the group lists alphabetically.
	for _, gl := range dglists {
		slices.Sort(gl)
	}
	// Sort the divisions according to list length.
	slices.SortFunc(dglists, func(a, b []string) int {
		return cmp.Compare(len(a), len(b))
	})
	// Eliminate divisions which are subsets.
	divs := [][]string{}
loop1:
	for i, gsl := range dglists {
		for _, gsl2 := range dglists[i+1:] {
			if len(gsl) < len(gsl2) {
				if subset(gsl2, gsl) {
					continue loop1
				}
			}
		}
		divs = append(divs, gsl)
	}
	return divs
}

/* TODO--?
// Both lists must be sorted (ascending)
func no_intersection(a []int, b []int) bool {
	blen := len(b)
	bix := 0
	for _, ag := range a {
		for {
			if bix == blen {
				return true
			}
			bg := b[bix]
			if bg > ag {
				break
			}
			if bg == ag {
				return false
			}
			bix++
		}
	}
	return true
}
*/

// Both lists must be sorted (ascending)
func subset(super []string, sub []string) bool {
	sublen := len(sub)
	subix := 0
	for _, g := range super {
		for {
			if subix == sublen {
				// All `bg` found
				return true
			}
			bg := sub[subix]
			if bg < g {
				return false
			}
			if bg != g {
				break // get next `g`
			}
			// match, get next `bg`
			subix++
		}
	}
	return false
}
