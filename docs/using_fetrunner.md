# Using `fetrunner`

The GUI ([Using the GUI](./using_the_gui.md)) allows selection of `FET` files and processing parameters in a fairly straightforward way. It also shows the progress of a run dynamically and can display resulting timetables. In some particular cases, however, the command-line tool may be more convenient. To use this, see [Running the command-line tool](./using_the_cli.md).

## How to understand the results of a `fetrunner` run

`fetrunner` produces at least two result files in the same folder as the source file: from "xxx.fet", these would be "xxx_Result.fet" and "xxx_Result.json". The command-line version also produces a log file, "xxx.log". In the GUI version, the current log information is displayed in one of the tabs (older log information is not retained).

If you open the "xxx_Result.fet" file in `FET`, you can see that some of the constraints have been deactivated, e.g. in the "Time" (and perhaps "Space") tab select "All". If they weren't already deactivated in the file provided as input to `fetrunner`, these are the ones which `fetrunner` decided were "difficult" or impossible.

At the end of the log, there should be a summary of the accepted constraints.

The "xxx_Result.json" file may contain some information about why a constraint was rejected, but it is rather intended for reading by other software. Nevertheless, a web browser (e.g. `Firefox`) can display this file quite neatly.

The fact that `fetrunner` has deactivated a constraint doesn't mean that the constraint is necessarily impossible, though it may be (at least, in combination with other constraints). Another run, perhaps with a different timeout, might give a different result. The results show constraints whose removal makes it easier for `FET` to generate a timetable. If these constraints are important, it may be necessary to change other constraints which somehow interact with the shown ones – finding these may not be easy ...

Looking at the resulting timetable in `fetrunner` GUI (which uses the placemens in "xxx_Result.json"), or by running `FET` on the "xxx_Result.fet" file (which is now known to be possible!), together with the deactivated constraints will – I hope – help you to discover how you might need to modify your data (activities and constraints) in order to get an acceptable result.

