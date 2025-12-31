/*
`fetrunner` takes timetable specification data from a FET file (".fet" ending)
or a Waldorf-365 timetable data set (JSON, ending "_w365.json") and repeatedly
runs the command-line version of FET with various subsets of the constraints
enabled.

Waldorf-365 data is not handled directly, but is first read into the internal
database structures defined in the "base" package, the root element of which
is `base.DbTopLevel`. This structure is converted to a FET file, using the
package "makefet", for the timetable generation. The correlation of the
Walforf-365 elements with their FET equivalents is achieved by placing pairs
of identifiers/references in the result file (ending "_Result.json").

All processing is done via the `Dispatch` function in the main "fetrunner"
package. This allows the bulk of the code, and especially the API, to be
shared with the "libfetrunner" package, which makes `fetrunner` available
as a C-linked library with a simple JSON interface.

The `autotimetable` package provides the main `fetrunner` algorithm. Its
basic data is in the structure `autotimetable.AutoTtData`, including the
run-time parameters, among other things.

Some further preprocessing of the data specifically for timetable purposes
is performed in the "timetable" package: `timetable.BasicSetup` produces a
`timetable.TtData` structure, which is used by the function `makefet.FetTree`
to produce a `fet.TtRunDataFet` structure. The latter contains information
specific to the `fetrunner` FET back-end, including the XML structure of the
FET file.

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
	"fetrunner/internal/base"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var (
	logfile *os.File
)

func main() {
	v := flag.Bool("v", false, "print version and exit")
	testing := flag.Bool("T", false, "run in testing mode")
	skip_hard := flag.Bool("h", false, "skip hard constraint testing phase")
	timeout := flag.Int("t", 300, "set timeout")
	nprocesses := flag.Int("p", 0, "max. parallel processes")
	debug := flag.Bool("d", false, "debug")
	fetpath := flag.String("fet", "", "FET executable: /path/to/fet-cl")
	tmppath := flag.String("tmp", "", "Folder for temporary files (FET): /path/to/tmp")

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

	do("TT_PARAMETER", "TIMEOUT", strconv.Itoa(*timeout))
	do("TT_PARAMETER", "MAXPROCESSES", strconv.Itoa(*nprocesses))
	do("TT_PARAMETER", "DEBUG", strconv.FormatBool(*debug))
	do("TT_PARAMETER", "TESTING", strconv.FormatBool(*testing))
	do("TT_PARAMETER", "SKIP_HARD", strconv.FormatBool(*skip_hard))

	if *tmppath != "" {
		// Set base directory for temporary files
		abstmppath, _ := filepath.Abs(*tmppath)
		if abstmppath != *tmppath {
			log.Fatalln("Invalid absolute path:", *tmppath)
		}
		fileInfo, err := os.Stat(abstmppath)
		if errors.Is(err, os.ErrNotExist) || !fileInfo.IsDir() {
			log.Fatalln("Not a directory:", abstmppath)
		}
		if !do("TMP_PATH", abstmppath) {
			return
		}
	}

	// Get the path to `fet-cl`, and its version number.
	if *fetpath == "" {
		*fetpath = "-"
	}
	strs, ok := fetrunner.Do("GET_FET", *fetpath)
	okv := false
	for _, s := range strs {
		logfile.WriteString(s + "\n")
		if strings.Contains(s, "FET_VERSION=") {
			okv = true
		}
	}
	if !ok {
		return
	}
	if !okv {
		logfile.WriteString(base.ERROR.String() + " NO_FET\n")
		return
	}

	if !do("SET_FILE", abspath) {
		return
	}
	do("RUN_TT_SOURCE")

	do("HARD_CONSTRAINTS")
	do("SOFT_CONSTRAINTS")

	go fetrunner.Termination()
	fetrunner.RunLoop(do)
}

func do(op string, data ...string) bool {
	strs, ok := fetrunner.Do(op, data...)
	for _, s := range strs {
		logfile.WriteString(s + "\n")
	}
	return ok
}
