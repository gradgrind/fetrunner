# Introducing `fetrunner`

This is primarily a tool for testing `FET` files. It runs multiple instances of `FET` (the command-line version) on a supplied `FET` file with various subsets of the constraints enabled. The aim is to assist in finding difficult (or impossible) constraints. In order to function as intended it needs to be able to run several processes in parallel – it should work with four processor cores, but better results are likely with at least six.

![fetrunner-gui](./images/Screenshot_00.png)

## `FET`

[`FET`](https://lalescu.ro/liviu/fet/) is a free timetable generator program for educational establishments. It is widely used and very good at what it does. However, in the case of timetable data which "doesn't work" (because of conflicting constraints), it can sometimes be difficult to find where the problem lies. Also, with some data (lessons/activities and constraints) the calculation of a "solution" (a conflict-free timetable) can take a very long time. Whilst working on a timetable, it can be useful to know which constraints may be difficult to fulfil, without waiting a long time for `fet` to complete (or not ...).

`fetrunner` aims to produce a "solution" within a specified time, if necessary by deactivating some of the constraints. The result is a "known working" `FET` file (possibly including deactivated constraints). There is also a log file, which is updated continually during the process, showing some details of the progress, and a JSON file containing the activity placements from the "successful" `FET` run together with information about the "failed" constraints. In the GUI version of `fetrunner`, the log is not output as a file, but is used to update the interface (and is also available to view, if desired).

## Using `fetrunner`

**Not running on Linux?**

`fetrunner` produces many temporary files, which might cause excessive wear on an SSD. See [Temporary files](./temporary_files.md) for ways to avoid this.

## GUI or command-line

The GUI ([Using the GUI](./using_the_gui.md)) allows selection of `FET` files and processing parameters in a fairly straightforward way. It also shows the progress of a run dynamically. In some cases, however, the command-line tool may be more convenient. To use this, see [Running the command-line tool](#running-the-command-line-tool).

## How to understand the results of a `fetrunner` run

`fetrunner` produces at least two result files in the same folder as the source file: from "xxx.fet", these would be "xxx_Result.fet" and "xxx_Result.json". The command-line version also produces a log file, "xxx.log". In the GUI version, the current log information is displayed in one of the tabs (older log information is not retained).

If you open the "xxx_Result.fet" file in `FET`, you can see that some of the constraints have been deactivated, e.g. in the "Time" (and perhaps "Space") tab select "All". These are the ones which `fetrunner` decided were "difficult" or impossible.

At the end of the log, there should be a summary of the accepted constraints.

The "xxx_Result.json" file may contain some information about why a constraint was rejected, but it is rather intended for reading by other software. Nevertheless, a web browser (e.g. `Firefox`) can display this file quite neatly.

The fact that `fetrunner` has deactivated a constraint doesn't mean that the constraint is necessarily impossible, though it may be (at least, in combination with other constraints). Another run, perhaps with a different timeout, might give a different result. The results show constraints whose removal makes it easier for `FET` to generate a timetable. If these constraints are important, it may be necessary to change other constraints which somehow interact with the shown ones – finding these may not be easy ...

Looking at a timetable generated (in `FET`) from the "xxx_Result.fet" file (which is now known to be possible!) and at the deactivated constraints will – I hope – help you to discover how you might need to modify your data (activities and constraints) in order to get an acceptable result.

Actually, there is a generated timetable (from "xxx_Result.fet") in "xxx_Result.json", but I haven't written the software to display this (yet).

## Running the command-line tool

Important: By default the `FET` command-line executable is expected to be in the same directory as the `fetrunner` executable, or else runnable by calling `fet-cl` (on Windows the executable is `fet-cl.exe`), i.e. in the user's `PATH`. There is, however, a command line option ("-fet") to specify a different location – the value must be a full, absolute path.

`fetrunner` can be run with just the source file as argument:

```
./fetrunner path/to/fetfile.fet
```

This will normally run for up to five minutes, placing the results in the same directory as the source file, "path/to/" in the case of the above command:

    `fetfile_Result.fet` – the `FET` file used to produce the result

    `fetfile_Result.json` – the results of the run, including the placements of the activities and the constraints which were deactivated

    `fetfile.log` – a log file giving information about the run

    `_fetfile.fet` – (optional, primarily for test purposes, with -xf option) should be essentially the same as the original `fetfile.fet`, but the constraints are indexed (in their "Comments" field) and the soft constraints are made "hard" (a derived weight, not the same as the original `FET` "Weight_Percentage" is added to the "Comments" field)

The log-file is updated continually during the run, so it is possible to monitor progress by reading this file.

A run can be stopped prematurely by pressing `Ctrl-C`. This will probably take a couple of seconds to work, as it tries to tidy up. The result files will be produced from the current state.

There are a few command-line options:

```
fetrunner -help
 ->
  -fet string
        FET executable: /path/to/fet-cl
  -h    skip hard constraint testing phase
  -p int
        max. parallel processes
  -s    the weights of soft constraints are retained
  -t int
        set timeout, s (default 300)
  -tmp string
        Folder for temporary files (FET): /path/to/tmp
  -v    print version and exit
  -xd
        run in debug mode
  -xf
        write fully-constrained FET file
  -xt
        run in testing mode
```

If it is known that the hard constraints are all satisfiable, the "-h" option can be used to always include the hard constraints (the unconstrained instance is not run) and test the sequential addition of just the soft constraints.
