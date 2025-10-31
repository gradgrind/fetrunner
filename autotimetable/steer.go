package autotimetable

import (
	"encoding/json"
	"fetrunner/base"
	"fmt"
	"os"
	"os/signal"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"slices"
	"syscall"
	"time"
)

func (bdata *BasicData) SetParameterDefault() {
	/* There is probably no general "optimum" value for the various
	parameters, that is likely to depend on the data. But perhaps values
	can be found which are frequently useful. It might be helpful to use
	shorter overall timeouts during the initial cycles of testing the data,
	to identify potential problem areas without long processing delays. For
	later cycles longer times may be necessary (depending on the difficulty
	of the data).
	*/
	bdata.Parameters.MAXPROCESSES = min(max(runtime.NumCPU(), 4), 6)

	bdata.Parameters.NEW_BASE_TIMEOUT_FACTOR = 12  // => 1.2
	bdata.Parameters.NEW_CYCLE_TIMEOUT_FACTOR = 15 // => 1.5
	bdata.Parameters.LAST_TIME_0 = 5
	bdata.Parameters.LAST_TIME_1 = 50

	bdata.Parameters.DEBUG = false
}

/*
One aim is to achieve a – possibly imperfect – timetable within a specified
time. As it is impossible to guarantee that all constraints will be satisfied
within a given time, it may be necessary to drop some of them in order to
place all the activities within this time. However, if it is possible
to satisfy all the constraints within this time, that should be done.

If constraints need to be dropped, these should give an indication as to which
ones are difficult. Perhaps more time would help, or a modification of some of
the constraints. Among the difficult ones there may also be constraints which
block the completion of the task and thus must be changed or dropped.

A certain degree of parallel processing is assumed – too few (less than four?)
processor cores is likely to result in a very significant slowdown.

The main function (`StartGeneration`) starts a run with the fully constrained
data and a second run with all the "soft" (non-compulsory) constraints removed.
A further instance is run in which basically all constraints are removed. If
this latter fails, then there is a serious problem with the activities, which
don't fit into the school week. As this unconstrained instance should complete
quickly in most real-life situations, it is given a short timeout. In theory
it is possible that the generation of a timetable could take a longer time
even with the unconstraind data, but such a case would need to be handled in a
different way (and is, in general, a difficult problem ...). This program
assumes that the unconstrained data will allow the rapid placement of all the
activities.

A `TtInstance` structure is constructed to manage the data for each
timetable generation run, each run having its own goroutine. Each instance
has its own individual timeout. There is also a global timeout to stop
all instances which are still running.

Once these initial instances have been started, a "tick-loop" (which is
triggered every second) is entered. This monitors the progress of each active
instance and handles the actions resulting from their completion, whether
successful or not.

Should the fully constrained instance complete successfully within the
allotted time, all other instances are terminated and its result will be
saved.

When the unconstrained instance completes successfully, a series of further
instances is queued for running, each specifying the addition of a list of
(hard) constraints of a single type. Thus for each type of constraint an
instance is constructed. Using timeouts leading to binary divisions of these
lists an attempt is made to find individual "difficult" constraints, which can
then be disabled in order to get full activity placement within a reasonable
time. Parallel processing can be of some assistance here.

As the constraint types are added one after the other, and often each step
will take longer than the previous one (as the number of constraints grows),
it should be clear why it is desirable that at least the early stages are
completed quickly.

When a single-constraint-type instance completes successfully, it is used as
a new base (`current_instance`) for the addition of further constraints. All
the remaining constraint-type instances are stopped and restarted with this
new base. If a constraint-type instance is timed out, it is stopped and split
into two halves, which then run in its place. If there are no halves (only
one constraint being added) there is no successor, the constraint is dropped
(from this accumulation round).

When an instance completes successfully within the allotted time, its result
is saved as a `Result` structure, the best result so far gradually
encompassing more of the constraints. When all the constraints have been
tested with a certain timeout, a new round is entered and the rejected ones
are tried again, but this time with longer timeouts.

When all the hard constraints have been included, the soft constraints are
added in basically the same way. If the initial instance with all hard
constraints and no soft constraints completes, this instance will be used as
the new base for adding the soft constraints and the accumulation loop will
be cancelled. If the accumulation loop should finish first (somewhat unlikely,
but possible), the initial instance with all hard constraints may be
terminated.

When everything has been added that can be in the given time, the latest
`Result` is saved as a JSON file, "Result.json". This includes diagnostic
information (at least an indication of which constraints were dropped), and
details of the last successful run.
*/

