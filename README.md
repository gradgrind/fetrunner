# fetrunner

This is primarily a tool for testing `FET` files. It runs multiple instances of `FET` (the command-line version) on a supplied `FET` file with various subsets of the constraints enabled. The aim is to assist in finding difficult (or impossible) constraints. In order to function as intended it needs to be able to run several processes in parallel – it should work with four processor cores, but better results are likely with at least six.

The basic idea is to behave similarly to a person looking for possible problems in a `FET` file. Initially, three instances are run:

 - fully constrained
 - all the hard constraints, but no soft constraints
 - unconstrained

Under normal circumstances these instances run until they complete naturally.

The constraints are divided into types and hard/soft collections.
Normally, the unconstrained instance will complete very quickly. When it does, further `FET` instances are added to a queue, one instance for each hard constraint type (regardless of how many constraints of this type there are). These will be run when processor cores are available.

When one of these completes, all the others are terminated and a new cycle is begun, taking the successfully completed instance as a new base for further `FET` instances, one for each of the remaining constraint types. Instances with constraints which are easily satisfied are favoured, timeouts are used to catch difficult constraints. In this way it should be possible to include as many constraints as possible within a limited time.

If an instance is timed out (or fails for some other reason), its list of added constraints will be split into two and new `FET` instances will be started for these.

With each successful run contributing its set of constraints to the new runs, there is always a "best so far" `FET` instance, which gradually gathers more and more constraints (ideally!).

Only when all the hard constraints have been successfully included (either via the basic hard-only instance or as the result of gradual accumulation of the constraint-type sets) are the soft constraints added, using the same algorithm.

When the time limit for the program is reached, or when it is manually interrupted, the latest successful run is taken as the result.

There are some difficult cases with which `fetrunner` can't help much, because the basic runs take too long, but for many `FET` files it can provide some useful information.

## Temporary files

`fetrunner` starts many `FET` (`fet-cl`) instances, each of which produces a number of output files. Only a fraction of these are needed by `fetrunner`, and none of them are retained. To reduce wear on SSD storage, these should probably be stored in an in-memory file system (RAM-disk, etc.). Linux has such a file-system "built-in" (at `/dev/shm`), and `fetrunner` uses it for these temporary files. On other operating systems it may be possible to provide something like this, but perhaps only with third-party software.

There are a number of utilities for Windows which can generate RAM-disks. Two free ones which seem to work are [AIMtk](https://sourceforge.net/projects/aim-toolkit) and [OSFMount](https://www.osforensics.com/tools/mount-disk-images.html). Of these OSFMount seems a bit easier to use, but AIMtk can produce a dynamic RAM-disk which only occupies as much memory as is needed – OSFMount allocates a fixed-size block of RAM. However, `fetrunner` would normally need relatively little space, and a few hundred megabytes should be more than enough. When `fetrunner` starts it looks for a "disk" mounted at "R:", so if possible a RAM-disk should be mounted here.

If no such file-system is available and detected, the standard temporary directory for the operating system will be used. With the command-line version of `fetrunner`, it is possible to specify the path to the directory to be used for temporary files using the "-tmp" option. If a Windows system has a RAM-disk mounted at "M:", the option would then be `-tmp M:\`. 

## Command line / program library / GUI

`fetrunner` started life as a command-line tool, written in `Go`. Subsequently `libfetrunner` was added, which makes the functionality available as a program library, using a JSON-based API. There is now also a GUI, written in `C++/Qt`, which uses `libfetrunner` as its back-end.

### Building the command-line tool

`fetrunner`, being written in `Go`, should be very portable. I have tested it on Linux and briefly on Windows, but it should also work on MacOS. To compile it, run this in the base directory (assuming the Go compiler has been installed!):

```
go build ./cmd/fetrunner
```

An executable should be produced in the same directory.

### Running the command-line tool

Important: By default the `FET` command-line executable is expected to be runnable by calling `fet-cl` (`fet-clw.exe` on Windows – see below), i.e. it must be in the user's `PATH`. There is, however, a command line option ("-fet") to specify a different location. The value must be a full, absolute path.

`fetrunner` can be run with just the source file as argument:

```
./fetrunner path/to/fetfile.fet
```

This will normally run for up to five minutes, placing the results in the same directory, "path/to/":

    fetfile_Result.fet – the fet-file used to produce the result

    fetfile_Result.json – the results of the run, including the placements of the activities and the constraints which were deactivated

    fetfile.log – a log file giving information about the run

    _fetfile.fet – should be essentially the same as the original fetfile.fet

The log-file is updated continuously during the run.

There are a few command-line options:

```
fetrunner -help
 ->
  -T	run in testing mode
  -d	debug
  -fet string
    	FET executable: /path/to/fet-cl
  -h	skip hard constraint testing phase
  -p int
    	max. parallel processes
  -t int
    	set timeout (default 300)
  -tmp string
    	Folder for temporary files (FET): /path/to/tmp
  -v	print version and exit
```

In normal usage, the most useful of these is probably "-t", which sets the overall timeout in seconds.

If it is known that the hard constraints are all satisfiable, the "-h" option can be used to always include the hard constraints and test the sequential addition of just the soft constraints.

If multiple instances of `fetrunner` are to be run simultaneously – which is generally inadvisable because of the limited processor cores – each should have a unique source file (name).

### Building the program library

See `libfetrunner/README`.

### Building the GUI

As this is written in `C++` this is more difficult. Perhaps the easiest way is to install the Qt development kit from the Qt website (qt.io). Then run Qt Creator and open the project in the subdirectory `gui` by loading the `CMakeLists.txt` file. See the Qt Creator documentation for further details. Note that `libfetrunner` must be built (as a static library) before building the GUI.

**Special note for Windows users**

The `fet-cl.exe` built when `FET` is built normally is compiled as a console application. This can be used by the command-line `fetrunner`, but if it is used by `fetrunner-gui` a new console will be popped up every time it is called – which is a lot. This makes a real mess! Thus a custom build of `fet-cl` without console output is required. I have given the executable a new name, `fet-clw.exe` to distinguish it; it can be generated as follows:

 - Copy `src/src-cl` to `src/src-clw` and remove `cmdline` from CONFIG in `src/src-clw.pro`.
 
 - Change TARGET in `src/src-clw` to `fet-clw`.

 - Add `src/src-clw.pro` to SUBDIRS in `fet.pro`.
 
 - If `fet` and `fet-cl` don't need to be compiled, `src/src.pro` and `src/src-cl.pro` can be removed from SUBDIRS in `fet.pro`.
 
 - Compile `FET` as usual.

### Building the GUI together with `FET`

It may be more convenient to build `fetrunner` together with `FET`, especially on Windows, where a custom build of `fet-cl` is required anyway. By adding a `qmake` build file to `fetrunner`, this can be achieved in a fairly straightforward way, see `docs/README_GUI.md`.
