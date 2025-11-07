package makefet

import (
	"encoding/xml"
	"fetrunner/timetable"
	"slices"
)

type fetActivity struct {
	XMLName           xml.Name `xml:"Activity"`
	Id                int
	Teacher           []string `xml:",omitempty"`
	Subject           string
	Activity_Tag      string   `xml:",omitempty"`
	Students          []string `xml:",omitempty"`
	Active            bool
	Total_Duration    int
	Duration          int
	Activity_Group_Id int
	Comments          string
}

type fetActivitiesList struct {
	XMLName  xml.Name `xml:"Activities_List"`
	Activity []fetActivity
}

type fetActivityTag struct {
	XMLName   xml.Name `xml:"Activity_Tag"`
	Name      string
	Printable bool
}

type fetActivityTags struct {
	XMLName      xml.Name `xml:"Activity_Tags_List"`
	Activity_Tag []fetActivityTag
}

// Generate the fet activities.
func getActivities(fetinfo *fetInfo) {
	tt_data := fetinfo.tt_data
	db0 := tt_data.Db

	// ************* Start with the activity tags
	tags := []fetActivityTag{}
	/* TODO ???
	s2tag := map[string]string{}
	for _, ts := range tagged_subjects {
		tag := fmt.Sprintf("Tag_%s", ts)
		s2tag[ts] = tag
		tags = append(tags, fetActivityTag{
			Name: tag,
		})
	}
	*/
	fetinfo.fetdata.Activity_Tags_List = fetActivityTags{
		Activity_Tag: tags,
	}

	// ************* Now the activities
	activities := []fetActivity{}

	for ai, tt_activity := range tt_data.Activities {
		cinfo := tt_data.CourseInfoList[tt_activity.CourseInfo]
		// Teachers
		tlist := []string{}
		for _, ti := range cinfo.Teachers {
			tlist = append(tlist, tt_data.Db.Teachers[ti].GetTag())
		}
		slices.Sort(tlist)
		// Groups
		glist := []string{}
		for _, cg := range cinfo.Groups {
			glist = append(glist, fetGroupTag(cg))
		}
		slices.Sort(glist)
		/* ???
		atag := ""
		if slices.Contains(tagged_subjects, sbj) {
			atag = fmt.Sprintf("Tag_%s", sbj)
		}
		*/

		// Get the total duration for this course.
		totalDuration := 0
		for _, aix := range cinfo.Activities {
			totalDuration += db0.Activities[aix].Duration
		}
		// Start FET activity indexes at 1
		agid := 0
		if len(cinfo.Activities) > 1 {
			agid = activityIndex2fet(tt_data, cinfo.Activities[0])
		}
		a := db0.Activities[ai]
		activities = append(activities,
			fetActivity{
				Id: activityIndex2fet(
					tt_data, timetable.ActivityIndex(ai)),
				Teacher:  tlist,
				Subject:  cinfo.Subject,
				Students: glist,
				//Activity_Tag:      atag,
				Active:            true,
				Total_Duration:    totalDuration,
				Duration:          a.Duration,
				Activity_Group_Id: agid,
				Comments:          string(a.GetRef()),
			},
		)
	}
	fetinfo.fetdata.Activities_List = fetActivitiesList{
		Activity: activities,
	}
}
