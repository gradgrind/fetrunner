package makefet

import (
	"slices"
	"strconv"
)

func fet_activity_index(aix int) string {
	return strconv.Itoa(aix + 1) // the FET activity Ids start at 1
}

// Generate the fet activities.
func (fetbuild *FetBuild) set_activities() {
	tt_data := fetbuild.ttdata
	db0 := tt_data.Db
	rundata := fetbuild.rundata

	fetactivities := fetbuild.fetroot.CreateElement("Activities_List")
	for ai, tt_activity := range tt_data.Activities {
		fetactivity := fetactivities.CreateElement("Activity")
		// The fet activities start at Id = 1
		aid := fet_activity_index(ai)
		fetactivity.CreateElement("Id").SetText(aid)

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
		agid := "0" // Activity_Group_Id
		if len(cinfo.Activities) > 1 {
			agid = fet_activity_index(cinfo.Activities[0])
		}
		a := db0.Activities[ai]
		fetactivity.CreateElement("Duration").
			SetText(strconv.Itoa(a.Duration))
		fetactivity.CreateElement("Activity_Group_Id").
			SetText(agid)

		rundata.ActivityIds = append(rundata.ActivityIds, IdPair{
			Source: string(a.GetRef()), Backend: aid})
	}
}
