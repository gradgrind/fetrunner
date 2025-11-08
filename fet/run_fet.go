package fet

import (
	"bufio"
	"context"
	"encoding/xml"
	"fetrunner/autotimetable"
	"fetrunner/base"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var TEMPORARY_FOLDER string

type FetBackend struct {
	basic_data *autotimetable.BasicData
}

func SetFetBackend(
	basic_data *autotimetable.BasicData,
) autotimetable.BackendInterface {
	if len(TEMPORARY_FOLDER) != 0 {
		os.RemoveAll(filepath.Join(TEMPORARY_FOLDER,
			filepath.Base(basic_data.WorkingDir)))
	}
	return &FetBackend{basic_data}
}

func (fbe *FetBackend) Tidy() {
	if len(TEMPORARY_FOLDER) == 0 {
		os.RemoveAll(filepath.Join(fbe.basic_data.WorkingDir, "tmp"))
	} else {
		os.RemoveAll(filepath.Join(TEMPORARY_FOLDER,
			filepath.Base(fbe.basic_data.WorkingDir)))
	}
}

func (fbe *FetBackend) RunBackend(
	instance *autotimetable.TtInstance,
) autotimetable.TtBackend {
	basic_data := fbe.basic_data
	fname := instance.Tag
	var odir string
	if len(TEMPORARY_FOLDER) == 0 {
		odir = filepath.Join(basic_data.WorkingDir, "tmp", fname)
	} else {
		odir = filepath.Join(TEMPORARY_FOLDER,
			filepath.Base(fbe.basic_data.WorkingDir),
			fname)
	}
	err := os.MkdirAll(odir, 0755)
	if err != nil {
		panic(err)
	}
	stemfile := filepath.Join(odir, fname)
	fetfile := stemfile + ".fet"

	// Construct the FET-file
	var fet_xml []byte
	basic_data.Source.PrepareRun(instance.ConstraintEnabled, &fet_xml)
	// Write FET file
	err = os.WriteFile(fetfile, fet_xml, 0644)
	if err != nil {
		panic("Couldn't write fet file to: " + fetfile)
	}
	if instance.Tag == "COMPLETE" {
		// Save fet file at top level of working directory.
		cfile := filepath.Join(basic_data.WorkingDir,
			filepath.Base(strings.TrimSuffix(
				basic_data.WorkingDir, "_fet")+".fet"))
		err = os.WriteFile(cfile, fet_xml, 0644)
		if err != nil {
			panic("Couldn't write fet file to: " + cfile)
		}
	}
	logfile := filepath.Join(odir, "logs", "max_placed_activities.txt")

	ctx, cancel := context.WithCancel(context.Background())
	// Note that it should be safe to call `cancel` multiple times.
	fet_data := &FetTtData{
		finished: false,
		ifile:    fetfile,
		fetxml:   fet_xml,
		odir:     odir,
		logfile:  logfile,
		cancel:   cancel,
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

	if basic_data.Parameters.TESTING {
		params = append(params,
			"--randomseeds10=10",
			"--randomseeds11=11",
			"--randomseeds12=12",
			"--randomseeds20=20",
			"--randomseeds21=21",
			"--randomseeds22=22")
	}

	runCmd := exec.CommandContext(ctx,
		//runCmd := exec.Command(
		"fet-cl", params...,
	)

	go run(fet_data, runCmd)
	return fet_data
}

type FetTtData struct {
	ifile       string
	fetxml      []byte
	odir        string // the working directory for this instance
	logfile     string
	rdfile      *os.File // this must be closed when the subprocess finishes
	reader      *bufio.Reader
	cancel      func()
	fet_timeout bool //
	finished    bool
	count       int
}

func (data *FetTtData) Abort() {
	data.cancel()
}

func (data *FetTtData) Clear() {
	//base.Message.Printf("### Remove %s\n", fttd.workingdir)
	os.RemoveAll(data.odir)
	//} else {
	//	base.Message.Printf("### No TtData: %s\n", instance.Tag)
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
	cmd.CombinedOutput()
	fet_data.finished = true
}

var pattern = "time (.*), FET reached ([0-9]+)"
var re *regexp.Regexp = regexp.MustCompile(pattern)

// `Tick` runs in the "tick" loop. Rather like a "tail" function it reads
// the FET progress from its log file, by simply polling for new lines.
func (data *FetTtData) Tick(
	basic_data *autotimetable.BasicData,
	instance *autotimetable.TtInstance,
) {
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
		var l [][]byte
		for {
			line, err := data.reader.ReadString('\n')
			if err == nil {
				l = re.FindSubmatch([]byte(line))
				continue
			}
			if err == io.EOF {
				if l != nil {
					count, err := strconv.Atoi(string(l[2]))
					if err == nil {
						if count > data.count {
							instance.LastTime = instance.Ticks
							percent := (count * 100) / int(basic_data.NActivities)
							if percent > instance.Progress {
								instance.Progress = percent
								base.Report(fmt.Sprintf("%s: %d @ %d\n",
									instance.Tag, percent, instance.Ticks))
							}
						}
					}
				}

				//TODO: Experiment to catch FET getting stuck soon after start.
				// It may need tweaking.
				if !data.fet_timeout && instance.LastTime < 2 &&
					instance.Ticks-instance.LastTime > 10 {
					data.fet_timeout = true
					data.Abort()
				}

				break
			}
			panic(err)
		}
	}
exit:
	if data.finished {
		if data.rdfile != nil {
			data.rdfile.Close()
		}
		if instance.Progress == 100 {
			instance.RunState = 1
		} else {
			instance.RunState = 2
		}

		efile, err := os.ReadFile(filepath.Join(data.odir, "logs", "errors.txt"))
		if err == nil {
			instance.Message = string(efile)
		} else if data.fet_timeout {
			instance.Message = fmt.Sprintf("FET stuck at beginning (%d%%)",
				instance.Progress)
		} else {
			return
		}
		if len(instance.Constraints) == 1 {
			basic_data.BlockConstraint[instance.Constraints[0]] = true
		}
	}
}

// If there is a result from the main process, there may be a
// corresponding result from the source.
func (data *FetTtData) FinalizeResult(basic_data *autotimetable.BasicData) {
	// Write FET file at top level of working directory.
	fetfile := filepath.Join(basic_data.WorkingDir, "Result.fet")
	err := os.WriteFile(fetfile, data.fetxml, 0644)
	if err != nil {
		panic("Couldn't write fet file to: " + fetfile)
	}
}

// Gather the results of the given run.
func (data *FetTtData) Results(
	basic_data *autotimetable.BasicData,
	instance *autotimetable.TtInstance,
) []autotimetable.ActivityPlacement {
	// Get placements
	xmlpath := filepath.Join(data.odir, "timetables", instance.Tag,
		instance.Tag+"_activities.xml")
	// Open the XML file
	xmlFile, err := os.Open(xmlpath)
	if err != nil {
		base.Bug.Print(err)
		return nil
	}
	// Remember to close the file at the end of the function
	defer xmlFile.Close()
	// read the opened XML file as a byte array.
	base.Message.Printf("Reading: %s\n", xmlpath)
	byteValue, _ := io.ReadAll(xmlFile)
	v := fetResultRoot{}
	err = xml.Unmarshal(byteValue, &v)
	if err != nil {
		base.Bug.Printf("XML error in %s:\n %v\n", xmlpath, err)
		return nil
	}
	// Need to prepare the `ActivityPlacment` fields: activities, days,
	// hours and rooms ...
	// ... room conversion
	room2index := map[string]int{}
	for i, r := range basic_data.Source.GetRooms() {
		room2index[r.Id] = i

	}
	// ... day conversion
	day2index := map[string]int{}
	for i, d := range basic_data.Source.GetDayTags() {
		day2index[d.Id] = i
	}
	// ... hour conversion
	hour2index := map[string]int{}
	for i, h := range basic_data.Source.GetHourTags() {
		hour2index[h.Id] = i
	}
	// ... activity conversion
	activity2index := map[string]int{}
	for i, a := range basic_data.Source.GetActivityRefs() {
		activity2index[a.Id] = i
	}

	// Gather the activities
	activities := make([]autotimetable.ActivityPlacement, len(v.Activities))
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
		activities[i] = autotimetable.ActivityPlacement{
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
