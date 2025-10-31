/*
	fetrunner runs multiple instances of FET with various sets of constraints enabled.

The files produced are saved in a (new) subdirectory of the directory of the
input file. This name of this subdirectory is based on the name of the input
file.

  - Log file (run.log): Contains error messages and warnings as well as
    information about the steps performed.

  - Initial FET file: The file to be fed to FET with all constraints active.
    Its name is based on that of the input file, the contents should be
    effectovely the same, though there may be some formatting differences.

  - Successful FET file (Result.fet): The last FET file to run successfully
    before the process ended.

  - Result file (Result.json): A processed view of the results of the last
    successful FET run (with Result.fet).

  - Other files will be generated temporarily, but removed before the process
    completes.
*/

package main

import (
	"fetrunner/autotimetable"
	"fetrunner/base"
	"fetrunner/fet"
	"flag"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func main() {

	bdata := &autotimetable.BasicData{}
	bdata.SetParameterDefault()

	flag.BoolVar(&base.CONSOLE, "c", false, "enable progress output")
	flag.BoolVar(&bdata.Parameters.TESTING, "T", false, "run in testing mode")
	flag.BoolVar(&bdata.Parameters.SKIP_HARD, "h", false,
		"skip hard constraint testing phase")
	timeout := flag.Int("t", 300, "set timeout")
	nprocesses := flag.Int("p", 0, "max. parallel processes")
	debug := flag.Bool("d", false, "debug")

	flag.Parse()

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

	f1 := filepath.Base(strings.TrimSuffix(abspath, filepath.Ext(abspath)))
	d1 := filepath.Dir(abspath)
	workingdir := filepath.Join(d1, "_"+f1)
	os.RemoveAll(workingdir)
	err = os.MkdirAll(workingdir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	bdata.WorkingDir = workingdir

	logpath := filepath.Join(workingdir, "run.log")
	base.OpenLog(logpath)

	bdata.Source, err = fet.FetRead(bdata, abspath)
	if err != nil {
		log.Fatal(err)
	}

	bdata.BackendInterface = fet.SetFetBackend(bdata)

	bdata.StartGeneration(*timeout)
}
