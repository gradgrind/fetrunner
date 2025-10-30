# fetrunner

This is primarily a tool for testing FET-files. It runs multiple instances of FET (the command-line version) on a supplied FET-file with various subsets of the constraints enabled. It is hoped that it can assist in finding difficult (or impossible) constraints. In order to function as intended it needs to be able to run several processes in parallel – it should work with four processor cores, but better results are likely with at least six.

Basically, it starts by running three instances:

 - fully constrained
 - all the hard constraints, but no soft constraints
 - unconstrained

Under normal circumstances these instances run until they complete naturally.

The constraints are divided into types and hard/soft collections.
Normally, the unconstrained instance will complete very quickly. When it does, further FET instances are added to a queue, one instance for each hard constraint type (regardless of how many constraints of this type there are). These will be run when processor cores are available.

When one of these completes, all the others are terminated and a new cycle is begun, taking the successfully completed instance as a new base for further FET instances, one for each of the remaining constraint types. Instances with constraints which are easily satisfied are favoured, timeouts are used to catch difficult constraints. In this way it should be possible to include as many constraints as possible within a limited time.

If an instance is timed out (or fails for some other reason), its list of constraints will be split into two and new FET instances will be started for these.

With each successful run contributing its set of constraints to the new runs, there is always a "best so far" FET instance, which gradually gathers more and more constraints (ideally!).

Only when all the hard constraints have been successfully included (either via the basic hard-only instance or as the result of gradual accumulation of the constraint-type sets) are the soft constraints added, using the same algorithm.

When the time limit for the program is reached, or when it is manually interrupted, the latest successful run is taken as the result.

There are some difficult cases with which `fetrunner` can't help much, because the basic runs take too long, but for many FET-files it can provide some useful information.

## Building the tool

At present, `fetrunner` is a command-line tool only. It is written in Go, which is very portable. I have only tested it on Linux, but it should also work on Windows and MacOS. To compile it, run this in the base directory (assuming the Go compiler has been installed!):

```
go build ./cmd/fetrunner
```

An executable should be produced in the same directory.

## Running the tool

It can be run with just the source file as argument:

```
./fetrunner path/to/fetfile.fet
```

This will normally run for up to five minutes, placing the results in a subdirectory, "path/to/_fetfile":

    Result.fet – the fet-file used to produce the result

    Result.json – the results of the run, including the placements of the activities and the constraints which were deactivated

    run.log – a log file giving information about the run

    _fetfile.fet – should be essentially the same as the original fetfile.fet

The log-file is updated continuously during the run.

There are a few command-line options:

```
fetrunner -help
 ->
  -T    run in testing mode
  -c    enable progress output
  -d    debug
  -h    skip hard constraint testing phase
  -p int
        max. parallel processes
  -t int
        set timeout (default 300)
```

In normal usage, the most useful of these is probably "-t", which sets the overall timeout in seconds.

If it is known that the hard constraints are all satisfiable, the "-h" option can be used to always include the hard constraints and test the sequential addition of just the soft constraints.
