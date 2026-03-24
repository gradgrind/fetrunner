package fet

import (
	"fmt"
	"strconv"
	"strings"
)

// Convert "base" constraints to "FET" constraints.

func activity_start(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.time_constraints_list.CreateElement("ConstraintActivityPreferredStartingTime")
	c.CreateElement("Weight_Percentage").SetText(w1)
	a := fetbuild.ActivityList[mapReadInt(constraint.Data, "Activity")]
	c.CreateElement("Activity_Id").SetText(a)
	t := mapReadTimeSlot(constraint.Data)
	c.CreateElement("Preferred_Day").SetText(fetbuild.DayList[t.Day])
	c.CreateElement("Preferred_Hour").SetText(fetbuild.HourList[t.Hour])
	c.CreateElement("Permanently_Locked").SetText("true")
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func activity_rooms(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	if fetbuild.no_room_constraints {
		return
	}
	w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
	c := fetbuild.space_constraints_list.CreateElement("ConstraintActivityPreferredRooms")
	c.CreateElement("Weight_Percentage").SetText(w1)
	a := fetbuild.ActivityList[mapReadInt(constraint.Data, "Activity")]
	c.CreateElement("Activity_Id").SetText(a)
	// Get rooms
	rooms_fixed := mapReadIndexList(constraint.Data, "FixedRooms")
	room_choices := mapReadIndexListList(constraint.Data, "RoomChoices")
	rooms := fetbuild.get_fet_rooms(rooms_fixed, room_choices)
	c.CreateElement("Number_of_Preferred_Rooms").
		SetText(strconv.Itoa(len(rooms)))
	for _, r := range rooms {
		c.CreateElement("Preferred_Room").SetText(r)
	}
	c.CreateElement("Active").SetText("true")
	c.CreateElement("Comments").SetText(comment)

	fetbuild.ConstraintElements[i] = append(
		fetbuild.ConstraintElements[i], c)
}

func (fetbuild *fet_build) get_fet_rooms(fixed_rooms []int, room_choices [][]int) []string {
	// The fet virtual rooms are cached at fetbuild.fet_virtual_rooms.
	var result []string
	// First get the Element Tags used as ids by FET.
	rtags := []string{}
	for _, rr := range fixed_rooms {
		rtags = append(rtags, fetbuild.RoomList[rr])
	}
	rctags := [][]string{}
	for _, rc := range room_choices {
		rcl := []string{}
		for _, rr := range rc {
			rcl = append(rcl, fetbuild.RoomList[rr])
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
