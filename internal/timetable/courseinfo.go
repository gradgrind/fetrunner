package timetable

import (
	"fetrunner/internal/base"
	"fmt"
	"slices"
	"strings"
)

// Make a shortish string view of a courseInfo – can be useful in tests
func (tt_data *TtData) View(cinfo *courseInfo, db *base.DbTopLevel) string {
	tlist := []string{}
	for _, t := range cinfo.Teachers {
		tlist = append(tlist, db.Teachers[t].GetTag())
	}
	glist := []string{}
	for _, g := range cinfo.Groups {
		glist = append(glist, base.GroupTag(g))
	}
	return fmt.Sprintf("<Course %s/%s:%s>",
		strings.Join(glist, ","),
		strings.Join(tlist, ","),
		cinfo.Subject,
	)
}

// Collect courses (Course and SuperCourse) and their activities.
// Build a list of courseInfo structures.
func (tt_data *TtData) CollectCourses(bdata *base.BaseData) {
	db := bdata.Db
	tt_data.ref2courseInfo = map[nodeRef]*courseInfo{}

	// The list of `TtActivity` items shadows the list of `Activity` items.
	tt_data.ttActivities = make([]*ttActivity, len(db.Activities))
	tt_data.fixedActivities = make([]*base.TimeSlot, len(db.Activities))
	tt_data.ref2ActivityIndex = map[nodeRef]activityIndex{}
	for i, a := range db.Activities {
		tt_data.ref2ActivityIndex[a.Id] = activityIndex(i)
	}

	// *** Gather the SuperCourses. ***
	for _, spc := range db.SuperCourses {
		cref := spc.Id
		groups := []*base.Group{}
		agroups := []atomicIndex{}
		teachers := []teacherIndex{}
		rooms := []roomIndex{}
		crooms := [][]roomIndex{}
		for _, sbc := range spc.SubCourses {
			// Add groups
			for _, gref := range sbc.Groups {
				g, ok := db.GetElement(gref).(*base.Group)
				if !ok {
					panic("Invalid Group ref: " + gref)
				}
				if !slices.Contains(groups, g) {
					groups = append(groups, g)
					agroups = append(agroups,
						tt_data.atomicGroup2Indexes[gref]...)
				}
			}
			// Add teachers
			for _, tref := range sbc.Teachers {
				t, ok := tt_data.teacher2Index[tref]
				if !ok {
					panic("Invalid Teacher ref: " + tref)
				}
				teachers = append(teachers, t)
			}
			// Add rooms
			if sbc.Room != "" {
				r, ok := tt_data.room2Index[sbc.Room]
				if ok {
					rooms = append(rooms, r)
					continue
				}

				// Not a `Room` – it can be a RoomGroup or RoomChoiceGroup

				gr := db.GetElement(sbc.Room)
				rg, ok := gr.(*base.RoomGroup)
				if ok {
					for _, rr := range rg.Rooms {
						r, ok = tt_data.room2Index[rr]
						if !ok {
							panic(fmt.Sprintf(
								"Bug: Unknown room in RoomGroup %s: %s",
								rr, sbc.Room))
						}
						rooms = append(rooms, r)
					}
					continue
				}

				rcg, ok := gr.(*base.RoomChoiceGroup)
				if ok {
					roomlist := []roomIndex{}
					for _, rr := range rcg.Rooms {
						r, ok = tt_data.room2Index[rr]
						if !ok {
							panic(fmt.Sprintf(
								"Bug: Unknown room in RoomChoiceGroup %s: %s",
								rr, sbc.Room))
						}
						roomlist = append(roomlist, r)
					}
					slices.Sort(roomlist)
					// Don't add if it is a duplicate
					for _, rl := range crooms {
						if slices.Equal(rl, roomlist) {
							goto skip
						}
					}
					crooms = append(crooms, roomlist)
				skip:
					continue
				}

				panic("Bug: Expecting room element, found: " + sbc.Room)
			}
		}

		// Eliminate duplicate resources by sorting and then compacting
		slices.Sort(agroups)
		slices.Sort(teachers)
		slices.Sort(rooms)

		sbj, ok := db.GetElement(spc.Subject).(*base.Subject)
		if !ok {
			panic("Invalid Subject ref: " + spc.Subject)
		}

		cinfo := &courseInfo{
			Id:                 cref,
			Subject:            sbj.Tag,
			Groups:             groups,
			AtomicGroupIndexes: slices.Compact(agroups),
			Teachers:           slices.Compact(teachers),
			FixedRooms:         slices.Compact(rooms),
			RoomChoices:        crooms,
			//Activities: set below,
		}

		// Build a `TtActivity` for each `Activity` – they are already sorted
		// with the longest first.
		for _, a := range spc.Activities {
			aix := tt_data.ref2ActivityIndex[a.Id]
			cinfo.Activities = append(cinfo.Activities, aix)
			tt_data.ttActivities[aix] = &ttActivity{
				Id: string(a.Id),
				// Tag: supplied by back-end
				Duration:           a.Duration,
				Subject:            cinfo.Subject,
				Groups:             cinfo.Groups,
				AtomicGroupIndexes: cinfo.AtomicGroupIndexes,
				Teachers:           cinfo.Teachers,
			}
			// Room placement
			tt_data.constraints = append(tt_data.constraints, &ttConstraint{
				Id:     "",
				CType:  base.C_ActivityRooms,
				Weight: base.MAXWEIGHT,
				Data: map[string]any{
					"Activity":    aix,
					"FixedRooms":  cinfo.FixedRooms,
					"RoomChoices": cinfo.RoomChoices,
				},
			})
		}

		// Filter out any "necessary" rooms from the choices
		tt_data.roomChoiceFilter(cinfo, bdata)

		tt_data.courseInfoList = append(
			tt_data.courseInfoList, cinfo)
		tt_data.ref2courseInfo[cref] = cinfo
	}

	// *** Gather the plain Courses. ***
	for _, c := range db.Courses {
		cref := c.Id

		// Get groups
		groups := []*base.Group{}
		agroups := []atomicIndex{}
		for _, gref := range c.Groups {
			g, ok := db.GetElement(gref).(*base.Group)
			if !ok {
				panic("Invalid Group ref: " + gref)
			}
			groups = append(groups, g)
			agroups = append(agroups, tt_data.atomicGroup2Indexes[gref]...)
		}

		// Get teachers
		teachers := []teacherIndex{}
		for _, tref := range c.Teachers {
			t, ok := tt_data.teacher2Index[tref]
			if !ok {
				panic("Invalid Teacher ref: " + tref)
			}
			teachers = append(teachers, t)
		}

		// Get rooms
		rooms := []roomIndex{}
		crooms := [][]roomIndex{}
		if c.Room != "" {
			r, ok := tt_data.room2Index[c.Room]
			if ok {
				rooms = append(rooms, r)
			} else {
				// Not a `Room` – it can be a RoomGroup or RoomChoiceGroup
				gr := db.GetElement(c.Room)
				rg, ok := gr.(*base.RoomGroup)
				if ok {
					for _, rr := range rg.Rooms {
						r, ok = tt_data.room2Index[rr]
						if !ok {
							panic(fmt.Sprintf(
								"Unknown room in RoomGroup %s: %s",
								rr, c.Room))
						}
						rooms = append(rooms, r)
					}
				} else {
					rcg, ok := gr.(*base.RoomChoiceGroup)
					if ok {
						roomlist := []roomIndex{}
						for _, rr := range rcg.Rooms {
							r, ok = tt_data.room2Index[rr]
							if !ok {
								panic(fmt.Sprintf(
									"Unknown room in RoomChoiceGroup %s: %s",
									rr, c.Room))
							}
							roomlist = append(roomlist, r)
						}
						crooms = append(crooms, roomlist)
					} else {
						panic("Expecting room element, found: " + c.Room)
					}
				}
			}
		}

		sbj, ok := db.GetElement(c.Subject).(*base.Subject)
		if !ok {
			panic("Invalid Subject ref: " + c.Subject)
		}

		//fmt.Printf("^^__ %s\n\n", sbj.Tag)

		// Sort and compact lists
		slices.Sort(agroups)
		slices.Sort(teachers) // shouldn't need compacting
		slices.Sort(rooms)    // shouldn't need compacting
		cinfo := &courseInfo{
			Id:                 cref,
			Subject:            sbj.Tag,
			Groups:             groups,
			AtomicGroupIndexes: slices.Compact(agroups),
			Teachers:           teachers,
			FixedRooms:         rooms,
			RoomChoices:        crooms,
			//Activities: set below,
		}

		// Build a `TtActivity` for each `Activity` – they are already sorted
		// with the longest first.
		for _, a := range c.Activities {
			aix := tt_data.ref2ActivityIndex[a.Id]
			cinfo.Activities = append(cinfo.Activities, aix)
			tt_data.ttActivities[aix] = &ttActivity{
				Id: string(a.Id),
				// Tag: supplied by back-end
				Duration:           a.Duration,
				Subject:            cinfo.Subject,
				Groups:             cinfo.Groups,
				AtomicGroupIndexes: cinfo.AtomicGroupIndexes,
				Teachers:           cinfo.Teachers,
			}
			// Room placement
			tt_data.constraints = append(tt_data.constraints, &ttConstraint{
				Id:     "",
				CType:  base.C_ActivityRooms,
				Weight: base.MAXWEIGHT,
				Data: map[string]any{
					"Activity":    aix,
					"FixedRooms":  cinfo.FixedRooms,
					"RoomChoices": cinfo.RoomChoices,
				},
			})
		}

		tt_data.courseInfoList = append(
			tt_data.courseInfoList, cinfo)
		tt_data.ref2courseInfo[cref] = cinfo
	}
}
