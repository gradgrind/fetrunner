package makefet

import (
	"fetrunner/timetable"
	"slices"
	"strconv"

	"github.com/beevik/etree"
)

// Generate the fet activities.
func set_activities(fetroot *etree.Element, tt_data *timetable.TtData) {
	db0 := tt_data.Db

	fetactivities := fetroot.CreateElement("Activities_List")
	for ai, tt_activity := range tt_data.Activities {
		fetactivity := fetactivities.CreateElement("Activity")
		// The fet activities start at Id = 1
		fetactivity.CreateElement("Id").SetText(strconv.Itoa(ai + 1))

		cinfo := tt_data.CourseInfoList[tt_activity.CourseInfo]

		// Teachers
		tlist := []string{}
		for _, ti := range cinfo.Teachers {
			tlist = append(tlist, tt_data.Db.Teachers[ti].GetTag())
		}
		slices.Sort(tlist)
		for _, t := range tlist {
			fetactivity.CreateElement("Teacher").SetText(t)
		}

		// Subject
		fetactivity.CreateElement("Subject").SetText(cinfo.Subject)

		// Student groups
		glist := []string{}
		for _, cg := range cinfo.Groups {
			glist = append(glist, fetGroupTag(cg))
		}
		slices.Sort(glist)
		for _, g := range glist {
			fetactivity.CreateElement("Students").SetText(g)
		}

		//TODO? Activity_Tag: tag (can be more than one of these)

		fetactivity.CreateElement("Active").SetText("true")

		// Get the total duration for this course.
		totalDuration := 0
		for _, aix := range cinfo.Activities {
			totalDuration += db0.Activities[aix].Duration
		}
		fetactivity.CreateElement("Total_Duration").
			SetText(strconv.Itoa(totalDuration))

		// Start FET activity indexes at 1
		agid := 0 // Activity_Group_Id
		if len(cinfo.Activities) > 1 {
			agid = int(cinfo.Activities[0]) + 1
		}
		a := db0.Activities[ai]
		fetactivity.CreateElement("Duration").
			SetText(strconv.Itoa(a.Duration))
		fetactivity.CreateElement("Activity_Group_Id").
			SetText(strconv.Itoa(agid))
		fetactivity.CreateElement("Comments").
			SetText(string(a.GetRef()))
	}
}