func (basic_data *BasicData) StartGeneration(TIMEOUT int) {
	basic_data.lastResult = nil
	basic_data.ConstraintErrors = map[ConstraintIndex]string{}

	// Catch termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	runqueue := &RunQueue{
		BasicData:  basic_data,
		Queue:      nil,
		Active:     map[*TtInstance]struct{}{},
		MaxRunning: basic_data.Parameters.MAXPROCESSES,
		Next:       0,
	}

	// Global data
	basic_data.Ticks = 0

	// First run: all constraints enabled.
	// On successful completion, all other instances should be stopped.
	// If it fails, just this instance should be wound up. Otherwise it
	// should run until it times out, at which point any other active
	// instances should be stopped and the "best" solution at this point
	// chosen.
	enabled := make([]bool, basic_data.NConstraints)
	for i := range basic_data.NConstraints {
		enabled[i] = true
	}
	basic_data.full_instance = &TtInstance{
		Tag:               "COMPLETE",
		Timeout:           0,
		ConstraintEnabled: enabled,
	}
	// Add to run queue
	runqueue.add(basic_data.full_instance)

	// Instance without soft constraints (if any, otherwise same as full
	// instance) – enable only the hard constraints.
	enabled = make([]bool, basic_data.NConstraints)
	for _, ilist := range basic_data.HardConstraintMap {
		for _, i := range ilist {
			enabled[i] = true
		}
	}
	basic_data.hard_instance = &TtInstance{
		Tag:               "HARD_ONLY",
		Timeout:           0,
		ConstraintEnabled: enabled,
	}
	// Add to run queue
	runqueue.add(basic_data.hard_instance)

	// Unconstrained instance
	basic_data.cycle_timeout = 0
	if basic_data.Parameters.SKIP_HARD {
		basic_data.phase = 2
	} else {
		enabled = make([]bool, basic_data.NConstraints)
		basic_data.null_instance = &TtInstance{
			Tag:               "UNCONSTRAINED",
			Timeout:           basic_data.cycle_timeout,
			ConstraintEnabled: enabled,
		}
		// Add to run queue
		runqueue.add(basic_data.null_instance)

		// Start phase 0
		base.Message.Printf(
			"[%d] Phase 0 ...\n",
			basic_data.Ticks)
		basic_data.phase = 0
	}

	basic_data.cycle = 0
	full_progress := 0       // current percentage
	full_progress_ticks := 0 // time of last increment
	hard_progress := 0       // current percentage
	hard_progress_ticks := 0 // time of last increment

	// *** Ticker loop ***
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer func() {
		// Tidy up
		r := recover()
		if r != nil {
			base.Bug.Printf("[%d] !!! RECOVER !!!\n=== %v\n+++\n%s\n---\n",
				basic_data.Ticks, r, debug.Stack())
			base.Report("!!! ERROR: see log\n")
		}
		for {
			// Wait for active instances to finish, stopping them if
			// necessary.
			count := 0
			for instance := range runqueue.Active {
				if instance.RunState == 0 {
					instance.Backend.Tick(basic_data, instance)
					count++
					basic_data.abort_instance(instance)
				}
			}
			if count == 0 {
				break
			}
			<-ticker.C
		}
		if !basic_data.Parameters.DEBUG {
			// Remove all remaining temporary files
			basic_data.BackendInterface.Tidy()
		}
		if basic_data.lastResult != nil {
			// Save result of last successful instance.
			//b, err := json.Marshal(LastResult)
			b, err := json.MarshalIndent(basic_data.lastResult, "", "  ")
			if err != nil {
				panic(err)
			}
			fpath := filepath.Join(basic_data.WorkingDir, "Result.json")
			f, err := os.Create(fpath)
			if err != nil {
				panic("Couldn't open output file: " + fpath)
			}
			defer f.Close()
			_, err = f.Write(b)
			if err != nil {
				panic("Couldn't write result to: " + fpath)
			}
			basic_data.current_instance.Backend.FinalizeResult(basic_data)
		}
	}()

