package makefet

import (
	"fetrunner/internal/base"
	"fetrunner/internal/timetable"
	"fmt"
	"strconv"
	"strings"
)

func (fetbuild *FetBuild) add_placement_constraints(without_rooms bool) {
	tt_data := fetbuild.ttdata
	db := fetbuild.basedata.Db
	rundata := fetbuild.rundata
	tclist := fetbuild.time_constraints_list
	sclist := fetbuild.space_constraints_list

	// Get the time-placements – not that those available in the `TtActivity`
	// items are only the hard constraints, so the list from `db0.Constraints`
	// is used here.
	type start_time struct {
		weight0 int
		weight  string
		day     int
		hour    int
	}
	ai2start := map[int]start_time{}
	for _, c0 := range db.Constraints[base.C_ActivityStartTime] {
		w := rundata.FetWeight(c0.Weight)
		data := c0.Data.(base.ActivityStartTime)
		ai := tt_data.Ref2ActivityIndex[data.Activity]
		ai2start[ai] = start_time{
			weight0: c0.Weight, weight: w, day: data.Day, hour: data.Hour}
	}

	for _, cinfo := range tt_data.CourseInfoList {
		var rooms []string
		// Set "preferred" rooms, if not blocked.
		if !without_rooms {
			rooms = fetbuild.get_fet_rooms(cinfo)
		}

		//--fmt.Printf("COURSE: %s\n", ttinfo.View(cinfo))
		//--fmt.Printf("   --> %+v\n", rooms)

		// Add the constraints.
		for _, ai := range cinfo.Activities {
			a := db.Activities[ai]
			aid := fet_activity_index(ai)
			if len(rooms) != 0 {
				c := sclist.CreateElement("ConstraintActivityPreferredRooms")
				c.CreateElement("Weight_Percentage").SetText("100")
				c.CreateElement("Activity_Id").SetText(aid)
				c.CreateElement("Number_of_Preferred_Rooms").
					SetText(strconv.Itoa(len(rooms)))
				for _, r := range rooms {
					c.CreateElement("Preferred_Room").SetText(r)
				}
				c.CreateElement("Active").SetText("true")

				fetbuild.add_space_constraint(c, param_constraint(
					base.C_SetRooms, a.Id, ai, 100))
			}

			start, ok := ai2start[ai]
			if ok {
				c := tclist.CreateElement("ConstraintActivityPreferredStartingTime")
				c.CreateElement("Weight_Percentage").SetText(start.weight)
				c.CreateElement("Activity_Id").SetText(aid)
				c.CreateElement("Preferred_Day").SetText(rundata.DayIds[start.day].Backend)
				c.CreateElement("Preferred_Hour").SetText(rundata.HourIds[start.hour].Backend)
				c.CreateElement("Permanently_Locked").SetText("true")
				c.CreateElement("Active").SetText("true")

				fetbuild.add_time_constraint(c, param_constraint(
					base.C_SetStartingTime, a.Id, ai, start.weight0))
			}
		}
	}
}

func (fetbuild *FetBuild) get_fet_rooms(cinfo *timetable.CourseInfo) []string {
	// The fet virtual rooms are cached at fetbuild.fet_virtual_rooms.
	var result []string

	// First get the Element Tags used as ids by FET.
	rooms := fetbuild.basedata.Db.Rooms
	rtags := []string{}
	for _, rr := range cinfo.FixedRooms {
		rtags = append(rtags,
			rooms[rr].GetResourceTag())
	}
	rctags := [][]string{}
	for _, rc := range cinfo.RoomChoices {
		rcl := []string{}
		for _, rr := range rc {
			rcl = append(rcl,
				rooms[rr].GetResourceTag())
		}
		rctags = append(rctags, rcl)
	}

	if len(rctags) == 0 && len(rtags) < 2 {
		result = rtags
	} else if len(rctags) == 1 && len(rtags) == 0 {
		result = rctags[0]
	} else {
		// Otherwise a virtual room is necessary.
		srctags := []string{}
		for _, rcl := range rctags {
			srctags = append(srctags, strings.Join(rcl, ","))
		}
		key := strings.Join(rtags, ",") + "+" + strings.Join(srctags, "|")
		vr, ok := fetbuild.fet_virtual_rooms[key]
		if !ok {
			// Make virtual room, using rooms list from above.
			vr = fmt.Sprintf(
				"%s%03d", VIRTUAL_ROOM_PREFIX, len(fetbuild.fet_virtual_rooms)+1)
			// Remember key/value
			fetbuild.fet_virtual_rooms[key] = vr
			nsets := len(rtags) + len(rctags)
			fetbuild.fet_virtual_room_n[vr] = nsets

			vroom := fetbuild.room_list.CreateElement("Room")

			vroom.CreateElement("Name").SetText(vr)
			//vroom.CreateElement("Long_Name").SetText("")
			vroom.CreateElement("Capacity").SetText("30000")
			vroom.CreateElement("Virtual").SetText("true")
			vroom.CreateElement("Number_of_Sets_of_Real_Rooms").
				SetText(strconv.Itoa(nsets))
			// Add necessary rooms
			for _, rt := range rtags {
				vset := vroom.CreateElement("Set_of_Real_Rooms")
				vset.CreateElement("Number_of_Real_Rooms").SetText("1")
				vset.CreateElement("Real_Room").SetText(rt)
			}
			// Add choice lists from above.
			for _, rtl := range rctags {
				vset := vroom.CreateElement("Set_of_Real_Rooms")
				vset.CreateElement("Number_of_Real_Rooms").
					SetText(strconv.Itoa(len(rtl)))
				for _, rt := range rtl {
					vset.CreateElement("Real_Room").SetText(rt)
				}
			}
			//vroom.CreateElement("Comments").SetText("")
		}
		result = []string{vr}
	}
	//--fmt.Printf("   --> %+v\n", result)
	return result
}
