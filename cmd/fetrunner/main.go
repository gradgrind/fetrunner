/*
`fetrunner` takes timetable specification data from a FET file (".fet" ending)
or a Waldorf-365 timetable data set (JSON, ending "_w365.json") and repeatedly
runs the command-line version of FET with various subsets of the constraints
enabled.

Waldorf-365 data is not handled directly, but is first read into the internal
database structures defined in the "base" package, the root element of which
is `base.DbTopLevel`. This structure is converted to a FET file for the
timetable generation. The correlation of the Walforf-365 elements with their
FET equivalents is achieved by placing pairs of identifiers/references in the
result file (ending "_Result.json").

All processing is done via the `Dispatch` function in the main "fetrunner"
package. This allows the bulk of the code, and especially the API, to be
shared with the "libfetrunner" package, which makes `fetrunner` available
as a C-linked library with a simple JSON interface.

The `autotimetable` package provides the main `fetrunner` algorithm. Its
basic data is in the structure `autotimetable.AutoTtData`, including the
run-time parameters, among other things.

Some further preprocessing of the data specifically for timetable purposes
is performed in the "timetable" package, producing a`timetable.TtData`
structure, which is used by the function `fet.BuildFet` to produce a
`fet.fet_build` structure. The latter contains information specific to the
`fetrunner` FET back-end, including the XML structure of the FET file.

The back-end using FET to generate timetables is set up by a call to
`fet.InitFetBackend` and the actual `fetrunner` algorithm is started
by calling the method `StartGeneration`.

The result-files are saved in the same directory as the input file and are
based on the stem of the input file name. Using "myfile_w365.json" as
input:

  - Log file (myfile_w365.log): Contains error messages and warnings as well
    as information about the steps performed. It can be read continuously
    to monitor progress.

  - Initial FET file (_myfile_w365.fet): The file to be fed to FET with all
    constraints active.

  - Successful FET file (myfile_w365_Result.fet): The last FET file to run
    successfully before the process ended.

  - Result file (myfile_w365_Result.json): A processed view of the results of
    the last successful FET run (with myfile_Result.fet).

  - Other files will be generated temporarily, but removed before the process
    completes.

For FET input files, the process is much simpler, as the data is already
suitably structured, but the results are similar.
*/

package main

import (
	"errors"
	"fetrunner"
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fetrunner/internal/fet"
	"flag"
	"fmt"
	"log"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
)

var (
	logfile *os.File
)

func main() {
	v := flag.Bool("v", false, "print version and exit")
	skip_hard := flag.Bool("h", false, "skip hard constraint testing phase")
	real_soft := flag.Bool("s", false, "the weights of soft constraints are retained")
	timeout := flag.Int("t", 300, "set timeout, s")
	nprocesses := flag.Int("p", 0, "max. parallel processes")
	FET_CL := "fet-cl"
	if runtime.GOOS == "windows" {
		FET_CL = "fet-cl.exe"
	}
	fetpath := flag.String("fet", FET_CL, "FET executable: /path/to/fet-cl")
	tmppath := flag.String("tmp", "", "Folder for temporary files (FET): /path/to/tmp")
	write_fet_file := flag.Bool("xf", false, "write fully-constrained FET file")
	testing := flag.Bool("xt", false, "run in testing mode")
	debug := flag.Bool("xd", false, "run in debug mode")

	flag.Parse()

	if *v {
		fmt.Println("fetrunner version:", fetrunner.VERSION)
		return
	}

	args := flag.Args()
	if len(args) != 1 {
		if len(args) == 0 {
			log.Fatalln("No input file")
		}
		log.Fatalf("Too many command-line arguments:  %+v\n", args)
	}
	abspath, err := filepath.Abs(args[0])
	if err != nil {
		log.Fatalln(err)
	}
	if _, err := os.Stat(abspath); errors.Is(err, os.ErrNotExist) {
		log.Fatalln(err)
	}

	logpath := strings.TrimSuffix(abspath, filepath.Ext(abspath)) + ".log"
	logfile, err = os.OpenFile(logpath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
	if err != nil {
		log.Fatalln(err)
	}
	defer logfile.Close()
	base.LogToFile(logfile)
	defer base.LogStop()

	fetrunner.Dispatch("VERSION")
	fetrunner.Dispatch("TT_PARAMETER TIMEOUT=" + strconv.Itoa(*timeout))
	fetrunner.Dispatch("TT_PARAMETER MAXPROCESSES=" + strconv.Itoa(*nprocesses))
	fetrunner.Dispatch("TT_PARAMETER DEBUG=" + strconv.FormatBool(*debug))
	fetrunner.Dispatch("TT_PARAMETER TESTING=" + strconv.FormatBool(*testing))
	fetrunner.Dispatch("TT_PARAMETER SKIP_HARD=" + strconv.FormatBool(*skip_hard))
	fetrunner.Dispatch("TT_PARAMETER REAL_SOFT=" + strconv.FormatBool(*real_soft))
	fetrunner.Dispatch("TT_PARAMETER WRITE_FET_FILE=" + strconv.FormatBool(*write_fet_file))

	if *tmppath == "" {
		base.SetTmpDir()
	} else {
		// Set base directory for temporary files
		abstmppath, _ := filepath.Abs(*tmppath)
		if abstmppath != *tmppath {
			log.Fatalln("Invalid absolute path:", *tmppath)
		}
		fileInfo, err := os.Stat(abstmppath)
		if errors.Is(err, os.ErrNotExist) || !fileInfo.IsDir() {
			log.Fatalln("Not a directory:", abstmppath)
		}
		if !fetrunner.Dispatch("TMP_PATH " + abstmppath) {
			return
		}
	}

	// Get the path to `fet-cl`, and its version number.
	fetrunner.Dispatch("GET_FET " + *fetpath)
	if fet.FETPATH == "" {
		base.LogError("--NO_FET")
		return
	}
	fetrunner.Dispatch("SET_FILE " + abspath)
	if len(base.DataBase.Name) == 0 {
		return
	}
	if fetrunner.Dispatch("RUN_TT_SOURCE") {
		fetrunner.Dispatch("TT_PRIORITY_CONSTRAINT_TYPES")
		fetrunner.Dispatch("TT_HARD_CONSTRAINTS")
		fetrunner.Dispatch("TT_SOFT_CONSTRAINTS")
		fetrunner.Dispatch("TT_NACTIVITIES")

		go termination() // catch stop signal

		fetrunner.Dispatch("RUN_TT")

		//* TODO-- just testing the new functions
		fmt.Println("Done")
		fetrunner.Dispatch("TT_DAYS")
		fetrunner.Dispatch("TT_HOURS")
		fetrunner.Dispatch("TT_CLASSES")
		fetrunner.Dispatch("TT_TEACHERS")
		fetrunner.Dispatch("TT_ROOMS")
		fetrunner.Dispatch("TT_ACTIVITIES")

		lres := autotimetable.AutoTt.GetLastResult()
		fetrunner.Dispatch("TT_CLASS_PLACEMENTS " + strconv.Itoa(len(lres.Classes)/2))
		fetrunner.Dispatch("TT_TEACHER_PLACEMENTS " + strconv.Itoa(len(lres.Teachers)/2))
		fetrunner.Dispatch("TT_ROOM_PLACEMENTS " + strconv.Itoa(len(lres.Rooms)/2))
		// */
	}
}

func termination() {
	// Catch termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	<-sigChan // wait for signal
	fetrunner.Dispatch("_STOP_TT")
}
