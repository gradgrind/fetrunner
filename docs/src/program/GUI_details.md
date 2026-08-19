# Some details of the GUI code

The GUI acts as a front-end to the actual `fetrunner` code, which is included as a static library, `libfetrunner`. This library has a very simple interface consisting of just two functions. Note that the use of the terms "front-end" and "back-end" here is distinct from their use in connection with the timetable generation code, where "back-end" refers to the actual timetable generation function (currently only `FET` is supported).

## Communicating with the back-end

`backend.cpp` and `backend.h` contain the code for handling `libfetrunner`.

The function `Backend::op` builds a single string by joining the actual command string with the arguments to the call using "|" as separator. This is used as the argument to the library function `void FetRunnerCommand(char *)`. Then it reads all logged lines, one by one, using `Backend::readlogline`, until one of the terminating lines is read.

`Backend::op` returns a list of "result" lines, parsed into key/value pairs. Also the lines reporting errors are collected (up to a limit) and passed by the `Backend::error` signal to the GUI, which can then show a pop-up message.

`Backend::readlogline` reads a single log entry using the other library function, `char* FetRunnerReadLog()`. Apart from return this entry as a key/value pair, it passes the entry to the log display widget, determining the colour according to the values in the key lookup table, `colours` (in `backend.cpp`).

Note that there are two types of key/value pair here! Initially, a log entry is split at the first space character to provide key and value. If the entry consists of only a single word, that will be the key, the value will be empty. In `Backend::op` the "result" entries ("$ key=value") have their value ("key=value") split at the "=" character to provide the result key/value pairs which are returned from the call to `Backend::op`.

## Dynamic updating during a run

Because it waits for log entries before returning, `Backend::op` it is only suitable for commands which will be executed quickly (otherwise the GUI would hang). This works for most of the available commands, but the command "RUN_TT", which starts the actual `fetrunner` run, takes a long time. When a command is really finished and it will generate no further log entries, "---" is logged. But there is also the log entry "***" which signals the beginning of the long-running part. "`Backend::op`" also returns when it reads this log entry. Subsequent entries are then handled in a separate thread.

In `threadrun.cpp` and `threadrun.h` there is a worker thread, `RunThreadWorker::ttrun()` which reads from the log using `Backend::readlogline` until it reads the completion signal "---". It acts on the various log entries indicating progress by emitting corresponding signals. These signals are forwarded to signals of the `RunThreadController`, which encapsulates the handling of the worker thread, isolating it from the main code. These signals are thus connected indirectly to methods (slots) of `FetRunner` with the same name which update the corresponding display elements.

 - ".TICK" -> `FetRunner::ticker()`

 - ".NCONSTRAINTS" -> `FetRunner::nconstraints()`

 - ".PROGRESS" -> `FetRunner::iprogress()`

 - ".START" -> `FetRunner::istart()`

 - ".END" -> `FetRunner::iend()`

 - ".ACCEPT" -> `FetRunner::iaccept()`

 - ".ELIMINATE" -> `FetRunner::ieliminate()`
