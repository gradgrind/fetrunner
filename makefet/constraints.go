package makefet

import (
	"encoding/xml"
)

type notAvailableTime struct {
	XMLName xml.Name `xml:"Not_Available_Time"`
	Day     string
	Hour    string
}

type teacherNotAvailable struct {
	XMLName                       xml.Name `xml:"ConstraintTeacherNotAvailableTimes"`
	Weight_Percentage             int
	Teacher                       string
	Number_of_Not_Available_Times int
	Not_Available_Time            []notAvailableTime
	Active                        bool
	Comments                      string
}

type studentsNotAvailable struct {
	XMLName                       xml.Name `xml:"ConstraintStudentsSetNotAvailableTimes"`
	Weight_Percentage             int
	Students                      string
	Number_of_Not_Available_Times int
	Not_Available_Time            []notAvailableTime
	Active                        bool
	Comments                      string
}

type startingTime struct {
	XMLName            xml.Name `xml:"ConstraintActivityPreferredStartingTime"`
	Weight_Percentage  int
	Activity_Id        int
	Preferred_Day      string
	Preferred_Hour     string
	Permanently_Locked bool
	Active             bool
	Comments           string
}

type minDaysBetweenActivities struct {
	XMLName                 xml.Name `xml:"ConstraintMinDaysBetweenActivities"`
	Weight_Percentage       string
	Consecutive_If_Same_Day bool
	Number_of_Activities    int
	Activity_Id             []int
	MinDays                 int
	Active                  bool
	Comments                string
}

// *** Teacher constraints
type lunchBreakT struct {
	XMLName             xml.Name `xml:"ConstraintTeacherMaxHoursDailyInInterval"`
	Weight_Percentage   int
	Teacher             string
	Interval_Start_Hour string
	Interval_End_Hour   string
	Maximum_Hours_Daily int
	Active              bool
	Comments            string
}

type maxGapsPerDayT struct {
	XMLName           xml.Name `xml:"ConstraintTeacherMaxGapsPerDay"`
	Weight_Percentage int
	Teacher           string
	Max_Gaps          int
	Active            bool
	Comments          string
}

type maxGapsPerWeekT struct {
	XMLName           xml.Name `xml:"ConstraintTeacherMaxGapsPerWeek"`
	Weight_Percentage int
	Teacher           string
	Max_Gaps          int
	Active            bool
	Comments          string
}

type minLessonsPerDayT struct {
	XMLName             xml.Name `xml:"ConstraintTeacherMinHoursDaily"`
	Weight_Percentage   int
	Teacher             string
	Minimum_Hours_Daily int
	Allow_Empty_Days    bool
	Active              bool
	Comments            string
}

type maxLessonsPerDayT struct {
	XMLName             xml.Name `xml:"ConstraintTeacherMaxHoursDaily"`
	Weight_Percentage   int
	Teacher             string
	Maximum_Hours_Daily int
	Active              bool
	Comments            string
}

type maxDaysT struct {
	XMLName           xml.Name `xml:"ConstraintTeacherMaxDaysPerWeek"`
	Weight_Percentage int
	Teacher           string
	Max_Days_Per_Week int
	Active            bool
	Comments          string
}

// for MaxAfternoons
type maxDaysinIntervalPerWeekT struct {
	XMLName             xml.Name `xml:"ConstraintTeacherIntervalMaxDaysPerWeek"`
	Weight_Percentage   int
	Teacher             string
	Interval_Start_Hour string
	Interval_End_Hour   string
	// Interval_End_Hour void ("") means the end of the day (which has no name)
	Max_Days_Per_Week int
	Active            bool
	Comments          string
}

// *** Class constraints

type lunchBreak struct {
	XMLName             xml.Name `xml:"ConstraintStudentsSetMaxHoursDailyInInterval"`
	Weight_Percentage   int
	Students            string
	Interval_Start_Hour string
	Interval_End_Hour   string
	Maximum_Hours_Daily int
	Active              bool
	Comments            string
}

type maxGapsPerDay struct {
	XMLName           xml.Name `xml:"ConstraintStudentsSetMaxGapsPerDay"`
	Weight_Percentage int
	Max_Gaps          int
	Students          string
	Active            bool
	Comments          string
}

type maxGapsPerWeek struct {
	XMLName           xml.Name `xml:"ConstraintStudentsSetMaxGapsPerWeek"`
	Weight_Percentage int
	Max_Gaps          int
	Students          string
	Active            bool
	Comments          string
}

type minLessonsPerDay struct {
	XMLName             xml.Name `xml:"ConstraintStudentsSetMinHoursDaily"`
	Weight_Percentage   int
	Minimum_Hours_Daily int
	Students            string
	Allow_Empty_Days    bool
	Active              bool
	Comments            string
}

type maxLessonsPerDay struct {
	XMLName             xml.Name `xml:"ConstraintStudentsSetMaxHoursDaily"`
	Weight_Percentage   int
	Maximum_Hours_Daily int
	Students            string
	Active              bool
	Comments            string
}

// for MaxAfternoons
type maxDaysinIntervalPerWeek struct {
	XMLName             xml.Name `xml:"ConstraintStudentsSetIntervalMaxDaysPerWeek"`
	Weight_Percentage   int
	Students            string
	Interval_Start_Hour string
	Interval_End_Hour   string
	// Interval_End_Hour void ("") means the end of the day (which has no name)
	Max_Days_Per_Week int
	Active            bool
	Comments          string
}

// for ForceFirstHour
type maxLateStarts struct {
	XMLName                       xml.Name `xml:"ConstraintStudentsSetEarlyMaxBeginningsAtSecondHour"`
	Weight_Percentage             int
	Max_Beginnings_At_Second_Hour int
	Students                      string
	Active                        bool
	Comments                      string
}
