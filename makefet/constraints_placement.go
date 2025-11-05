package makefet

import "strconv"

func addPlacementConstraints(fetinfo *fetInfo) {
    tt_data := fetinfo.tt_data

    armap := map[int]struct{}{}
    for _, a0 := range tt_data.HardConstraints[ActivityRooms] {
        armap[int(a0.(*timetable.ActivityRoomConstraint).ActivityIndex)] = struct{}{}
    }

    for _, cinfo := range tt_data.CourseInfoList {
        var rooms []string
        // Set "preferred" rooms, if not blocked.
        if !tt_data.WITHOUT_ROOM_PLACEMENTS {
            rooms = fetinfo.getFetRooms(cinfo)
        }

        //--fmt.Printf("COURSE: %s\n", ttinfo.View(cinfo))
        //--fmt.Printf("   --> %+v\n", rooms)

        // Add the constraints.
        scl := &fetinfo.fetdata.Space_Constraints_List
        tcl := &fetinfo.fetdata.Time_Constraints_List
        for i, l := range cinfo.Activities {
            aid := cinfo.Activities[i]
            _, ok := armap[int(aid)]
            if ok && len(rooms) != 0 {
                scl.ConstraintActivityPreferredRooms = append(
                    scl.ConstraintActivityPreferredRooms,
                    roomChoice{
                        Weight_Percentage:         100,
                        Activity_Id:               aid,
                        Number_of_Preferred_Rooms: len(rooms),
                        Preferred_Room:            rooms,
                        Active:                    true,
                    },
                )
            }
            if l.Day < 0 {
                continue
            }
            if !l.Fixed {
                continue
            }
            tcl.ConstraintActivityPreferredStartingTime = append(
                tcl.ConstraintActivityPreferredStartingTime,
                startingTime{
                    Weight_Percentage:  100,
                    Activity_Id:        aid,
                    Preferred_Day:      strconv.Itoa(l.Day),
                    Preferred_Hour:     strconv.Itoa(l.Hour),
                    Permanently_Locked: l.Fixed,
                    Active:             true,
                },
            )

            // The Rooms field of a Lesson item is not used for building
            // FET input files. All room constraints are handled by
            // "ConstraintActivityPreferredRooms".
        }
    }
}
