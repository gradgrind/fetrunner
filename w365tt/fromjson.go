package w365tt

import (
	"encoding/json"
	"fetrunner/base"
	"fetrunner/db"
	"io"
	"os"
	"strconv"
	"strings"
)

// Read to the local, tweaked DbTopLevel
func ReadJSON(jsonpath string) *DbTopLevel {
	// Open the  JSON file
	jsonFile, err := os.Open(jsonpath)
	if err != nil {
		base.Error.Fatal(err)
	}
	// Remember to close the file at the end of the function
	defer jsonFile.Close()
	// read the opened XML file as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)
	base.Message.Printf("*+ Reading: %s\n", jsonpath)
	v := DbTopLevel{}
	err = json.Unmarshal(byteValue, &v)
	if err != nil {
		base.Error.Fatalf("Could not unmarshal json: %s\n", err)
	}
	return &v
}

func LoadJSON(newdb *db.DbTopLevel, jsonpath string) {
	dbi := ReadJSON(jsonpath)
	newdb.Info = db.Info(dbi.Info)
	newdb.ModuleData = map[string]any{
		"FetData": dbi.FetData,
	}
	dbi.readDays(newdb)
	dbi.readHours(newdb)
	dbi.readTeachers(newdb)
	dbi.readSubjects(newdb)
	dbi.readRooms(newdb)
	dbi.readRoomGroups(newdb)
	// To manage potentially incomplete Tag and Name fields for RoomGroups
	// from W365, perform the checking after all room types have been "read".
	dbi.checkRoomGroups(newdb)
	dbi.readClasses(newdb)
	dbi.readCourses(newdb)
	dbi.readSuperCourses(newdb)
	dbi.readLessons(newdb)
	dbi.readConstraints(newdb)
}

func (dbi *DbTopLevel) readDays(newdb *db.DbTopLevel) {
	for _, e := range dbi.Days {
		n := newdb.NewDay(e.Id)
		n.Tag = e.Tag
		n.Name = e.Name
	}
}

func (dbi *DbTopLevel) readHours(newdb *db.DbTopLevel) {
	for i, e := range dbi.Hours {
		tag := e.Tag
		if tag == "" {
			tag = "(" + strconv.Itoa(i+1) + ")"
		}
		n := newdb.NewHour(e.Id)
		n.Tag = tag
		n.Name = e.Name
		// If the input times have seconds, strip these off.
		ts := strings.Split(e.Start, ":")
		if len(ts) == 3 {
			n.Start = ts[0] + ":" + ts[1]
		} else {
			n.Start = e.Start
		}
		ts = strings.Split(e.End, ":")
		if len(ts) == 3 {
			n.End = ts[0] + ":" + ts[1]
		} else {
			n.End = e.End
		}
	}
}

func (dbi *DbTopLevel) readTeachers(newdb *db.DbTopLevel) {
	dbi.TeacherMap = map[NodeRef]bool{}
	for _, e := range dbi.Teachers {
		// MaxAfternoons = 0 has a special meaning (all blocked)
		amax := e.MaxAfternoons
		tsl := dbi.handleZeroAfternoons(e.NotAvailable, amax)
		if amax == 0 {
			amax = -1
		}
		n := newdb.NewTeacher(e.Id)
		n.Tag = e.Tag
		n.Name = e.Name
		n.Firstname = e.Firstname

		//TODO: add constraints ...

		n.NotAvailable = tsl
		n.MinActivitiesPerDay = e.MinLessonsPerDay
		n.MaxActivitiesPerDay = e.MaxLessonsPerDay
		n.MaxDays = e.MaxDays
		n.MaxGapsPerDay = e.MaxGapsPerDay
		n.MaxGapsPerWeek = e.MaxGapsPerWeek
		n.MaxAfternoons = amax
		n.LunchBreak = e.LunchBreak

		dbi.TeacherMap[e.Id] = true
	}
}
