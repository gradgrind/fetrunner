# fetrunner
    – an assistant for the generation of school timetables

## History

`fetrunner` stems from a series of converter tools allowing certain commercial timetable programs to use [`FET`](https://lalescu.ro/liviu/fet/) for automatic timetable generation. This was desirable because of limitations of the commercial programs' internal generators.

This worked pretty well. However, after several years of practical experience (generating timetables for a school), it became clear that quite a lot of insight into the generation process was necessary to get the best out of `FET`. That meant many runs with tweaked conditions (constraints) until a usable timetable was produced. It was sometimes quite an effort to trace the conditions which prevented the successful generation of a timetable meeting the specifications.

The question arose as to whether this "debugging" process could be automated to some extent. It seemed likely as the steps one would typically perform were usually very similar from one case to another, involving systematic disabling and re-enabling of various groups of constraints. `fetrunner` is the result of efforts to answer this question.

## The problem to solve

When trying to produce a timetable with `FET` it can happen that the generation step doesn't run smoothly. In some cases it won't even start, returning an error message which may help in finding the unsatisfiable condition in the configuration file ("xxx.fet"), though sometimes the message may seem a bit cryptic. In other cases the process will start, but get stuck at an early stage. In other cases the progress will be very slow, so that it is not clear when (or if!) the process will be completed.

`fetrunner` is an attempt to assist in finding the conditions which lead to these problems. It does this by repeatedly running the command-line version of `FET`, each time with a different subset of the constraints specified in the configuration file.

Somewhat as a side-product `fetrunner` also produces a fully populated timetable from its last successful configuration. That means that probably not all the constraints are satisfied, so it might not be all that useful, but under certain circumstances it might be helpful. Only if even the unconstrained configuration fails to run will there be no such placement data.

## How it works

Basically it starts by deactivating all the constraints. If even that doesn't help, there may be an error message from `FET` which can offer some useful information.

Then it activates a single type of constraint, which means all the individual constraints of a particular type, and starts a `FET` generation run. This is repeated, in principle, for all the constraint types used in the source file. Several of these processes can be run simultaneously (see [Parallel processing](#parallel-processing)).

When one of these generation runs completes successfully, all the others are stopped and discarded. This first successful run is now taken as a new basis. For each of the remaining constraint types a new generation run is started, with the new constraint type being added to the successfully completed one. Again, the first successfully completing run is taken as a new basis, and so on until all constraint types have been enabled.

Well, given an "easy" configuration file that might work within a reasonable time. But there would not be much point, because we might as well just run the original configuration file with all constraints enabled, which would be a lot simpler and quicker. `fetrunner` becomes interesting when one or more constraints are causing errors or holding up the generation run in some other way.

Normally the first stages of the algorithm (the addition of the "easiest" constraint types) will run quickly, within a few seconds, allowing most of the constraints to be enabled in a fairly short time. But the longer the run times, the more difficult it is to cover all the remaining constraint types in a given overall run time – some "easy" constraints might never get tested.

There are essentially two cases that need dealing with. Perhaps the simplest is when a constraint blocks the generation completely, in such a way that `FET` returns an error message. The other is when the generation run gets stuck or progresses only very slowly (it is not always easy to distinguish between these two). At first, while there are still constraint types which can be added whose runs complete quickly, non-completing runs are no problem because they will simply be discarded. But at some point, only the difficult constraint types remain.

One of the aims of `fetrunner` is to produce a timetable within a limited, specified time. When the configuration is "challenging", this may only be possible by leaving some of the constraints disabled, even though they might be satisfiable (given enough time).

### Hard and soft constraints

A constraint is categorised as "hard" if it *must* be satisfied in an acceptable timetable. `fetrunner` loosens this requirement a bit in order to get all the activities placed in a given time. The resulting timetable is then presumably not yet acceptable, but may be helpful. A "soft" constraint is one which should be satisfied "if possible", but a resulting timetable may be acceptable even if the constraint is not satisfied. In `FET`, hard constraints have "weight" 100 (%) – for any smaller value the constraint is considered soft (even if, say, 99.99% is not particularly soft).

`fetrunner` takes the distinction between hard and soft constraints into account in that it doesn't touch the soft constraints (it leaves them deactivated) until all the hard constraints have been satisfied (or rejected as "impossible"). Thus, it is possible that the soft constraints never get tested, if the processing of the hard constraints takes up all the available time.

Because of the way `fetrunner` works, it doesn't make much sense to actually use soft constraints in their original form. In `FET` a soft constraint is rejected if a number (dependent on the weight) of placement attempts conflict with it. `fetrunner` does something vaguely similar, in that it tries to use the constraint, only accepting it if a run completes successfully. In `fetrunner`, the weight of the soft constraints is used to determine the initial order in which the constraints are enabled for testing, so that ones with a higher weight are introduced first. The constraint itself is run in `FET` as a hard constraint (weight 100). As this approach has not yet been widely tested, there is a flag which can be set to cause the original weight to be used, so that the effectiveness can be compared.

### Handling "difficult" constraints

Many constraints will be satisfied easily, and `fetrunner` tries to enable these first, so that as many constraints as possible get enabled in the given time. As soon as time-consuming ("difficult") constraints are enabled, each individual test run takes longer, so fewer tests are possible. Some constraint types may be more prone to being "difficult" than others, but in general this depends on the input data.

The algorithm uses various tests to determine whether a set of constraints (of one type) is "difficult". These have been developed more or less by trial and error, and may not be at all optimal, if indeed there is such a thing in this situation. The primary considerations are whether a run is likely to complete, and how long it will take. A `FET` run produces a completion measure – how many activities have been placed. In combination with the elapsed time, `fetrunner` uses this to calculate a rate of progress.

If the rate of progress is too low, the run may be terminated and the list of constraints split into two halves, which are then (later, when processing capacity becomes available) run separately. In this way a binary search procedure for potentially blocking constraints is implemented. If there is such a constraint, this process will lead eventually to a run with a single new constraint. If this run is deemed to be "stuck", the constraint is marked as "impossible" and removed from consideration.

In some cases a constraint really will be impossible, at least in combination with others, and there may be a more or less helpful message from `FET`. In other cases it may just be that more time would be needed to incorporate this constraint. It is in the nature of `fetrunner` that it cannot always distinguish between the two (because of the general complexity of the timetabling problem).

The acceptable rate of progress varies according to the overall state of the process. At the beginning only rapidly progressing runs will be accepted, so that as many "easy" constraints as possible get accepted. As more constraints are enable, the run time will generally increase, so lower rates of progress must be tolerated. The acceptable rate is made dependent on the completion time for the last successful run.

If at some stage of the overall proceedings the available "processes" (threads, processor cores) are not all used, a set of constraints may be split even before it is judged to be too slow.

### Parallel processing

The number of timetable-generation runs which are started at any time depends on the number of processor cores available. Experiments indicate, however, that there may be something like an optimal number, too many processes perhaps even reducing the overall performance. To keep the main part of the algorithm separate from the handling of parallel processes, a process queue is used. Any number of generation runs may be added to the queue, but a run will only be actually started if one of the limited number of CPU processes is available.

As the algorithm depends on multiple processes running simultaneously, the effectiveness depends on there being enough processor cores available. A minimum of four parallel processes is assumed, so if fewer cores than this are available, performance will suffer greatly. A few experiments indicated that around six parallel processes can offer a significant improvement, but that increasing the number further may not be helpful.

It is not easy to determine what the optimum number of parallel processes is, indeed it probably depends on the data fed in. Many runs are simply discarded along the way, so in a sense it is rather wasteful of computer resources. Also, the more parallel runs there are, the more are discarded. This may be the main factor leading to loss of performance when larger numbers of processes are run in parallel, perhaps connected with the memory accesses involved – although the multiple cores may be (more or less) independent of each other, they all access the same memory.

At present the program allows at least four parallel processes, even if fewer real processor cores are available, and a maximum of up to six, depending on the number of available cores. Thus a CPU with at least six cores is recommended.

### "Special" generation runs

In addition to the timetable-generation runs described above, which test individual constraint types, there are three additional ones. One configuration is run in which all constraints are disabled. Only when this completes successfully is the main part of the algorithm (as described above) started. If this doesn't complete successfully, there is no point in trying to add constraints. Should this happen there will normally be an error report from FET, which is recorded in the log. No result can be generated in this case, because no run completed successfully.

Then there is a fully constrained run. This covers cases in which the original data can be processed successfully within the given time, making the rest of the algorithm redundant.

Finally there is a run with all the hard constraints and no soft constraints enabled. If there are no soft constraints anyway, it will just be a second fully constrained run. Otherwise, it may complete before the step-wise accumulation of the constraint types finishes (with the hard constraints), allowing that part of the algorithm to be aborted, jumping straight to the start of the soft-constraint accumulation. This can be useful in cases where the hard constraints can all be satisfied within a convenient time, allowing soft-constraint testing to start earlier.

## Progress monitoring and completion

Apart from the timeouts of the individual runs, there is also an overall timeout, which determines the maximum length of a `fetrunner` run. Around five minutes seems reasonable in many cases and this is set as the default. But there are situations in which a shorter or longer run may be desirable, so it is possible to specify any number of seconds. If the timeout is set to 0, there is no timeout, so the run could possibly go on indefinitely. In such a case it may be necessary to terminate the run manually.

When the algorithm finishes, the results are available as a [JSON object](Result_JSON.md) detailing the constraints which were deactivated and the activity placements resulting from the last successful run (the latest "basis").

While the algorithm is running it produces a log, documenting the progress. This can be useful for diagnosing program errors, but may also be used for run-time feedback of the progress.
