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

//TODO: Buffer the class view placements, so that fractional tiles can be
// constructed and placed. It would be useful to have ordered lists of
// divisions, if it is possible to derive these from the atomic groups of
// each student group.

func ClassDivisions(last_result *Result, cix int) [][]string {
	clist := last_result.Classes
	cdata := clist[cix]
	//caglist := cdata.AtomicIndexes
	cgrestlist := cdata.Groups
	fmt.Printf("§A\n")

	dglist := []string{}
	daglist := []AtomicIndex{}
	glists := build_divisions(cgrestlist, daglist, dglist)
	// This list can contain elements which are subsets of other elements These
	// should be removed.
	// Sort the group lists alphabetically.
	for _, gl := range glists {
		slices.Sort(gl)
	}
	// Sort the divisions according to list length.
	slices.SortFunc(glists, func(a, b []string) int {
		return cmp.Compare(len(a), len(b))
	})
	// Eliminate divisions which are subsets.
	divs := [][]string{}
loop1:
	for i, gsl := range glists {
		for _, gsl2 := range glists[i+1:] {
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

var bdcount int = 0

func build_divisions(
	cgrestlist []*TtGroup,
	daglist []AtomicIndex,
	dglist []string,
) [][]string {
	dglists := [][]string{}
	cgr := []string{}
	for _, g := range cgrestlist {
		cgr = append(cgr, g.Tag)
	}

	//TODO: The problem with this algorithm is that it explodes when fed a lot of
	// combinable groups.

	bdcount++
	fmt.Printf("§B %d %d %#v\n", bdcount, len(cgrestlist), cgr)
	if bdcount == 100 {
		panic("X")
	}

	for i, g := range cgrestlist {
		if no_intersection(g.AtomicIndexes, daglist) {
			daglist2 := append(slices.Clone(daglist), g.AtomicIndexes...)
			slices.Sort(daglist2)
			dglist2 := append(slices.Clone(dglist), g.Tag)
			new_glist := build_divisions(cgrestlist[i+1:], daglist2, dglist2)
			dglists = append(dglists, new_glist...)
		}
	}
	if len(dglists) == 0 && len(dglist) != 0 {
		dglists = append(dglists, dglist)
	}
	return dglists
}

// TODO: remove duplicates
func Build_divisions2(glist []*TtGroup, ags []ActivityIndex) [][]string {
	dglists := [][]string{}

	// Map each atomic group to the groups including it
	agmap := map[ActivityIndex][]int{}
	// Make a vector of all groups for each group, marking "blocked" ones,
	// i.e. those sharing an atomic group
	gblock := make([][]bool, len(glist))
	for gix, g := range glist {
		for _, ag := range g.AtomicIndexes {
			agmap[ag] = append(agmap[ag], gix)
		}
		gblock[gix] = make([]bool, len(glist))
	}
	for _, gixlist := range agmap {
		for _, gix := range gixlist {
			gb := gblock[gix]
			for _, gix2 := range gixlist {
				if gix2 != gix {
					gb[gix2] = true
				}
			}
		}
	}
	dgilist := [][]int{}
	for _, gixlist := range gblock {
		d := []int{}
		for gix, blocked := range gixlist {
			if !blocked {
				d = append(d, gix)
			}
		}

		if len(d) != 0 {
			//TODO No duplicates ...
			dgilist = append(dgilist, d)
			//fmt.Printf("§+ %+v\n", d)
		}
	}

	for _, gixl := range dgilist {
		gl := []string{}
		for _, gix := range gixl {
			gl = append(gl, glist[gix].Tag)
		}
		dglists = append(dglists, gl)
		fmt.Printf("§+ %+v\n", gl)
	}

	return dglists
}

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
