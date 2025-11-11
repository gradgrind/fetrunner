# `fetrunner`

When trying to produce a timetable with FET, it can happen that the generation step doesn't run smoothly. In some cases it won't even start, returning an error message which may help in finding the unsatisfiable condition in the configuration file (`xxx.fet`) , though sometimes the message may seem a bit cryptic. In other cases the process will start, but get stuck at an early stage. In other cases the progress will be very slow, so that it is not clear when (or if!) the process will be completed.

`fetrunner` is a utility which attempts to assist in finding the places in the configuration file which lead to these problems. It does this by repeatedly running the command-line version of FET, each time with a different subset of the conditions specified in the configuration file.

Somewhat as a side-product `fetrunner` also produces a fully populated timetable from the last successful configuration. That means that probably not all the constraints are satisfied, so it might not be all that useful, but there might be circumstances under which it is helpful. Only if even the unconstrained configuration fails to run will there be no such placement data.

## How it works

Basically it starts by deactivating all the constraints. If even that doesn't help, there may be an error message from FET which can offer some useful information.

Then it activates a single type of constraint, which means all the individual constraints of a particular type, and starts a generation run. This is repeated, in principle, for all the constraint types used in the source file (see [[#Parallel processing]]).

When one of these generation runs completes successfully, all the others are stopped and discarded. This first successful run is now taken as a new basis. For each of the remaining constraint types a new generation run is started, with the new constraint type being added to the successfully completed one. Again, the first successfully completing run is taken as a new basis, and so on until all constraint types have been enabled.

Well, for "easy" configuration files that might work within a reasonable time, but there would not be much point, because we might as well just run the original configuration file with all constraints enabled, which would be a lot quicker. It becomes interesting when one or more constraints are causing errors or holding up the generation run in some other way.

Normally the first stages (the addition of the quickest constraint types) will run quickly, within a few seconds, allowing most of the constraints to be enabled in a fairly short time. But the longer the run times, the more difficult it is to cover all the remaining constraint types in a given overall run time – some "easy" constraints might never get tested.

There are essentially two things that need dealing with. Perhaps the easiest is when a constraint blocks the generation completely, in such a way that FET returns an error message. The other is when the generation run gets stuck or progresses only very slowly (it is not always easy to distinguish the two). At first, while there are still constraint types which can be added whose runs complete quickly, non-completing runs are no problem because they will simply be discarded. But at some point, only the difficult constraint types remain.

### Hard and soft constraints

A constraint is categorised as "hard" if it must be satisfied in an acceptable timetable. `fetrunner` loosens this a bit in order to get all the activities placed in a given time. This result is then presumably not yet acceptable, but may be helpful. A "soft" constraint is one which should be satisfied "if possible", but a resulting timetable may be acceptable even if the constraint is not satisfied. In FET hard constraints have "Weight" 100 (%), Anything less and the constraint is considered soft (even if, say, 99.99% is not particularly soft).

`fetrunner` takes the distinction between hard and soft constraints into account in that it doesn't touch the soft constraints (it leaves them deactivated) until all the hard constraints have been satisfied. Thus, it is possible that the soft constraints never get tested.

### Handling "difficult" constraint types

Note that the difficult constraint types may not always be the same, it depends on the data. Nevertheless, there may be certain constraint types which turn out to be difficult more often in practice.

The algorithm uses various tests to determine whether a set of constraints (of one type) is "difficult". These have been developed more or less by trial and error, and may not be at all optimal, if indeed there is such a thing in this case. In essence there is a combination of rather flexible "timeouts" and estimates of whether a run is getting "stuck". Two sorts of "getting stuck" seem to prevail – some runs stop progressing soon after starting, others just progress very slowly.

It is to be expected that the run times will gradually increase as more and more constraints are enabled. To compensate for this the basic timeouts are gradually increased, taking into account the run time of the latest "basis".

When a run trialling a set of constraints is determined to be "stuck", it is stopped and divided into two halves. This is essentially a binary search for difficult constraints within a set, so that – eventually – the individual constraints which are not difficult will be accepted (enabled) and the failing ones can be reported.

If at some stage of the overall proceedings, the available "processes" (threads, processor cores) are not all used, a set of constraints may be split even before it has timed out.

### Parallel processing

The number of generation runs which are started depends on the number of processor cores available. Experiments indicate, however, that there may be something like an optimal number, too many processes perhaps even reducing the overall performance. To keep the main part of the algorithm separate from the handling of parallel processes, a process queue is used. Any number of generation runs may be added to the queue, but a run will only be actually started if one of the limited number of processes is available.

As the algorithm depends on multiple processes running simultaneously, the effectiveness depends on there being enough processor cores available. A minimum of four parallel processes is assumed, so if fewer cores than this are available, performance will suffer greatly. A few experiments indicated that around six parallel processes can offer a significant improvement, but that increasing the number further may not be helpful.

It is not easy to determine what the optimum number of processes is, indeed it probably depends on the data fed in. Many runs are simply discarded along the way, so in a sense it is rather wasteful of computer resources. Also, the more parallel runs there are, the more are discarded. This may be the main factor leading to loss of performance when larger numbers of processes are run in parallel, perhaps connected with the memory accesses involved – although the multiple cores may be (more or less) independent of each other, they all access the same memory.

At present the program allows at least four parallel processes, even if fewer real processor cores are available, and a maximum of up to six, depending on the number of available cores. Thus a CPU with at least six cores is recommended.

### "Special" generation runs

In addition to the runs described above, which test individual constraint types, there are three additional ones. One configuration is run in which all constraints are disabled. Only when this completes successfully is the main part of the algorithm (as described above) started. If this doesn't complete successfully, there is no point in trying anything else. Should this happen there will normally be an error report from FET, which is recorded in the log. No result can be generated in this case, because no run completed successfully.

Then there is a fully constrained run. This covers cases in which the original data can be processed successfully within the given time, making the rest of the algorithm redundant.

Finally there is a run with all the hard constraints and no soft constraints enabled. If there are no soft constraints anyway, it will just be a second fully constrained run. Otherwise, it may complete before the step-wise accumulation of the constraint types finishes (with the hard constraints), allowing that part of the algorithm to be aborted, jumping straight to the start of the soft-constraint accumulation. This can be useful in cases where the hard constraints can all be satisfied within a convenient time, allowing soft-constraint testing to start earlier.

## Progress monitoring and completion

Apart from the timeouts of the individual runs, there is also an overall timeout, which determines the maximum length of a `fetrunner` run. Around five minutes seems reasonable in many cases and this is set as the default. But there are situations in which a shorter or longer run may be desirable, so it is possible to specify any number of seconds. If the timeout is set to 0, there is no timeout, so the run could possibly go on indefinitely. In such a case it may be necessary to terminate the run manually.

When the algorithm finishes, the results are available as a JSON object detailing the constraints which were deactivated and the activity placements resulting from the last successful run (the latest "basis").

While the algorithm is running it produces a log, documenting the progress. This can be useful for diagnosing program errors, but may also be used for run-time feedback of the progress.

## Structure of the JSON result

The JSON object presents information about the last successful "basis", which constraints were disabled and the placement of the activities. For individual constraints which led to FET errors the FET message is recorded. The information is presented in terms of item indexes. That is, each presented item (day, hour, activity, constraint, room) is referred to by an index (starting at 0). In order to correlate these items with those in the FET configuration, there is a list of references for each of these item types. These references are objects containing two entries:

**The "Id" field** is a reference to the corresponding FET node. For an Activity that is the Activity's "Id" field. For a Room, a Day or an Hour that is the item's "Name" field. In FET constraints have no clear identifying tag, so this is handled by using the "Comments" field. A new FET file is built in which the constraint comments are prefixed by some number followed by ")".  Thus each constraint can get its own identifier tag.

**There is also a "Ref" field**. This is used when the original FET file is generated by some other software. This field then references the object in that other software. In order to make these references available to `fetrunner`, the "Comments" fields of the various FET nodes are used. In the case of constraint nodes these references themselves consist of various elements:

 - constraint-name (as used in the source, not the FET name)
 - source-reference
 - parameters (comma separated)

These are combined into a single string thus:

```
    constraint-name.source-reference:parameters
```

The interpretation of the parameters is dependent on the constraint-name. Not every source constraint need have its own reference, some may be attached to a teacher or class, for example. In this case the item referred to would be passed in the parameter field, the source-reference being empty. A constraint on a group of activities might place these activities as a comma-separated list in the parameter field.

Bear in mind that the constraints list consists of the FET constraints, which may well not correspond 1:1 with the source constraints. Multiple FET constraints might derive from a single source constraint. The reverse case is currently not supported.

### Reporting "impossible" constraints

If any individual constraint was found to cause a FET error report, these reports are recorded in the "ConstraintErrors" map, the key being the constraint index, the value the message. Also constraints whose corresponding run was determined to get stuck at the beginning are reported here.
