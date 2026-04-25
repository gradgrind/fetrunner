package autotimetable

import (
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
