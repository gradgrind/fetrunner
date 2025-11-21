/*
fetrunner runs multiple instances of FET with various sets of constraints enabled.

The `autotimetable` package provides the main algorithm. Its basic data is in
the structure `autotimetable.BasicData`, including the run-time parameters,
among other things.

After dealing with the parameters and file paths, the input file is read by
calling `fet.FetRead`, which produces an `fet.TtRunDataFet` structure
containing information specific to the FET back-end, including the XML
structure of the input FET file, so that modified versions can be produced
easily. Also, further information is added to the `autotimetable.BasicData`
structure.

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
	"fetrunner/fet"
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
			log.Fatalln("ERROR: No input file")
		}
		log.Fatalf("ERROR: Too many command-line arguments:\n  %+v\n", args)
	}
	abspath, err := filepath.Abs(args[0])
	if err != nil {
		log.Fatalf("ERROR: Couldn't resolve file path: %s\n", args[0])
	}
	if !strings.HasSuffix(strings.ToLower(abspath), ".fet") {
		log.Fatalf("ERROR: Source file without '.fet' ending: %s\n", abspath)
	}
	if _, err := os.Stat(abspath); errors.Is(err, os.ErrNotExist) {
		log.Fatalf("ERROR: Source file doesn't exist: %s\n", abspath)
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

	fet.FetRead(bdata, abspath)

	// Set up FET back-end and start processing
	fet.SetFetBackend(bdata)
	bdata.StartGeneration(*timeout)
}
