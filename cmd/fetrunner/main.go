/*
	TODO

fetrunner runs multiple instances of FET with various sets of constraints
enabled.

The files produced are saved in the same directory as the input file:

  - Log file: Contains error messages and warnings as well as information
    about the steps performed. The file name (given "myfile_w365.json" as
    input) is "myfile_w365.log" – just the ending is changed.

  - FET file: The file to be fed to FET. The standard name (given
    "myfile_w365.json" as input) is "myfile.fet". However, by supplying a
    "FetFile" property (without the ".fet" ending) in the "FetData" object,
    this can be changed.

  - Map file: Correlates the FET Activity numbers to the Waldorf 365 Lesson
    references ("Id"). The standard name (given "myfile_w365.json" as input)
    is "myfile.map". However, by supplying a "MapFile" property
    (without the ".map" ending) in the "FetData" object, this can be changed.

Note that, at present, the Activity and Room objects in the FET file have the
corresponding Waldorf 365 references ("Id") in their "Comments" fields.

Firstly, the input file is read and processed so that the data can be stored
in a form independent of Waldorf 365. This form is managed in the [base]
package, the primary data structure being [base.DbTopLevel].

There are some useful pieces of information which are not stored directly
in the basic data loaded from an input file, but which can be derived from it.
The method [base.PrepareDb] performs the first of this processing and also
checks for certain errors in the data.

For processing of timetable information there are further useful data
structures which can be derived from the input data. This information is
handled primarily in the [ttbase] package, its primary data structure being
[ttbase.TtInfo].

A further stage of processing the timetable data is handled by the method
[ttbase.PrepareCoreData]. This builds further data structures representing
the allocation of resources, so that a number of errors in the data can be
detected.

Finally, the data structures are used by the function [fet.MakeFetFile] to
produce the XML-structure of the FET file and the reference mapping
information to be stored in the map file.
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

	cdata := fet.SetInputFet()

	stempath := strings.TrimSuffix(abspath, filepath.Ext(abspath))

	// May want to change this with a different back-end ...
	workingdir := stempath + "_fet"
	os.RemoveAll(workingdir)
	err = os.MkdirAll(workingdir, 0755)
	if err != nil {
		panic(err)
	}

	logpath := filepath.Join(workingdir, "run.log")
	base.OpenLog(logpath)

	var source autotimetable.TtSource
	source, err = fet.FetRead(cdata, abspath)
	if err != nil {
		panic(err)
	}
	//_ = x
	//TODO-- This is just for testing FET backend
	source.(*fet.FetDoc).WriteFET(stempath + "_mod.fet")

	autotimetable.RunBackend = fet.RunFet

	bdata.StartGeneration(cdata, *timeout)

	//db.SaveDb(stempath + "_DB2.json")
}
