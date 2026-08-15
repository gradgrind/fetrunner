# How `fetrunner` works

The basic idea is to behave similarly to a person looking for possible problems in a `FET` file. The constraints are divided into types and hard/soft collections. By progressively enabling the groups of constraints belonging to each type (in a certain order, set in the program), the program attempts to include as many as possible of the specified constraints consistent with `FET` delivering a solution within a reasonable time.

The program has strategies for deciding whether a run has got stuck or is progressing too slowly. When an instance fails or is determined to be too slow, it is aborted. If the list of new constraints it is testing has more than one entry, two new instances will be generated, each with half of the original constraints. In this way, individual constraints which are difficult can be isolated and eliminated. As more and more constraints are enabled, it is to be expected that the time to reach a solution will increase. The time-out strategies try to take this into account.

Initially, three or four "special" instances are started:

 - fully constrained
 - all the hard constraints, but no soft constraints
 - "priority" constraints only (hard, non-available time slots and fixed activity placements)
 - unconstrained

These instances are not "splittable". Under normal circumstances these instances run until they complete naturally. If there are no priority constraints, this instance will not be run. If there are no hard constraints, the second instance will be a duplicate of the first (but will still run).

When a non-special instance completes successfully, all other current non-special instances are aborted. The successful instance is then used as a new basis for regenerating the next instances to be tested, one for each of the remaining constraint types relevant for this phase. The successful completion of a special instance is handled separately.

The overall operation is divided into "phases". On entering the first phase (phase 0), these special instances are already running. Instances for each of the priority-constraint types are generated, so that testing starts with just these. New instances are added to a queue, so that they can be run when processor cores become available.

Normally, the unconstrained instance will complete very quickly, so that there is at least a very minimal successful run available as a result. If even the unconstrained instance fails or is aborted (timed out), there will be no result at all.

The first phase ends when either the priority-constraints instance completes successfully or when as many as possible of the priority constraints have been included in a successful `FET` run.

In the second phase (phase 1), all the remaining hard constraints are tested. On entry, all running (and queued) instances except the specials "fully constrained" and "all hard constraints" will be cancelled. A new set of instances, one for each of the remaining hard constraint types, is generated based on the last successful instance from the first phase. These are added to the queue to be run when processor cores become available.

The program tries to favour constraint types which are easily satisfied, by testing them earlier (it has an internal ordering for the various types). In this way it should be possible to include as many constraints as possible within a limited time. However, there is no sure way of knowing which types are going to be easier, so it's little more than guess-work.

The second phase ends when either the "all hard constraints" instance completes successfully or when as many as possible of the hard constraints have been included in a successful `FET` run.

In the third phase (phase 2), all the soft constraints are tested. On entry, all running (and queued) instances except the special "fully constrained" will be cancelled. A new set of instances, one for each of the soft constraint types, is generated based on the last successful instance from the second phase. These are added to the queue to be run when processor cores become available.

The order in which the soft constraint types are added depends on their weight, those with the largest weight being tested first. It is worth noting that, by default, `fetrunner` runs the soft constraints in `FET` as if they were hard. This can be seen as an alternative way of dealing with soft constraints which is more in tune with the way `fetrunner` works. Indeed, even the hard constraints are treated by `fetrunner` as if they were soft, by eliminating them if they seem to be too "difficult". Thus the main difference between hard and soft constraints is that the soft constraints are only tested when testing of the hard constraints is finished.

The third phase ends when either the "fully constrained" instance completes successfully or when as many as possible of the soft constraints have been included in a successful `FET` run. At this point the tidy-up phase (phase 3) is entered, which just waits for all remaining active (but now cancelled) instances to finish cleanly. Then the final result can be constructed.

With each successful run contributing its set of constraints to the new runs, there is always a "best so far" `FET` instance, which gradually gathers more and more constraints (ideally!).

When the time limit for the program is reached, or when it is manually interrupted, the latest successful run is taken as the final result.

There are some difficult cases with which `fetrunner` can't help much, because the basic runs take too long, but for many `FET` files it can provide some useful information.

