package makefet

import "fetrunner/db"

func addPlacementConstraints(fetinfo *fetInfo, without_rooms bool) {
	tt_data := fetinfo.tt_data
	db0 := tt_data.Db

	for _, cinfo := range tt_data.CourseInfoList {
		var rooms []string
		// Set "preferred" rooms, if not blocked.
		if !without_rooms {
			rooms = fetinfo.getFetRooms(cinfo)
		}

		//--fmt.Printf("COURSE: %s\n", ttinfo.View(cinfo))
		//--fmt.Printf("   --> %+v\n", rooms)

		// Add the constraints.
		scl := &fetinfo.fetdata.Space_Constraints_List
		tcl := &fetinfo.fetdata.Time_Constraints_List
		for _, ai := range cinfo.Activities {
			a := db0.Activities[ai]
			aid := activityIndex2fet(tt_data, ai)
			if len(rooms) != 0 {
				scl.ConstraintActivityPreferredRooms = append(
					scl.ConstraintActivityPreferredRooms,
					roomChoice{
						Weight_Percentage:         100,
						Activity_Id:               aid,
						Number_of_Preferred_Rooms: len(rooms),
						Preferred_Room:            rooms,
						Active:                    true,
						Comments: resource_constraint(
							"", a.Id, db.C_SetRooms),
					},
				)
			}
			tta := tt_data.Activities[ai]
			start := tta.FixedStartTime
			if start != nil {
				tcl.ConstraintActivityPreferredStartingTime = append(
					tcl.ConstraintActivityPreferredStartingTime,
					startingTime{
						Weight_Percentage:  100,
						Activity_Id:        aid,
						Preferred_Day:      day2Tag(db0, start.Day),
						Preferred_Hour:     hour2Tag(db0, start.Hour),
						Permanently_Locked: true,
						Active:             true,
						Comments: resource_constraint(
							"", a.Id, db.C_SetStartingTime),
					},
				)
			}
		}
	}
}
