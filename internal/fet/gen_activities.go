package fet

import (
    "slices"
    "strconv"
)

// Generate the fet activities.
func (fetbuild *fet_build) set_activities() {
    alist := fetbuild.ttsource.GetActivities()
    fetactivities := fetbuild.fetroot.CreateElement("Activities_List")
    fetbuild.ActivityList = make([]string, len(alist))
    for ai, tt_activity := range alist {
        fetactivity := fetactivities.CreateElement("Activity")
        //fetbuild.ActivityElementList = append(fetbuild.ActivityElementList, fetactivity)
        // The fet activities start at Id = 1
        aid := strconv.Itoa(ai + 1)
        fetbuild.ActivityList[ai] = aid
        fetactivity.CreateElement("Id").SetText(aid)
        // Teachers
        tlist := []string{}
        for _, ti := range tt_activity.Teachers {
            tlist = append(tlist, fetbuild.TeacherList[ti])
        }
        slices.Sort(tlist)
        for _, t := range tlist {
            fetactivity.CreateElement("Teacher").SetText(t)
        }
        // Subject
        fetactivity.CreateElement("Subject").SetText(tt_activity.Subject)
        // Student groups
        glist := []string{}
        for _, cg := range tt_activity.Groups {
            glist = append(glist, fetGroupTag(cg))
        }
        slices.Sort(glist)
        for _, g := range glist {
            fetactivity.CreateElement("Students").SetText(g)
        }
        //TODO? Activity_Tag: tag (can be more than one of these)
        fetactivity.CreateElement("Active").SetText("true")
        // Activities are not grouped in the FET generated here, so the
        // total duration is the same as that of the individual activity
        // and the activity group id is "0".
        duration := tt_activity.Duration
        fetactivity.CreateElement("Total_Duration").
            SetText(strconv.Itoa(duration))
        fetactivity.CreateElement("Duration").
            SetText(strconv.Itoa(duration))
        fetactivity.CreateElement("Activity_Group_Id").
            SetText("0")
    }
}
