package timetable

import (
	"fetrunner/internal/base"
	"slices"
)

func (tt_data *TtData) roomChoiceFilter(cinfo *courseInfo, bdata *base.BaseData) {
	necessary := slices.Clone(cinfo.FixedRooms)
	rclist := cinfo.RoomChoices
	failures := false
	// The validity of the rooms in a RoomChoiceGroup has already been checked.
	// They have been ordered while converting to RoomIndexes.
	//fmt.Printf("§§§ FIXED: %+v\n", necessary)
	//fmt.Printf("§§§ CHOICES: %v\n", rclist)
stage1:
	newlist := [][]roomIndex{}
	for i, rc0 := range rclist {
		// Filter out fixed rooms from the choice list
		var rc []roomIndex = nil
		for _, r := range rc0 {
			if !slices.Contains(necessary, r) {
				rc = append(rc, r)
			}
		}
		if len(rc) >= 2 {
			// Sort the elements.
			slices.Sort(rc)
			newlist = append(newlist, rc)
		} else if len(rc) == 1 {
			necessary = append(necessary, rc[0])
			//fmt.Printf("!!! FIXED %d, REPEATING\n", rc[0])
			rclist = append(newlist, rclist[i+1:]...)
			//fmt.Printf("(+): %+v\n", rclist)
			goto stage1
		} else {
			// Drop the choice list.
			//fmt.Printf("§§§ DROPPING %+v\n", rc0)
			failures = true
		}
	}
	rclist = newlist
	//fmt.Printf("§§§++ FIXED: %+v\n", necessary)
	//fmt.Printf("§§§++ CHOICES: %v\n", rclist)

	// Rooms which occur in all possible combinations of the choices can be
	// moved to the "necessary" list. To find them, build the Cartesian product
	// of the choice lists, omitting C-p values with duplicate rooms and
	// also duplicate C-p values generally.
	cp := [][]roomIndex{{}}   // build Cartesian product values here
	newlist = [][]roomIndex{} // collect "accepted" choice lists
	for i, rc := range rclist {
		//fmt.Printf("CP + RC: %+v\n", rc)
		// Add next choice list, extending the entries in `cp`
		newcp := [][]roomIndex{} // build new `cp` here
		for _, cp0 := range cp { // for each C-p value
			for _, r := range rc { // add each room in current choice list
				if !slices.Contains(cp0, r) { // ... if not a duplicate
					cp1 := append(slices.Clone(cp0), r)
					// ... and if the new C-p value is not a duplicate
					slices.Sort(cp1)
					for _, cp2 := range newcp {
						if slices.Equal(cp2, cp1) {
							goto next
						}
					}
					newcp = append(newcp, cp1)
				next:
				}
			}
		}
		//fmt.Printf("§CP++ %+v\n", newcp)
		if len(newcp) == 0 {
			// Report error and skip to next choice list
			//fmt.Printf("--- %+v\n", rc)
			failures = true
			continue
		}
		newlist = append(newlist, rc)
		// Collect rooms which appear in all C-p values
		do_restart := false
		for _, r := range rc {
			for _, cp0 := range newcp {
				if !slices.Contains(cp0, r) {
					goto next2
				}
			}
			// r is in all combinations
			necessary = append(necessary, r)
			//fmt.Printf("!!! FIXED %d\n", r)
			do_restart = true
		next2:
		}
		if do_restart {
			rclist = append(newlist, rclist[i+1:]...)
			//fmt.Printf("Restarting: %+v\n", rclist)
			goto stage1
		}

		cp = newcp
	}

	slices.Sort(necessary)
	cinfo.FixedRooms = necessary
	cinfo.RoomChoices = newlist
	//fmt.Printf("§§§== FIXED: %+v\n", necessary)
	//fmt.Printf("§§§== CHOICES: %v\n", rclist)
	if failures {
		bdata.Logger.Warning(
			"RoomChoiceGroupsFailure: Course %s, room choices may be represented inaccurately",
			tt_data.View(cinfo, bdata.Db))
	}
}
