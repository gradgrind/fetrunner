package autotimetable

import (
	"fmt"
	"strings"
)

func (attdata *AutoTtData) GetPlacements() []string {
	placements := []string{}
	activities := attdata.lastResult.Activities
	teachers := attdata.lastResult.Teachers
	rooms := attdata.lastResult.Rooms
	for _, p := range attdata.lastResult.Placements {
		ai := p.Activity
		di := p.Day
		hi := p.Hour
		rlist := []string{}
		for _, ri := range p.Rooms {
			rlist = append(rlist, rooms[ri].Tag)
		}
		a := activities[ai]
		sbj := a.Subject
		tlist := []string{}
		for _, ti := range a.Teachers {
			tlist = append(tlist, teachers[ti].Tag)
		}
		glist := []string{}
		for _, g := range a.Groups {
			glist = append(glist, g.Tag)
		}
		nag := len(a.AtomicGroupIndexes)
		l := a.Duration

		placements = append(placements, fmt.Sprintf("%d:%d:%d:%s:%s:%d:%s%s",
			di, hi, l, sbj,
			strings.Join(glist, ","), nag,
			strings.Join(tlist, ","),
			strings.Join(rlist, ",")))
	}

	return placements
}
