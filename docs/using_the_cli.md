# Running the command-line tool

**Important**: By default the `FET` command-line executable is expected to be in the same directory as the `fetrunner` executable, or else runnable by calling `fet-cl` (on Windows the executable is `fet-cl.exe`), i.e. in the user's `PATH`. There is, however, a command line option ("-fet") to specify a different location – the value must be a full, absolute path.

`fetrunner` can be run with just the source file as argument:

```
./fetrunner path/to/fetfile.fet
```

This will normally run for up to five minutes, placing the results in the same directory as the source file, "path/to/" in the case of the example command above:

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
