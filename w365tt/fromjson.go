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
func ReadJSON(logger base.BasicLogger, jsonpath string) *W365TopLevel {
	// Open the  JSON file
	jsonFile, err := os.Open(jsonpath)
	if err != nil {
		logger.Error("%v", err)
		return nil
	}
	// Remember to close the file at the end of the function
	defer jsonFile.Close()
	// read the opened XML file as a byte array.
	byteValue, _ := io.ReadAll(jsonFile)
	logger.Info("*+ Reading: %s\n", jsonpath)
	v := W365TopLevel{}
	err = json.Unmarshal(byteValue, &v)
	if err != nil {
		logger.Error("Could not unmarshal json: %s\n", err)
		return nil
	}
	return &v
}

func LoadJSON(newdb *db.DbTopLevel, jsonpath string) bool {
	dbi := ReadJSON(newdb.Logger, jsonpath)
	if dbi == nil {
		return false
	}
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
	return true
}

func (dbi *W365TopLevel) readDays(newdb *db.DbTopLevel) {
	for _, e := range dbi.Days {
		n := newdb.NewDay(e.Id)
		n.Tag = e.Tag
		n.Name = e.Name
	}
}

func (dbi *W365TopLevel) readHours(newdb *db.DbTopLevel) {
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

func (dbi *W365TopLevel) readTeachers(newdb *db.DbTopLevel) {
	dbi.TeacherMap = map[NodeRef]struct{}{}
	tagmap := map[string]struct{}{} // to test for duplicate tags
	for _, e := range dbi.Teachers {
		// Perform some checks and add to the tag map.
		_, nok := tagmap[e.Tag]
		if nok {
			newdb.Logger.Error(
				"Teacher Tag (Shortcut) defined twice: %s\n",
				e.Tag)
			continue
		}
		tagmap[e.Tag] = struct{}{}

		// Make new teacher node
		n := newdb.NewTeacher(e.Id)
		n.Tag = e.Tag
		n.Name = e.Name
		n.Firstname = e.Firstname

		dbi.TeacherMap[e.Id] = struct{}{} // flag the Id as valid teacher

		// +++ Add constraints ...

		// MaxAfternoons = 0 has a special meaning (all blocked), so the
		// corresponding constraint is not needed, see `handleZeroAfternoons`.
		amax := e.MaxAfternoons
		if amax > 0 {
			newdb.NewTeacherMaxAfternoons(
				"", db.MAXWEIGHT, n.Id, amax)
		}
		// Not available times – add all afternoons if amax == 0
		tsl := dbi.handleZeroAfternoons(e.NotAvailable, amax)
		if len(tsl) != 0 {
			// Add a constraint
			newdb.NewTeacherNotAvailable("", db.MAXWEIGHT, n.Id, tsl)
		}

		// MinActivitiesPerDay
		if e.MinLessonsPerDay > 0 {
			newdb.NewTeacherMinActivitiesPerDay(
				"", db.MAXWEIGHT, n.Id, e.MinLessonsPerDay)
		}
		// MaxActivitiesPerDay
		if e.MaxLessonsPerDay > 0 {
			newdb.NewTeacherMaxActivitiesPerDay(
				"", db.MAXWEIGHT, n.Id, e.MaxLessonsPerDay)
		}
		// MaxDays
		if e.MaxDays > 0 {
			newdb.NewTeacherMaxDays(
				"", db.MAXWEIGHT, n.Id, e.MaxDays)
		}
		// MaxGapsPerDay
		if e.MaxGapsPerDay >= 0 {
			newdb.NewTeacherMaxGapsPerDay(
				"", db.MAXWEIGHT, n.Id, e.MaxGapsPerDay)
		}
		// MaxGapsPerWeek
		if e.MaxGapsPerWeek >= 0 {
			newdb.NewTeacherMaxGapsPerWeek(
				"", db.MAXWEIGHT, n.Id, e.MaxGapsPerWeek)
		}
		// LunchBreak
		if e.LunchBreak {
			newdb.NewTeacherLunchBreak(
				"", db.MAXWEIGHT, n.Id)
		}
	}
}
