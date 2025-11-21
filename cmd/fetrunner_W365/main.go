/*
fetrunner_W365 produces a FET configuration file from a supplied Waldorf 365
data set (JSON). It then runs the `fetrunner` back-end on this file.

The name of the input file should ideally end with "_w365.json", for example
"myfile_w365.json". This will enable a consistent automatic naming of the
generated files.

The correlation of the Walforf 365 elements with their FET equivalents is
achieved by placing pairs of identifiers/references in the result file
("Result.json").

The `autotimetable` package provides the main `fetrunner` algorithm. Its
basic data is in the structure `autotimetable.BasicData`, including the
run-time parameters, among other things.

After dealing with the parameters and file paths, the input file is read
and processed so that the data can be stored in a form independent of
Waldorf 365. This form is managed in the "db" package, the primary data
structure being `db.DbTopLevel`.

There are some useful pieces of information which are not contained directly
directly in the input file, but which can be derived from it. The method
`db.PrepareDb` performs the first of this processing and also checks for
certain errors in the data.

Some further preprocessing of the data specifically for timetable purposes
is performed by `timetable.BasicSetup`, which produces a `timetable.TtData`
structure.

The function `makefet.FetTree` uses the above structures to produce an
`fet.TtRunDataFet` structure containing information specific to the
`fetrunner` FET back-end, including the XML structure of the FET file,
so that modified versions can be produced easily. Also, further information
is added to the `autotimetable.BasicData` structure.

The `fetrunner` back-end using FET to generate timetables is set up by the
call to `fet.SetFetBackend` and the actual `fetrunner` algorithm is started
by calling the method `StartGeneration`.

The result-files are saved in a (new) subdirectory of the directory of the
input file. The name of this subdirectory is based on the name of the input
file.

  - Log file (run.log): Contains error messages and warnings as well as
    information about the steps performed. It can be read continueously
    to monitor progress.

  - Initial FET file: The file to be fed to FET with all constraints active.
    Its name is based on that of the input file, the contents should be
    essentiallly the same, but the constraints are tagged with identifiers
    in their "Comments" fields and there there may be some minor formatting
    differences.

  - Successful FET file (Result.fet): The last FET file to run successfully
    before the process ended.

  - Result file (Result.json): A processed view of the results of the last
    successful FET run (with Result.fet).

  - Other files will be generated temporarily, but removed before the process
    completes.
*/
package main

import (
	"errors"
	"fetrunner/autotimetable"
	"fetrunner/base"
	"fetrunner/db"
	"fetrunner/fet"
	"fetrunner/makefet"
	"fetrunner/timetable"
	"fetrunner/w365tt"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	bdata := &autotimetable.BasicData{}
	bdata.SetParameterDefault()

	v := flag.Bool("v", false, "print version and exit")
	flag.BoolVar(&base.CONSOLE, "c", false, "enable progress output")
	flag.BoolVar(&bdata.Parameters.TESTING, "T", false, "run in testing mode")
	flag.BoolVar(&bdata.Parameters.SKIP_HARD, "h", false,
		"skip hard constraint testing phase")
	timeout := flag.Int("t", 300, "set timeout")
	nprocesses := flag.Int("p", 0, "max. parallel processes")
	debug := flag.Bool("d", false, "debug")

	flag.Parse()

	if *v {
		fmt.Printf("fetrunner version %s\n", base.VERSION)
		return
	}

	if *nprocesses > 0 {
		bdata.Parameters.MAXPROCESSES = *nprocesses
	}
	if *debug {
		bdata.Parameters.DEBUG = true
	}

	args := flag.Args()
	if len(args) != 1 {
		if len(args) == 0 {
			log.Fatalln("ERROR* No input file")
		}
		log.Fatalf("*ERROR* Too many command-line arguments:\n  %+v\n", args)
	}
	abspath, err := filepath.Abs(args[0])
	if err != nil {
		log.Fatalf("*ERROR* Couldn't resolve file path: %s\n", args[0])
	}
	if !strings.HasSuffix(strings.ToLower(abspath), ".json") {
		log.Fatalf("*ERROR* Source file without '.json' ending: %s\n", abspath)
	}
	if _, err := os.Stat(abspath); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("*ERROR* Source file doesn't exist: %s\n", abspath)
	}

	f1 := filepath.Base(strings.TrimSuffix(abspath, filepath.Ext(abspath)))
	d1 := filepath.Dir(abspath)
	workingdir := filepath.Join(d1, "_"+f1)
	os.RemoveAll(workingdir)
	err = os.MkdirAll(workingdir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	bdata.WorkingDir = workingdir

	logger := base.NewLogger()
	logpath := filepath.Join(workingdir, "run.log")
	go base.LogToFile(logger, logpath)
	bdata.Logger = logger

	db0 := db.NewDb(logger)
	w365tt.LoadJSON(db0, abspath)
	db0.PrepareDb()

	db0.SaveDb(filepath.Join(d1, f1+"_DB.json"))

	// Make FET file
	tt_data := timetable.BasicSetup(db0)
	/*
		base.Report(fmt.Sprintf("Atomic Groups: %d\n",
			len(tt_data.AtomicGroups)))
		base.Report(fmt.Sprintf("Teachers: %d\n",
			len(db0.Teachers)))
		base.Report(fmt.Sprintf("Rooms: %d\n",
			len(db0.Rooms)))
		base.Report(fmt.Sprintf("Activities: %d\n",
			len(tt_data.Activities)))
	*/
	makefet.FetTree(bdata, tt_data)

	// Set up FET back-end and start processing
	fet.SetFetBackend(bdata)
	bdata.StartGeneration(*timeout)
}
