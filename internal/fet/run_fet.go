package fet

import (
	"bufio"
	"context"
	"encoding/xml"
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"

	"github.com/beevik/etree"
)

var (
	FETPATH string
	FET_CL  string = "fet-cl" // "default" value for `FETPATH`, command-line version
	FET_CLW string = "fet-cl" // "default" value for `FETPATH`, GUI version
	// FET_CLW and FET_CL are the same except on Windows: see fet/platform_windows.go.
)

func InitBackend(attdata *autotimetable.AutoTtData) {
	bdata := attdata.BaseData
	if base.TEMPORARY_DIR == "" {
		bdata.SetTmpDir()
	}
	tmpdir := filepath.Join(base.TEMPORARY_DIR, bdata.Name)
	os.RemoveAll(tmpdir)
	var fetbuild *fet_build
	switch stype := bdata.Source.SourceType(); stype {
	case "DB":
		fetbuild = BuildFet(attdata)
	case "FET":
		fetbuild = PrepareFet(attdata)
	default:
		panic("Unsupported timetable source type: " + stype)
	}
	fetbuild.tmpdir = tmpdir
}

// Construct a `fet_build` structure as back-end from a `TtSourceFet`.
func PrepareFet(attdata *autotimetable.AutoTtData) *fet_build {
	source := attdata.Source
	fetsource := source.(*TtSourceFet) // source
	cllist := make([][]*etree.Element, len(fetsource.constraintElements))
	real_soft := attdata.Parameters.REAL_SOFT
	for _, sw := range fetsource.softWeights {
		c := fetsource.constraintElements[sw.Index]
		if real_soft {
			c.SelectElement("Weight_Percentage").SetText(sw.Weight)
		} else {
			c.SelectElement("Weight_Percentage").SetText("100")
		}
	}
	for i, c := range fetsource.constraintElements {
		cllist[i] = []*etree.Element{c}
	}
	fetbuild := &fet_build{
		real_soft:           real_soft,
		no_room_constraints: attdata.Parameters.WITHOUT_ROOM_CONSTRAINTS,
		ttsource:            source,

		Doc: fetsource.doc,
		//WeightTable:        MakeFetWeights(),
		ConstraintElements: cllist,

		//fetroot                *etree.Element
		//room_list              *etree.Element // needed for adding virtual rooms
		//activity_tag_list      *etree.Element // in case these are needed
		//time_constraints_list  *etree.Element
		//space_constraints_list *etree.Element

		//DayList      []string
		//HourList     []string
		//ClassList    []string
		//TeacherList  []string
		//SubjectList  []string
		//RoomList     []string
		//ActivityList []string

		//hard_teacher_blocks [][]base.TimeSlot
		//hard_class_blocks   [][]base.TimeSlot

		// Cache for FET virtual rooms, "hash" -> FET-virtual-room tag
		//fet_virtual_rooms  map[string]string
		//fet_virtual_room_n map[string]int // FET-virtual-room tag -> number of room sets

		//tmpdir string // must be set later, before using
	}

	for _, d := range source.GetDays() {
		fetbuild.DayList = append(fetbuild.DayList, d.Tag)
	}
	for _, h := range source.GetHours() {
		fetbuild.HourList = append(fetbuild.HourList, h.Tag)
	}
	for _, r := range source.GetRooms() {
		fetbuild.RoomList = append(fetbuild.RoomList, r.Tag)
	}

	// The Tag field is, in this case, set in the source.
	for _, a := range source.GetActivities() {
		fetbuild.ActivityList = append(fetbuild.ActivityList, a.Tag)
	}

	attdata.Backend = fetbuild
	return fetbuild
}

func (fetbuild *fet_build) Tidy(bdata *base.BaseData) {
	os.RemoveAll(fetbuild.tmpdir)
}

