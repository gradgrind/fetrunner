# Threads and Processes

`fetrunner` is at its very core multi-threaded. It runs several timetable generations in parallel, seeking those that complete most quickly. Each timetable generation has its own thread which runs `FET` (`fet-cl`) as a blocking sub-process. As not all of these sub-process will run to completion, it must be possible to cancel them from another thread. In Go, so-called "goroutines" are used to manage threads.

## Running `FET` on a data set

The function `RunBackend` in `fet/runfet.go` is called to execute `FET` on a set of data supplied as a `TtInstance` structure. It first constructs a `FET` file from the `TtInstance`, then, by using `exec.CommandContext` together with a `context.WithCancel` to run `FET` on this file, a function is made available which can be called to abort the `FET` process. The actual process is started in a separate goroutine, allowing `RunBackend` to return immediately. It returns a data structure (`FetTtData`), which allows the progress of the `FET` process to be monitored and controlled.


