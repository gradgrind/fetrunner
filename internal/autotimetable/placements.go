package autotimetable

import (
	"fmt"
	"strconv"
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
		aglist := []string{}
		for _, ag := range a.AtomicGroupIndexes {
			aglist = append(aglist, strconv.Itoa(ag))
		}
		placements = append(placements, fmt.Sprintf("%d:%d:%d:%s:%s:%s:%s:%s",
			di, hi, a.Duration, sbj,
			strings.Join(glist, ","),
			strings.Join(aglist, ","),
			strings.Join(tlist, ","),
			strings.Join(rlist, ",")))
	}

	return placements
}
