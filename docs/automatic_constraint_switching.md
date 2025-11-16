# Automation of Constraint Testing

If the timetable generation process can't even start, there will be an error message which should help find the problem. Sometimes the process will start, but get stuck at a very early stage. In these cases it may not be too difficult to find the problem, but some assistance may be useful. However, it can happen that the computation time is uncomfortably long (whatever that means ...) and the user is uncertain as to whether it will ever complete. Although this is a perfectly normal for more complicated timetables, it can make it difficult to trace problem areas. How long should they wait before deciding the timetable generation is not going to work? And then, what steps need to be taken to attain a timetable that might at least be a useful starting point?

## The aims

 - To deliver some sort of "complete" timetable in a limited time, if necessary by disabling some of the constraints.

 - To identify constraints which are difficult (slowing the timetable generation significantly) or impossible to satisfy. Note that it might be difficult to distinguish between these two cases.

 - To permit, on demand, indefinite generation times, with the possibility of manual interruption. Gradually more and more of the constraints would be enabled and the latest "successful run" would then be available.

## Automation

A common approach to troubleshooting a timetable is to disable and re-enable groups of constraints, aiming to find those constraints that are difficult (or impossible) to satisfy by narrowing down the areas in which they can lie.

With few constraints the generation of solutions will often be very quick, allowing many tests to be carried out in a short time. Automation can be very helpful at this stage. It won't be able to replace the insights of an expert timetable constructor, but can speed up the process even for them. As the run times get longer the advantages may be less obvious, but some automation of the process can still assist less experienced timetable constructors. The possibility of running tests in parallel when multiple processor cores are available can also offer speed gains.

Automation along these lines should be seen as a valuable assistant, but not as a panacea. In general, it can't tell you exactly what needs changing, it can only point to areas where changes might be necessary. Often there will be several – perhaps seemingly unrelated – areas in which changes in the data (constraints) could lead to more computable timetables. Of course, although such a technique can be helpful, experience and analytical expertise are still very desirable qualities in a person charged with the task of constructing a timetable.

## The algorithm used here

The constraints are divided into "hard" ones – which must be satisfied – and "soft" ones – which are more or less desirable (a weighting is specified), but whose non-fulfilment would not necessarily prevent the generation of an acceptable timetable.

In order to get as many constraints as possible enabled, an attempt is made to begin with the "easy" ones. Note that it may not be immediately clear which constraints are "easy"! I don't know whether this is really the best approach, but it seems reasonable. Once more "difficult" constraints are included, the run times will be longer, so fewer tests can be carried out in a given time.

The constraints are further divided into "types" (such as minimum-days-between-lessons or maximum-gaps-per-day). Some of these types have been given priorities, so that they are tested earlier or later in the process. Again, I don't know if this really helps – perhaps in some cases it does, in others not. Also, whether a particular constraint type can be generally regarded as more difficult than another is not clear. Constraints without a priority get the value 0. Positive priorities are handled earlier, negative ones later.

### Multiprocessing

It is assumed that multiple processor cores are available. If not, this algorithm will probably not be very helpful. A minimum of four is suggested. The program will still run if there are fewer processor cores, but up to four processes will be started anyway, so the whole thing will run slower and the timeouts will perhaps be too short.

Also if too many processes are started, efficiency is likely to suffer (more runs are started, only to be terminated after a while without actually contributing anything). This has not been extensively tested, but setting a maximum of about six (when the cores are available) seems to be a good compromise.

The limiting of the number of running processes is achieved by using a queue of active processes. New runs may be initiated at any time, but they will only be actually started if there are "free" processors.

### Monitoring and steering the processes

Overall control of the processes lies in a "tick" loop. Timeouts are managed and the progress (percentage of placed activities) of each run is monitored. Every second, the state of each run is read and necessary actions are taken, such as starting new runs from the queue or switching to the next phase (see below).

There is a timeout for the whole program (as well as for individual timetable generation runs). When this is reached, all runs will be stopped and the latest successful run taken as the result. It is also possible to start the program without a timeout, so that it could – in principle – run indefinitely. The program can then be terminated externally. At least on Linux (other operating systems haven't been tested), using the "terminate" signal (Ctrl-C on the command line), the latest successful run should then be available as a result.

### Phase 0

Initially, three runs are started:

 - the unmodified, fully constrained data set,

 - the same, but with all soft constraints (if any) disabled,

 - a version with all constraints disabled, which will check that it is at all possible to place the activities within the time slots.

Whichever phase is currently active, the successful completion of the fully constrained run will lead to the ending of the whole process.

Phase 0 ends when the unconstrained run completes successfully. This will be used as the basis for phase 1.

If the second run (with all hard, but no soft constraints) completes successfully, phase 1 will be skipped or aborted (if it has already started), and this run will be used as the basis for phase 2.

### Phase 1

In this phase, hard constraints are added type-by-type. Within this phase (as also in phase 2) there are "cycles". In a cycle, all currently inactive constraints are collected and a run is initiated (queued) for each of the types. Thus each of these runs is an attempt to add a number of constraints of a single type to the current (known working) base. Each of these runs has a timeout, which is initially very short (a few seconds).

When a run completes successfully, all other runs within this cycle are terminated and the successful run is taken as the basis for the next cycle.

If a run fails or is timed out, it is split into two (each covering one half of the initial new constraints) and these are added to the back of the run queue. If there was only a single constraint in the run, it is lost from the cycle (but still available to the next cycle).

The handling of timeouts is somewhat flexible – in an attempt to increase overall efficiency. If a run is progressing "too slowly" it can be terminated before its actual timeout. On the other hand, if the timeout is reached, but the run has been progressing really well, the timeout can be postponed for a few ticks.

Phase 1 ends when there are no unsatisfied hard constraints. It is quite possible that this never occurs (at least, not before the program timeout), so phase 2 may never be started.

### Phase 2

This is basically the same as phase 1, but it deals with the soft constraints (all hard constraints now being enabled).

## Possible development areas.

If the unconstrained instance fails, at present the error report comes from the back-end, which may be a bit cryptic and general. Further steps could be taken to trace difficulties within the activity collection, perhaps identifying "difficult" classes or teachers.

At present the soft constraints are treated basically the same as the hard ones (they are just introduced later). If there are difficult hard constraints, the soft constraints may not play a role – they are only added when all the hard constraints have been shown to be satisfiable. But if the difficulties arise in connection with the soft constraints, there may be better ways of handling them. The current approach may produce reasonable results, but no alternatives have been tested yet.

One possibility would be to run the soft constraints as though they were hard. They should still be started later, so as not to interfere with the testing of the hard constraints. Perhaps they could somehow be ordered according to weight. A penalty could be calculated for each successful run based on the weights.