func (fetbuild *fet_build) RunBackend(
	attdata *autotimetable.AutoTtData,
	instance *autotimetable.TtInstance,
) {
	instance.InstanceBackend = nil
	bdata := attdata.BaseData
	fname := fmt.Sprintf("z%05d~%s", instance.Index, instance.ConstraintType)
	var odir string
	odir = filepath.Join(fetbuild.tmpdir, fname)
	err := os.MkdirAll(odir, 0700)
	if err != nil {
		panic("INVALID_TMP_DIR: " + odir)
	}
	stemfile := filepath.Join(odir, fname)
	fetfile := stemfile + ".fet"

	// Construct the FET-file
	enabled := instance.ConstraintEnabled
	for i, clist := range fetbuild.ConstraintElements {
		for _, c := range clist {
			active := c.SelectElement("Active")
			if enabled[i] {
				active.SetText("true")
			} else {
				active.SetText("false")
			}
		}
	}
	fetbuild.Doc.Indent(2)
	fet_xml, err := fetbuild.Doc.WriteToBytes()
	if err != nil {
		panic(err)
	}

	// Write FET file
	err = os.WriteFile(fetfile, fet_xml, 0600)
	if err != nil {
		panic("WRITE_TMP_FET_FILE_FAILED: " + fetfile)
	}
	if instance.ConstraintType == "_COMPLETE" &&
		attdata.Parameters.WRITE_FET_FILE {
		// Save "complete" fet file with "_" prefix in working directory.
		cfile := filepath.Join(bdata.SourceDir, "_"+bdata.Name+".fet")
		err = os.WriteFile(cfile, fet_xml, 0600)
		if err != nil {
			panic("WRITE_FET_FILE_FAILED: " + cfile)
		}
	}
	logfile := filepath.Join(odir, "logs", "max_placed_activities.txt")
	resultfile := filepath.Join(odir, "timetables", fname, fname+"_activities.xml")

	ctx, cancel := context.WithCancel(context.Background())
	// Note that it should be safe to call `cancel` multiple times.
	fet_data := &FetTtData{
		finished:   0,
		ifile:      fetfile,
		fetxml:     fet_xml,
		odir:       odir,
		logfile:    logfile,
		resultfile: resultfile,
		cancel:     cancel,
		count:      0,
	}

	params := []string{
		"--inputfile=" + fetfile,
		"--writetimetablesstatistics=false",
		"--writetimetablesdayshorizontal=false",
		"--writetimetablesdaysvertical=false",
		"--writetimetablestimehorizontal=false",
		"--writetimetablestimevertical=false",
		"--writetimetablessubgroups=false",
		"--writetimetablesgroups=false",
		"--writetimetablesyears=false",
		"--writetimetablesteachers=false",
		"--writetimetablesteachersfreeperiods=false",
		"--writetimetablesbuildings=false",
		"--writetimetablesrooms=false",
		"--writetimetablessubjects=false",
		"--outputdir=" + odir,
	}

	if attdata.Parameters.TESTING {
		params = append(params,
			"--randomseeds10=10",
			"--randomseeds11=11",
			"--randomseeds12=12",
			"--randomseeds20=20",
			"--randomseeds21=21",
			"--randomseeds22=22")
	}

	runCmd := exec.CommandContext(ctx, FETPATH, params...)

	bdata.Logger.Result(".START", fmt.Sprintf("%d.%s.%d.%d",
		instance.Index,
		fetbuild.ConstraintName(instance),
		len(instance.Constraints),
		instance.Timeout))

	go run(fet_data, runCmd)
	instance.InstanceBackend = fet_data
}

type FetTtData struct {
	ifile      string
	fetxml     []byte
	odir       string // the working directory for this instance
	logfile    string
	resultfile string
	rdfile     *os.File // this must be closed when the subprocess finishes
	reader     *bufio.Reader
	cancel     func()
	//fet_timeout bool
	finished int
	count    int
	errormsg string // record error message
}

func (data *FetTtData) Abort() {
	data.cancel()
}

func (data *FetTtData) Clear() {
	// Remove instance directory
	os.RemoveAll(data.odir)
}

// The executable, `fet-cl`, places any messages in the `log` directory, as
// `result.txt` (which is probably not so interesting), `warnings.txt` (which
// might possibly containt something of diagnostic interest) and `errors.txt`
// (which may well contain diagnostic information that should ideally
// have been caught earlier ...). The `warnings.txt` and `errors.txt`
// files are present only if there is something to report. If there is an
// `errors.txt`, it should certainly be reported (in `fet_data.message`).

// The completion code of `fet-cl` is not particularly helpful, so the
// success of the run is determined by checking the number of placed
// activities and the existence of an `errors.txt` file.