tickloop:
	// Start queued instances if there are free processors.
	for {
		if runqueue.update_queue() == 0 {
			base.Message.Printf(
				"[%d] Run-queue empty\n",
				basic_data.Ticks)
		}

		//TODO
		base.Message.Printf(
			"[%d] -TICK\n",
			basic_data.Ticks)

		select {
		case <-ticker.C:
		case <-sigChan:
			base.Message.Printf("[%d] !!! INTERRUPTED !!!\n",
				basic_data.Ticks)
			break tickloop
		}

		basic_data.Ticks++
		//TODO
		base.Message.Printf(
			"[%d] +TICK\n",
			basic_data.Ticks)

		runqueue.update_instances()

		if basic_data.full_instance.ProcessingState == 1 {
			// Cancel all other runs and return this instance as result.
			basic_data.current_instance = basic_data.full_instance
			basic_data.new_current_instance(basic_data.current_instance)
			base.Message.Printf("[%d] +A+ All constraints OK +++\n",
				basic_data.Ticks)
			break
		} else {
			p := basic_data.full_instance.Progress
			if p > full_progress {
				full_progress = p
				full_progress_ticks = basic_data.Ticks
				base.Message.Printf(
					"[%d] ? %s (%d @ %d)\n",
					basic_data.Ticks,
					basic_data.full_instance.Tag,
					full_progress,
					full_progress_ticks,
				)
			}
		}

		if basic_data.phase != 2 {
			if basic_data.hard_instance.ProcessingState == 1 {
				// Set as current and start processing soft constraints.
				basic_data.current_instance = basic_data.hard_instance
				basic_data.new_current_instance(basic_data.current_instance)
				base.Message.Printf(
					"[%d] +H+ All hard constraints OK +++\n",
					basic_data.Ticks)
				// Cancel everything except full instance.
				if basic_data.null_instance.ProcessingState == 0 {
					basic_data.abort_instance(basic_data.null_instance)
				}
				for _, instance := range basic_data.constraint_list {
					if instance.ProcessingState == 0 {
						basic_data.abort_instance(instance)
					}
					// Indicate that a queued instance is not to be started
					instance.ProcessingState = 3
				}
				basic_data.constraint_list = nil
				basic_data.phase = 2
				base.Message.Printf(
					"[%d] Phase 2 <- %s\n",
					basic_data.Ticks, basic_data.hard_instance.Tag)
			} else {
				p := basic_data.hard_instance.Progress
				if p > hard_progress {
					hard_progress = p
					hard_progress_ticks = basic_data.Ticks
					base.Message.Printf(
						"[%d] ? %s (%d @ %d)\n",
						basic_data.Ticks,
						basic_data.hard_instance.Tag,
						hard_progress,
						hard_progress_ticks,
					)
				}
			}
		} else if basic_data.Parameters.SKIP_HARD {
			if basic_data.current_instance == nil {
				if basic_data.hard_instance.ProcessingState == 1 {
					// First successful instance.
					basic_data.current_instance = basic_data.hard_instance
				}
			} else if basic_data.hard_instance.ProcessingState == 0 {
				basic_data.abort_instance(basic_data.hard_instance)
			}

			if basic_data.hard_instance.ProcessingState == 1 &&
				basic_data.current_instance == nil {
				// First successful instance.
				basic_data.current_instance = basic_data.hard_instance
			}
		}

		if basic_data.Ticks == TIMEOUT {
			base.Message.Printf(
				"[%d] !!! TIMEOUT !!!\n + %s: %d @ %d\n + %s: %d @ %d\n",
				basic_data.Ticks,
				basic_data.full_instance.Tag,
				full_progress,
				full_progress_ticks,
				basic_data.hard_instance.Tag,
				hard_progress,
				hard_progress_ticks,
			)
			break
		}

		if basic_data.phase == 0 {
			// During phase 0 only `full_instance`, `hard_instance` and
			// `null_instance` are running.
			switch runqueue.phase0() {
			case 0:
				continue

			case 1:
				basic_data.phase = 1
				base.Message.Printf(
					"[%d] Phase 1 ...\n",
					basic_data.Ticks)

			case -1:
				base.Error.Printf(
					"[%d] Couldn't process input data!\n",
					basic_data.Ticks)
				base.Report("!!! Couldn't process input data!\n")
				return

			default:
				panic("basic_data.phase0() -> invalid return value")
			}
		}

		// There should be no problem if there are no constraints to add.

		if runqueue.mainphase() {
			break
		}
	} // tickloop: end
	base.Message.Printf(
		"[%d] Phase 3 ... finalizing ...\n",
		basic_data.Ticks)

	hnn := 0
	hnall := 0
	snn := 0
	snall := 0
	result := basic_data.current_instance
	for c, clist := range basic_data.HardConstraintMap {
		n := 0
		for _, cix := range clist {
			if result != nil && result.ConstraintEnabled[cix] {
				n++
			}
		}
		if len(clist) != 0 {
			base.Message.Printf("$ (HARD) %s: %d / %d\n", c, n, len(clist))
			hnn += n
			hnall += len(clist)
		}
	}
	for c, clist := range basic_data.SoftConstraintMap {
		n := 0
		for _, cix := range clist {
			if result != nil && result.ConstraintEnabled[cix] {
				n++
			}
		}
		if len(clist) != 0 {
			base.Message.Printf("$ (SOFT) %s: %d / %d\n", c, n, len(clist))
			snn += n
			snall += len(clist)
		}
	}
	report := fmt.Sprintf(
		"::: ALL CONSTRAINTS: (hard) %d / %d  (soft) %d / %d\n",
		hnn, hnall, snn, snall)
	base.Message.Print(report)
	base.Report(report)

	if result != nil {
		base.Message.Printf("Result: %s\n", result.Tag)
	}
}

func (basic_data *BasicData) abort_instance(instance *TtInstance) {
	if !instance.Stopped {
		instance.Backend.Abort()
		instance.Stopped = true
	}
}

func (basic_data *BasicData) new_instance(
	instance_0 *TtInstance,
	constraint_type ConstraintType,
	constraint_indexes []ConstraintIndex,
	timeout int,
) *TtInstance {
	enabled := slices.Clone(instance_0.ConstraintEnabled)
	// Add the new constraints
	for _, c := range constraint_indexes {
		enabled[c] = true
	}

	// Make a new `TtInstance`
	basic_data.instanceCounter++
	instance := &TtInstance{
		Tag: fmt.Sprintf("z%05d~%s",
			basic_data.instanceCounter, constraint_type),
		Timeout:      timeout,
		BaseInstance: instance_0,

		ConstraintEnabled: enabled,
		ConstraintType:    constraint_type,
		Constraints:       constraint_indexes,

		// Run time
		//BackEndData     any
		Ticks:   0,
		Stopped: false,
		//ProcessingState int
	}
	return instance
}
