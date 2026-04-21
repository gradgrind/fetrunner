package w365tt

import (
    "encoding/json"
    "fetrunner/internal/base"
    "io"
    "os"
    "slices"
    "strconv"
    "strings"
)

// Read to the local, tweaked DbTopLevel
func ReadJSON(jsonpath string) *W365TopLevel {
    // Open the  JSON file
    jsonFile, err := os.Open(jsonpath)
    if err != nil {
        base.LogError("--W365_OPEN_JSON %v", err)
        return nil
    }
    // Remember to close the file at the end of the function
    defer jsonFile.Close()
    // read the opened XML file as a byte array.
    byteValue, _ := io.ReadAll(jsonFile)
    v := W365TopLevel{}
    err = json.Unmarshal(byteValue, &v)
    if err != nil {
        base.LogError("--W365_JSON_UNMARSHALL %s", err)
        return nil
    }
    return &v
}

func LoadJSON(jsonpath string) bool {
    dbi := ReadJSON(jsonpath)
    if dbi == nil {
        return false
    }
    ndb := base.DataBase.Db
    ndb.Institution = dbi.Info.Institution
    ndb.FirstAfternoonHour = dbi.Info.FirstAfternoonHour
    ndb.Reference = dbi.Info.Reference
    { // Add lunch-break times
        mb := dbi.Info.MiddayBreak
        if len(mb) != 0 {
            if len(mb) > 1 {
                // Sort and check contiguity.
                slices.Sort(mb)
            }
            mb0 := mb[0]
            mb1 := mb[len(mb)-1]
            if mb1-mb0 >= len(mb) {
                base.LogError("--W365_MiddayBreak_NOT_CONTIGUOUS")
            } else {
                ndb.MiddayBreak0 = mb0
                ndb.MiddayBreak1 = mb1
            }
        }
    }
    ndb.ModuleData = map[string]any{
        "FetData": dbi.FetData,
    }
    dbi.readDays()
    dbi.readHours()
    dbi.readTeachers()
    dbi.readSubjects()
    dbi.readRooms()
    dbi.readRoomGroups()
    // To manage potentially incomplete Tag and Name fields for RoomGroups
    // from W365, perform the checking after all room types have been "read".
    dbi.checkRoomGroups()
    dbi.readClasses()
    dbi.readCourses()
    dbi.readSuperCourses()
    dbi.readLessons()
    dbi.readConstraints()
    return true
}

func (dbi *W365TopLevel) readDays() {
    for _, e := range dbi.Days {
        n := base.NewDay(e.Id)
        n.Tag = e.Tag
        n.Name = e.Name
    }
}

func (dbi *W365TopLevel) readHours() {
    for i, e := range dbi.Hours {
        tag := e.Tag
        if tag == "" {
            tag = "(" + strconv.Itoa(i+1) + ")"
        }
        n := base.NewHour(e.Id)
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

func (dbi *W365TopLevel) readTeachers() {
    dbi.TeacherMap = map[NodeRef]struct{}{}
    tagmap := map[string]struct{}{} // to test for duplicate tags
    for _, e := range dbi.Teachers {
        // Perform some checks and add to the tag map.
        _, nok := tagmap[e.Tag]
        if nok {
            base.LogError(
                "--W365_TEACHER_TAG_DEFINED_TWICE %s",
                e.Tag)
            continue
        }
        tagmap[e.Tag] = struct{}{}

        // Make new teacher node
        n := base.NewTeacher(e.Id)
        n.Tag = e.Tag
        n.Name = e.Name
        n.Firstname = e.Firstname

        dbi.TeacherMap[e.Id] = struct{}{} // flag the Id as valid teacher

        // +++ Add constraints ...
        ndb := base.DataBase.Db

        // MaxAfternoons = 0 has a special meaning (all blocked), so the
        // corresponding constraint is not needed, see `handleZeroAfternoons`.
        amax := e.MaxAfternoons
        if amax > 0 {
            ndb.NewTeacherMaxAfternoons(
                "", base.MAXWEIGHT, n.Id, amax)
        }
        // Not available times – add all afternoons if amax == 0
        tsl := dbi.handleZeroAfternoons(e.NotAvailable, amax)
        if len(tsl) != 0 {
            // Add a constraint
            ndb.NewTeacherNotAvailable("", base.MAXWEIGHT, n.Id, tsl)
        }

        // MinActivitiesPerDay
        if e.MinLessonsPerDay > 0 {
            ndb.NewTeacherMinHoursPerDay(
                "", base.MAXWEIGHT, n.Id, e.MinLessonsPerDay)
        }
        // MaxActivitiesPerDay
        if e.MaxLessonsPerDay > 0 {
            ndb.NewTeacherMaxHoursPerDay(
                "", base.MAXWEIGHT, n.Id, e.MaxLessonsPerDay)
        }
        // MaxDays
        if e.MaxDays > 0 {
            ndb.NewTeacherMaxDays(
                "", base.MAXWEIGHT, n.Id, e.MaxDays)
        }
        // MaxGapsPerDay
        if e.MaxGapsPerDay >= 0 {
            ndb.NewTeacherMaxGapsPerDay(
                "", base.MAXWEIGHT, n.Id, e.MaxGapsPerDay)
        }
        // MaxGapsPerWeek
        if e.MaxGapsPerWeek >= 0 {
            ndb.NewTeacherMaxGapsPerWeek(
                "", base.MAXWEIGHT, n.Id, e.MaxGapsPerWeek)
        }
        // LunchBreak
        if e.LunchBreak {
            ndb.NewTeacherLunchBreak(
                "", base.MAXWEIGHT, n.Id)
        }
    }
}