// `run` is a goroutine. The last item to be changed must be `fet_data.state`,
// to avoid potential race conditions.
func run(fet_data *FetTtData, cmd *exec.Cmd) {
	_, err := cmd.CombinedOutput()
	if err != nil {
		e := err.Error()
		// Some errors arise as a result of a termination signal, or after
		// calling `Abort`, which leads to the context being cancelled.
		// The actual message seems to be system-dependent, so probably not
		// that interesting ... except if there is a bug somewhere?
		fet_data.finished = -1
		fet_data.errormsg = e
	} else {
		fet_data.finished = 1
	}
	fet_data.cancel()
}

// Regexp for reading the progress of a run from the FET log file
var pattern = "time (.*), FET reached ([0-9]+)"
var re *regexp.Regexp = regexp.MustCompile(pattern)

// `DoTick` runs in the "tick" loop. Rather like a "tail" function it reads
// the FET progress from its log file, by simply polling for new lines.
func (data *FetTtData) DoTick(
	attdata *autotimetable.AutoTtData,
	instance *autotimetable.TtInstance,
) {
	instance.Ticks++
	logger := attdata.BaseData.Logger
	progressed := false
	if data.reader == nil {
		// Await the existence of the log file
		file, err := os.Open(data.logfile)
		if err != nil {
			goto exit
		}
		data.rdfile = file // this needs closing
		data.reader = bufio.NewReader(file)
	}
	{
		var l [][]byte = nil
		for {
			line, err := data.reader.ReadString('\n')
			if err != nil {
				if err == io.EOF {
					break
				}
				panic(err)
			}
			l = re.FindSubmatch([]byte(line))
		}
		if l != nil {
			count, err := strconv.Atoi(string(l[2]))
			if err == nil {
				if count > data.count {
					data.count = count
					instance.LastTime = instance.Ticks
					percent := (count * 100) / int(attdata.NActivities)
					if percent > instance.Progress {
						instance.Progress = percent
						logger.Result(".PROGRESS",
							fmt.Sprintf("%d.%d.%d",
								instance.Index,
								instance.Progress,
								instance.Ticks))
						progressed = true
					}
				}
			}
		}
	}
	if !progressed && data.finished == 0 {
		logger.Result(".NOPROGRESS",
			fmt.Sprintf("%d.%d",
				instance.Index,
				instance.Ticks))
	}

exit:
	if data.finished != 0 {
		if data.rdfile != nil {
			data.rdfile.Close()
		}
		instance.Finished = true
		if instance.RunState == autotimetable.INSTANCE_RUNNING {
			if instance.Progress == 100 {
				instance.RunState = autotimetable.INSTANCE_SUCCESSFUL
			} else {
				instance.RunState = autotimetable.INSTANCE_FAILED
			}
		}
		logger.Result(".END", fmt.Sprintf("%d.%d.%d",
			instance.Index, instance.Progress, instance.RunState))
		efile, err := os.ReadFile(filepath.Join(data.odir, "logs", "errors.txt"))
		if err == nil {
			instance.Message = string(efile)
		}
		return
	}

	//TODO: Timeouts ...

	if instance.RunState == autotimetable.INSTANCE_RUNNING {

		/* TODO: testing ...
		   remaining := attdata.Parameters.TIMEOUT - attdata.Ticks
		   prog := instance.Progress
		   ticks := instance.Ticks
		   if prog != 0 {
		       expect := (100 - prog) * ticks / prog
		       if expect > remaining {
		           logger.Info("FET_TooSlow %d:%s @ %d, p: %d%% n: %d",
		               instance.Index,
		               instance.ConstraintType,
		               ticks,
		               prog,
		               len(instance.Constraints))
		           data.Abort()
		           instance.RunState = autotimetable.ABORT_TIMED_OUT
		       }
		   }
		   //... TODO */

		//TODO: Experiment to catch FET getting stuck soon after start.
		// It may need tweaking.
		if instance.LastTime < 2 &&
			instance.Ticks-instance.LastTime > 10 {
			logger.Info("FET_Stuck_0 %d:%s @ %d, p: %d%% n: %d",
				instance.Index,
				instance.ConstraintType,
				instance.Ticks,
				instance.Progress,
				len(instance.Constraints))
			data.Abort()
			instance.RunState = autotimetable.ABORT_TIMED_OUT
		}

		t := instance.Timeout
		if t == 0 {
			// Check for lack of progress for instances with no timeout
			if instance.LastTime < attdata.Parameters.LAST_TIME_0 &&
				instance.Ticks >= attdata.Parameters.LAST_TIME_1 {
				// Stop instance
				logger.Info(
					"FET_Slow_0 %d:%s @ %d, p: %d n: %d",
					instance.Index,
					instance.ConstraintType,
					instance.Ticks,
					instance.Progress,
					len(instance.Constraints))
				data.Abort()
				instance.RunState = autotimetable.ABORT_TIMED_OUT
			}
		} else {
			limit := (instance.Ticks * 20) / t // the factor was originally 50
			//TODO: This is not really a timeout! And the multiplier is highly experimental.
			// It's more of a "progress on course" criterion.
			if instance.Progress < limit {
				// Progress is too slow ...
				logger.Info("FET_Slow_1 %d:%s @ %d, p: %d n: %d",
					instance.Index,
					instance.ConstraintType,
					instance.Ticks,
					instance.Progress,
					len(instance.Constraints))
				data.Abort()
				instance.RunState = autotimetable.ABORT_TIMED_OUT
			}
		}
	}
}

