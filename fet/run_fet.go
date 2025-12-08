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
	attdata *autotimetable.AutoTtData
}

func SetFetBackend(bdata *base.BaseData, attdata *autotimetable.AutoTtData) {
	if len(TEMPORARY_FOLDER) != 0 {
		os.RemoveAll(filepath.Join(TEMPORARY_FOLDER,
			filepath.Base(bdata.SourceDir)))
	}
	attdata.BackendInterface = &FetBackend{attdata}
}

func (fbe *FetBackend) Tidy(bdata *base.BaseData) {
	if TEMPORARY_FOLDER == "" {
		os.RemoveAll(filepath.Join(bdata.SourceDir, "tmp", bdata.Name))
	} else {
		os.RemoveAll(filepath.Join(TEMPORARY_FOLDER,
			filepath.Base(bdata.Name)))
	}
}

func (fbe *FetBackend) RunBackend(
	bdata *base.BaseData,
	instance *autotimetable.TtInstance,
) autotimetable.TtBackend {
	attdata := fbe.attdata

	fname := fmt.Sprintf("z%05d~%s", instance.Index, instance.ConstraintType)
	var odir string
	if TEMPORARY_FOLDER == "" {
		odir = filepath.Join(bdata.SourceDir, "tmp", bdata.Name, fname)
	} else {
		odir = filepath.Join(TEMPORARY_FOLDER,
			filepath.Base(bdata.Name),
			fname)
	}
	err := os.MkdirAll(odir, 0755)
	if err != nil {
		//TODO?
		panic(err)
	}
	stemfile := filepath.Join(odir, fname)
	fetfile := stemfile + ".fet"

	// Construct the FET-file
	var fet_xml []byte
	attdata.Source.PrepareRun(instance.ConstraintEnabled, &fet_xml)
	// Write FET file
	err = os.WriteFile(fetfile, fet_xml, 0644)
	if err != nil {
		//TODO?
		panic("Couldn't write fet file to: " + fetfile)
	}
	if instance.ConstraintType == "_COMPLETE" {
		// Save "complete" fet file with "_" prefix in working directory.
		cfile := filepath.Join(bdata.SourceDir, "_"+bdata.Name+".fet")
		err = os.WriteFile(cfile, fet_xml, 0644)
		if err != nil {
			panic("Couldn't write fet file to: " + cfile)
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

	runCmd := exec.CommandContext(ctx,
		//runCmd := exec.Command(
		bdata.Logger.GetConfig("FET"), params...,
	)

	go run(fet_data, runCmd)
	return fet_data
}

type FetTtData struct {
	ifile       string
	fetxml      []byte
	odir        string // the working directory for this instance
	logfile     string
	resultfile  string
	rdfile      *os.File // this must be closed when the subprocess finishes
	reader      *bufio.Reader
	cancel      func()
	fet_timeout bool //
	finished    int
	count       int
}

func (data *FetTtData) Abort() {
	data.cancel()
}

func (data *FetTtData) Clear() {
	//base.Message.Printf("### Remove %s\n", fttd.workingdir)
	os.RemoveAll(data.odir)
	//} else {
	//  base.Message.Printf("### No TtData: %s\n", instance.Tag)
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
		if strings.HasPrefix(e, "signal") || strings.HasPrefix(e, "context") {
			fet_data.finished = -1
		} else {

			//TODO--
			//fmt.Printf("§§§ %s\n", err)
			//fmt.Printf("  *** %s\n", fet_data.ifile)
			//panic("§§§ ...")

			fet_data.finished = -2 // program failed
		}
	} else {
		fet_data.finished = 1
	}
}

// Regexp for reading the progress of a run from the FET log file
var pattern = "time (.*), FET reached ([0-9]+)"
var re *regexp.Regexp = regexp.MustCompile(pattern)

// `Tick` runs in the "tick" loop. Rather like a "tail" function it reads
// the FET progress from its log file, by simply polling for new lines.
func (data *FetTtData) Tick(
	bdata *base.BaseData,
	attdata *autotimetable.AutoTtData,
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
							percent := (count * 100) / int(attdata.NActivities)
							if percent > instance.Progress {
								instance.Progress = percent
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
	if data.finished != 0 {
		if data.rdfile != nil {
			data.rdfile.Close()
		}
		if instance.Progress == 100 {
			instance.RunState = 1
		} else {
			instance.RunState = 2
		}
		if data.finished == -2 {
			bdata.Logger.Error("FET_Failed: %s", bdata.Logger.GetConfig("FET"))
			return
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
			attdata.BlockConstraint[instance.Constraints[0]] = true
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
		panic("Couldn't write fet file to: " + fetfile)
	}
}

// Gather the results of the given run.
func (data *FetTtData) Results(
	bdata *base.BaseData,
	attdata *autotimetable.AutoTtData,
	instance *autotimetable.TtInstance,
) []autotimetable.TtActivityPlacement {
	logger := bdata.Logger
	// Get placements
	xmlpath := data.resultfile
	// Open the XML file
	xmlFile, err := os.Open(xmlpath)
	if err != nil {
		logger.Bug("%v", err)
		return nil
	}
	// Remember to close the file at the end of the function
	defer xmlFile.Close()
	// read the opened XML file as a byte array.
	logger.Info("Reading: %s\n", xmlpath)
	byteValue, _ := io.ReadAll(xmlFile)
	v := fetResultRoot{}
	err = xml.Unmarshal(byteValue, &v)
	if err != nil {
		logger.Bug("XML error in %s:\n %v\n", xmlpath, err)
		return nil
	}
	// Need to prepare the `ActivityPlacment` fields: activities, days,
	// hours and rooms ...
	// ... room conversion
	room2index := map[string]int{}
	for i, r := range attdata.Source.GetRooms() {
		room2index[r.Backend] = i

	}
	// ... day conversion
	day2index := map[string]int{}
	for i, d := range attdata.Source.GetDays() {
		day2index[d.Backend] = i
	}
	// ... hour conversion
	hour2index := map[string]int{}
	for i, h := range attdata.Source.GetHours() {
		hour2index[h.Backend] = i
	}
	// ... activity conversion
	activity2index := map[string]int{}
	for i, a := range attdata.Source.GetActivities() {
		activity2index[a.Backend] = i
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
