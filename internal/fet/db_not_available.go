package fet

import (
	"fetrunner/internal/base"
	"strconv"
)

// Convert "DB" constraints to "FET" constraints.

func room_blocked(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	roomix := mapReadInt(constraint.Data, "Room")
	// `notAvailable` is an ordered list of time-slots in which the
	// room is to be regarded as not available for the timetable.
	notAvailable := mapReadTimeSlots(constraint.Data)
	if len(notAvailable) != 0 {
		w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
		cna := fetbuild.space_constraints_list.CreateElement("ConstraintRoomNotAvailableTimes")
		cna.CreateElement("Weight_Percentage").SetText(w1)
		cna.CreateElement("Room").SetText(fetbuild.RoomList[roomix])
		cna.CreateElement("Number_of_Not_Available_Times").
			SetText(strconv.Itoa(len(notAvailable)))
		for _, slot := range notAvailable {
			nat := cna.CreateElement("Not_Available_Time")
			nat.CreateElement("Day").SetText(fetbuild.DayList[slot.Day])
			nat.CreateElement("Hour").SetText(fetbuild.HourList[slot.Hour])
		}
		cna.CreateElement("Active").SetText("true")
		cna.CreateElement("Comments").SetText(comment)

		fetbuild.ConstraintElements[i] = append(
			fetbuild.ConstraintElements[i], cna)
	}
}

func teacher_blocked(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	tix := mapReadInt(constraint.Data, "Teacher")
	// `notAvailable` is an ordered list of time-slots in which the
	// teacher is to be regarded as not available for the timetable.
	notAvailable := mapReadTimeSlots(constraint.Data)
	if len(notAvailable) != 0 {
		w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
		if constraint.Weight == base.MAXWEIGHT {
			// Collect hard blocks for determining lunch-breaks.
			fetbuild.hard_teacher_blocks[tix] = notAvailable
		}
		cna := fetbuild.time_constraints_list.CreateElement("ConstraintTeacherNotAvailableTimes")
		cna.CreateElement("Weight_Percentage").SetText(w1)
		cna.CreateElement("Teacher").SetText(fetbuild.TeacherList[tix])
		cna.CreateElement("Number_of_Not_Available_Times").
			SetText(strconv.Itoa(len(notAvailable)))
		for _, slot := range notAvailable {
			nat := cna.CreateElement("Not_Available_Time")
			nat.CreateElement("Day").SetText(fetbuild.DayList[slot.Day])
			nat.CreateElement("Hour").SetText(fetbuild.HourList[slot.Hour])
		}
		cna.CreateElement("Active").SetText("true")
		cna.CreateElement("Comments").SetText(comment)

		fetbuild.ConstraintElements[i] = append(
			fetbuild.ConstraintElements[i], cna)
	}
}

func class_blocked(
	fetbuild *fet_build,
	i constraintIndex,
	constraint *ttConstraint,
) {
	classix := mapReadInt(constraint.Data, "Class")
	// `notAvailable` is an ordered list of time-slots in which the
	// class is to be regarded as not available for the timetable.
	notAvailable := mapReadTimeSlots(constraint.Data)
	if len(notAvailable) != 0 {
		w1, comment := fetbuild.constraintWeight(i, constraint.Weight)
		if constraint.Weight == base.MAXWEIGHT {
			// Collect hard blocks for determining lunch-breaks.
			fetbuild.hard_class_blocks[classix] = notAvailable
		}
		cna := fetbuild.time_constraints_list.CreateElement("ConstraintStudentsSetNotAvailableTimes")
		cna.CreateElement("Weight_Percentage").SetText(w1)
		cna.CreateElement("Students").SetText(fetbuild.ClassList[classix])
		cna.CreateElement("Number_of_Not_Available_Times").
			SetText(strconv.Itoa(len(notAvailable)))
		for _, slot := range notAvailable {
			nat := cna.CreateElement("Not_Available_Time")
			nat.CreateElement("Day").SetText(fetbuild.DayList[slot.Day])
			nat.CreateElement("Hour").SetText(fetbuild.HourList[slot.Hour])
		}
		cna.CreateElement("Active").SetText("true")
		cna.CreateElement("Comments").SetText(comment)

		fetbuild.ConstraintElements[i] = append(
			fetbuild.ConstraintElements[i], cna)
	}
}