// If there is a result from the main process, there may be a
// corresponding result from the source.
func (data *FetTtData) FinalizeResult(
	bdata *base.BaseData,
	attdata *autotimetable.AutoTtData) {
	// Write FET file at top level of working directory.
	fetfile := filepath.Join(bdata.SourceDir, bdata.Name+"_Result.fet")
	err := os.WriteFile(fetfile, data.fetxml, 0644)
	if err != nil {
		bdata.Logger.Error("%s", err)
	}
}

// Gather the results of the given run.
func (fetbuild *fet_build) Results(
	logger *base.Logger,
	instance *autotimetable.TtInstance,
) []autotimetable.TtActivityPlacement {
	// Get placements
	xmlpath := instance.InstanceBackend.(*FetTtData).resultfile
	// Open the XML file
	xmlFile, err := os.Open(xmlpath)
	if err != nil {
		logger.Bug("%v", err)
		return nil
	}
	// Remember to close the file at the end of the function
	defer xmlFile.Close()
	// read the opened XML file as a byte array.
	logger.Info("Reading: %s", xmlpath)
	byteValue, _ := io.ReadAll(xmlFile)
	v := fetResultRoot{}
	err = xml.Unmarshal(byteValue, &v)
	if err != nil {
		logger.Bug("XML error in %s:\n %v\n", xmlpath, err)
		return nil
	}
	// Need to prepare the `ActivityPlacement` fields: activities, days,
	// hours and rooms ...
	// ... room conversion
	room2index := map[string]int{}
	for i, r := range fetbuild.RoomList {
		room2index[r] = i
	}
	// ... day conversion
	day2index := map[string]int{}
	for i, d := range fetbuild.DayList {
		day2index[d] = i
	}
	// ... hour conversion
	hour2index := map[string]int{}
	for i, h := range fetbuild.HourList {
		hour2index[h] = i
	}
	// ... activity conversion
	activity2index := map[string]int{}
	for i, a := range fetbuild.ActivityList {
		activity2index[a] = i
	}

	// Gather the activities
	activities := make([]autotimetable.TtActivityPlacement, len(v.Activities))
	for i, a := range v.Activities {
		rooms := []int{}
		if len(a.Real_Room) != 0 {
			for _, r := range a.Real_Room {
				ix, ok := room2index[r]
				if !ok {
					panic("Unknown room: " + r)
				}
				rooms = append(rooms, ix)
			}
		} else if len(a.Room) != 0 {
			ix, ok := room2index[a.Room]
			if !ok {
				panic("Unknown room: " + a.Room)
			}
			rooms = append(rooms, ix)
		}
		activities[i] = autotimetable.TtActivityPlacement{
			Activity: activity2index[a.Id],
			Day:      day2index[a.Day],
			Hour:     hour2index[a.Hour],
			Rooms:    rooms,
		}
	}
	return activities
}

type fetResultRoot struct { // The root node.
	XMLName    xml.Name            `xml:"Activities_Timetable"`
	Activities []fetResultActivity `xml:"Activity"`
}

type fetResultActivity struct {
	XMLName   xml.Name `xml:"Activity"`
	Id        string
	Day       string
	Hour      string
	Room      string
	Real_Room []string
}
