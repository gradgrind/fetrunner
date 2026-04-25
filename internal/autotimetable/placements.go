package autotimetable

import (
	"fmt"
	"slices"
	"strconv"
	"strings"
)

const (
	PF_DAY = iota
	PF_HOUR
	PF_LENGTH
	PF_SUBJECT
	PF_GROUPS
	PF_ATOMICS
	PF_TEACHERS
	PF_ROOMS
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

type PlacementData struct {
	Subject  string
	Day      int
	Hour     int
	Length   int
	Teachers []string
	Groups   []string
	Rooms    []string
	Atomics  []int
}

// Select a group of placements given their indexes.
func placements_selected(last_result *Result, pixlist []int) []*PlacementData {
	activities := last_result.Activities
	teachers := last_result.Teachers
	rooms := last_result.Rooms
	placements := last_result.Placements
	pdlist := []*PlacementData{}
	for _, pix := range pixlist {
		p := placements[pix]
		rlist := []string{}
		for _, ri := range p.Rooms {
			rlist = append(rlist, rooms[ri].Tag)
		}
		a := activities[p.Activity]
		tlist := []string{}
		for _, ti := range a.Teachers {
			tlist = append(tlist, teachers[ti].Tag)
		}
		glist := []string{}
		for _, g := range a.Groups {
			glist = append(glist, g.Tag)
		}
		pdlist = append(pdlist, &PlacementData{
			Subject:  a.Subject,
			Day:      p.Day,
			Hour:     p.Hour,
			Length:   a.Duration,
			Teachers: tlist,
			Groups:   glist,
			Rooms:    rlist,
			Atomics:  a.AtomicGroupIndexes,
		})
	}
	return pdlist
}

func SerializePlacement(p *PlacementData) string {
	aglist := []string{}
	for _, ag := range p.Atomics {
		aglist = append(aglist, strconv.Itoa(ag))
	}
	return fmt.Sprintf("%d:%d:%d:%s:%s:%s:%s:%s",
		p.Day, p.Hour, p.Length, p.Subject,
		strings.Join(p.Groups, ","),
		strings.Join(aglist, ","),
		strings.Join(p.Teachers, ","),
		strings.Join(p.Rooms, ","))
}

func TeacherPlacements(last_result *Result, tix int) []*PlacementData {
	activities := last_result.Activities
	pixlist := []int{}
	for pix, p := range last_result.Placements {
		ai := p.Activity
		a := activities[ai]
		if slices.Contains(a.Teachers, tix) {
			pixlist = append(pixlist, pix)
		}
	}
	return placements_selected(last_result, pixlist)
}

func RoomPlacements(last_result *Result, rix int) []*PlacementData {
	pixlist := []int{}
	for pix, p := range last_result.Placements {
		if slices.Contains(p.Rooms, rix) {
			pixlist = append(pixlist, pix)
		}
	}
	return placements_selected(last_result, pixlist)
}

// Whether a placement is relevant for a class can be determined by the
// atomic groups. This is probably safer, more general, than an attempt to
// extract the class from a group name. However, the group lists could
// be used in a similar way ... if they were provided by all input readers
// (currently not the case for FET).
func ClassPlacements(last_result *Result, cix int) []*PlacementData {
	clist := last_result.Classes
	cdata := clist[cix]
	caglist := cdata.AtomicIndexes
	activities := last_result.Activities
	pixlist := []int{}
	for pix, p := range last_result.Placements {
		ai := p.Activity
		a := activities[ai]
		for _, agi := range a.AtomicGroupIndexes {
			if slices.Contains(caglist, agi) {
				pixlist = append(pixlist, pix)
				break
			}
		}
	}
	return placements_selected(last_result, pixlist)
}
