/*
`fetrunner_W365` produces a FET configuration file from a supplied Waldorf 365
data set (JSON). It then runs `fetrunner` on this file.

The name of the input file should ideally end with "_w365.json", for example
"myfile_w365.json". This will enable a consistent automatic naming of the
generated files.

The correlation of the Walforf 365 elements and their FET equivalents is
achieved by placing the the Waldorf 365 references ("Id") in the FET
"Comments" fields.

Firstly, the input file is read and processed so that the data can be stored
in a form independent of Waldorf 365. This form is managed in the "db"
package, the primary data structure being `db.DbTopLevel`.

There are some useful pieces of information which are not stored directly
in the basic data loaded from an input file, but which can be derived from it.
The method `db.PrepareDb` performs the first of this processing and also
checks for certain errors in the data.

The function `makefet.MakeFetFile` uses the above structures to produce the
XML-structure of the FET file which is generated.
*/
package main

import (
	"fetrunner/autotimetable"
	"fetrunner/base"
	"fetrunner/db"
	"fetrunner/fet"
	"fetrunner/timetable"
	"fetrunner/w365tt"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

const VERSION = "0.2.1"

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
		fmt.Printf("fetrunner version %s\n", VERSION)
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

	db0 := db.NewDb()
	w365tt.LoadJSON(db0, abspath)
	db0.PrepareDb()

	db0.SaveDb(filepath.Join(d1, f1+"_DB.json"))

	// Make FET file
	//TODO
	tt_data := timetable.BasicSetup(db0)
	base.Report(fmt.Sprintf("Atomic Groups: %d\n",
		len(tt_data.AtomicGroups)))
	base.Report(fmt.Sprintf("Teachers: %d\n",
		len(db0.Teachers)))
	base.Report(fmt.Sprintf("Rooms: %d\n",
		len(db0.Rooms)))
	base.Report(fmt.Sprintf("Activities: %d\n",
		len(tt_data.Activities)-1))

	// Process the FET file

	bdata.Source, err = fet.FetRead(bdata, abspath)
	if err != nil {
		log.Fatal(err)
	}

	bdata.BackendInterface = fet.SetFetBackend(bdata)

	bdata.StartGeneration(*timeout)
}
