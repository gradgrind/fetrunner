/*
fetrunner_W365 produces a FET configuration file from a supplied Waldorf 365
data set (JSON). It then runs the `fetrunner` back-end on this file.

The name of the input file should end with "_w365.json", for example
"myfile_w365.json". This allows consistent automatic naming of the
generated files.

The correlation of the Walforf 365 elements with their FET equivalents is
achieved by placing pairs of identifiers/references in the result file
("myfile_Result.json").

The `autotimetable` package provides the main `fetrunner` algorithm. Its
basic data is in the structure `autotimetable.AutoTtData`, including the
run-time parameters, among other things.

After dealing with the parameters and file paths, the input file is read
and processed so that the data can be stored in a form independent of
Waldorf 365. This form is managed in the "base" package, the primary data
structure being `base.DbTopLevel`.

There are some useful pieces of information which are not contained directly
in the input file, but which can be derived from it. The method
`base.PrepareDb` performs the first of this processing and also checks for
certain errors in the data.

Some further preprocessing of the data specifically for timetable purposes
is performed by `timetable.BasicSetup`, which produces a `timetable.TtData`
structure.

The function `makefet.FetTree` uses the above structures to produce a
`fet.TtRunDataFet` structure containing information specific to the
`fetrunner` FET back-end, including the XML structure of the FET file,
so that modified versions can be produced easily. Also, further information
is added to the `autotimetable.AutoTtData` structure.

The `fetrunner` back-end using FET to generate timetables is set up by the
call to `fet.SetFetBackend` and the actual `fetrunner` algorithm is started
by calling the method `StartGeneration`.

The result-files are saved in the same directory as the input file and are
based on the stem of the input file name. Using "myfile_w365.json" as
input:

  - Log file (myfile.log): Contains error messages and warnings as well as
    information about the steps performed. It can be read continuously
    to monitor progress.

  - Initial FET file (myfile.fet): The file to be fed to FET with all
    constraints active.

  - Successful FET file (myfile_Result.fet): The last FET file to run
    successfully before the process ended.

  - Result file (myfile_Result.json): A processed view of the results of
    the last successful FET run (with myfile_Result.fet).

  - Other files will be generated temporarily, but removed before the process
    completes.
*/
package main

import (
	"errors"
	"fetrunner"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
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
	fetpath := flag.String("fet", "", "/path/to/fet-cl")

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
	if *fetpath != "" {
		do("TT_PARAMETER", "FETPATH", *fetpath)
	}

	if !do("GET_FET") {
		return
	}

	if !do("SET_FILE", abspath) {
		return
	}
	do("RUN_TT_SOURCE")

	go fetrunner.Termination()
	var wg sync.WaitGroup
	wg.Go(func() { fetrunner.RunLoop(do) })
	wg.Wait()
}

func do(op string, data ...string) bool {
	strs, ok := fetrunner.Do(op, data...)
	for _, s := range strs {
		logfile.WriteString(s + "\n")
	}
	return ok
}
