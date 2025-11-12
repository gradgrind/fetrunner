package makefet

import (
	"fetrunner/db"
	"fetrunner/timetable"
	"fmt"
	"strconv"
	"strings"
)

func add_placement_constraints(
	tt_data *timetable.TtData, without_rooms bool,
) {
	db0 := tt_data.Db
	tclist := tt_data.BackendData.(*FetData).time_constraints_list
	sclist := tt_data.BackendData.(*FetData).space_constraints_list

	for _, cinfo := range tt_data.CourseInfoList {
		var rooms []string
		// Set "preferred" rooms, if not blocked.
		if !without_rooms {
			rooms = get_fet_rooms(tt_data, cinfo)
		}

		//--fmt.Printf("COURSE: %s\n", ttinfo.View(cinfo))
		//--fmt.Printf("   --> %+v\n", rooms)

		// Add the constraints.
		for _, ai := range cinfo.Activities {
			a := db0.Activities[ai]
			aid := int(ai) + 1 // fet activities start at 1
			if len(rooms) != 0 {
				c := sclist.CreateElement("ConstraintActivityPreferredRooms")
				c.CreateElement("Weight_Percentage").SetText("100")
				c.CreateElement("Activity_Id").SetText(strconv.Itoa(aid))
				c.CreateElement("Number_of_Preferred_Rooms").
					SetText(strconv.Itoa(len(rooms)))
				for _, r := range rooms {
					c.CreateElement("Preferred_Room").SetText(r)
				}
				c.CreateElement("Active").SetText("true")
				c.CreateElement("Comments").SetText(resource_constraint(
					db.C_SetRooms, "", a.Id))
			}

			tta := tt_data.Activities[ai]
			start := tta.FixedStartTime
			if start != nil {
				c := tclist.CreateElement("ConstraintActivityPreferredStartingTime")
				c.CreateElement("Weight_Percentage").SetText("100")
				c.CreateElement("Activity_Id").SetText(strconv.Itoa(aid))
				c.CreateElement("Preferred_Day").SetText(db0.Days[start.Day].GetTag())
				c.CreateElement("Preferred_Hour").SetText(db0.Hours[start.Hour].GetTag())
				c.CreateElement("Permanently_Locked").SetText("true")
				c.CreateElement("Active").SetText("true")
				c.CreateElement("Comments").SetText(resource_constraint(
					db.C_SetStartingTime, "", a.Id))
			}
		}
	}
}

func get_fet_rooms(
	tt_data *timetable.TtData, cinfo *timetable.CourseInfo,
) []string {
	fetdata := tt_data.BackendData.(*FetData)
	// The fet virtual rooms are cached at fetdata.fet_virtual_rooms.
	var result []string

	// First get the Element Tags used as ids by FET.
	rtags := []string{}
	for _, rr := range cinfo.FixedRooms {
		rtags = append(rtags,
			tt_data.Db.Rooms[rr].GetResourceTag())
	}
	rctags := [][]string{}
	for _, rc := range cinfo.RoomChoices {
		rcl := []string{}
		for _, rr := range rc {
			rcl = append(rcl,
				tt_data.Db.Rooms[rr].GetResourceTag())
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
		vr, ok := fetdata.fet_virtual_rooms[key]
		if !ok {
			// Make virtual room, using rooms list from above.
			vr = fmt.Sprintf(
				"%s%03d", VIRTUAL_ROOM_PREFIX, len(fetdata.fet_virtual_rooms)+1)
			// Remember key/value
			fetdata.fet_virtual_rooms[key] = vr
			nsets := len(rtags) + len(rctags)
			fetdata.fet_virtual_room_n[vr] = nsets

			vroom := fetdata.room_list.CreateElement("Room")

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
