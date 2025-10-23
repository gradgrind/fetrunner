package autotimetable

import (
	"bufio"
	"context"
	"encoding/xml"
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

func (basic_data *BasicData) FetSetup() {
	basic_data.RunTimeBackend = &TtBackend{
		//New: ?
		Run:     runFet,
		Abort:   fetRunAbort,
		Tick:    fetTick,
		Clear:   fetRunClear,
		Tidy:    fetRunTidy,
		Results: fetResults,
	}
}

func fetRunAbort(instance *TtInstance) {
	instance.BackEndData.(*FetTtData).cancel()
}

func fetRunTidy(workingdir string) {
	os.RemoveAll(filepath.Join(workingdir, "tmp"))
}

func fetRunClear(instance *TtInstance) {
	fttd, ok := instance.BackEndData.(*FetTtData)
	if ok {
		//base.Message.Printf("### Remove %s\n", fttd.workingdir)
		os.RemoveAll(fttd.workingdir)
		//} else {
		//	base.Message.Printf("### No TtData: %s\n", instance.Tag)
	}
}

func runFet(basic_data *BasicData, instance *TtInstance) {
	fname := instance.Tag
	dir_n := filepath.Join(basic_data.WorkingDir, "tmp", fname)
	err := os.MkdirAll(dir_n, 0755)
	if err != nil {
		panic(err)
	}
	instance.InstanceDir = dir_n
	stemfile := filepath.Join(dir_n, fname)
	fetfile := stemfile + ".fet"

	// Construct the FET-file
	var fet_xml []byte
	basic_data.PrepareRun(
		basic_data, instance.ConstraintEnabled, &fet_xml)
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
	//cwd := filepath.Dir(fetfile)
	//odir := filepath.Join(cwd, "out")
	odir := filepath.Join(dir_n, "out")
	os.RemoveAll(odir)
	logfile := filepath.Join(odir, "logs", "max_placed_activities.txt")

	ctx, cancel := context.WithCancel(context.Background())
	// Note that it should be safe to call `cancel` multiple times.
	fet_data := &FetTtData{
		finished:   false,
		activities: int(basic_data.NActivities),
		ifile:      fetfile,
		fetxml:     fet_xml,
		//workingdir: cwd,
		workingdir: dir_n,
		odir:       odir,
		logfile:    logfile,
		cancel:     cancel,
	}
	instance.BackEndData = fet_data

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

type FetTtData struct {
	activities int // total number of activities to place
	ifile      string
	fetxml     []byte
	workingdir string
	odir       string
	logfile    string
	rdfile     *os.File // this must be closed when the subprocess finishes
	reader     *bufio.Reader
	cancel     func()
	finished   bool
}

// `fetTick` runs in the "tick" loop. Rather like a "tail" function it reads
// the FET progress from its log file, by simply polling for new lines.
func fetTick(instance *TtInstance) {
	data := *instance.BackEndData.(*FetTtData)
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
						percent := count * 100 / data.activities
						if percent > instance.Progress {
							instance.Progress = percent
							instance.LastTime = instance.Ticks

							base.Report(fmt.Sprintf("%s: %d @ %d\n",
								instance.Tag, percent, instance.Ticks))
						}
					}
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
		}
	}
}

// Gather the results of the given run.
func fetResults(
	basic_data *BasicData,
	instance *TtInstance,
) []ActivityPlacement {
	data := *instance.BackEndData.(*FetTtData)

	// Write FET file at top level of working directory.
	fetfile := filepath.Join(basic_data.WorkingDir, "Result.fet")
	err := os.WriteFile(fetfile, data.fetxml, 0644)
	if err != nil {
		panic("Couldn't write fet file to: " + fetfile)
	}

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

	room2index := map[string]RoomIndex{}
	for _, r := range basic_data.Resources {
		if r.Type == RoomResource {
			room2index[r.Tag] = r.Index
		}
	}
	activities := make([]ActivityPlacement, len(v.Activities))
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
		activities[i] = ActivityPlacement{
			Id:    a.Id,
			Day:   a.Day,
			Hour:  a.Hour,
			Rooms: rooms,
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
	Id        ActivityIndex
	Day       int
	Hour      int
	Room      string
	Real_Room []string
}
