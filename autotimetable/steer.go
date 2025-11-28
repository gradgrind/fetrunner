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
	"strings"
	"syscall"
	"time"
)

func (attdata *AutoTtData) SetParameterDefault() {
	/* There is probably no general "optimum" value for the various
	parameters, that is likely to depend on the data. But perhaps values
	can be found which are frequently useful. It might be helpful to use
	shorter overall timeouts during the initial cycles of testing the data,
	to identify potential problem areas without long processing delays. For
	later cycles longer times may be necessary (depending on the difficulty
	of the data).
	*/
	attdata.Parameters.MAXPROCESSES = min(max(runtime.NumCPU(), 4), 6)

	attdata.Parameters.NEW_BASE_TIMEOUT_FACTOR = 12  // => 1.2
	attdata.Parameters.NEW_CYCLE_TIMEOUT_FACTOR = 15 // => 1.5
	attdata.Parameters.LAST_TIME_0 = 5
	attdata.Parameters.LAST_TIME_1 = 50

	attdata.Parameters.DEBUG = false
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
are tried again, but this time with longer timeouts. Note that instances
which are trying to add just one constraint are started without a timeout

When all the hard constraints have been included, the soft constraints are
added in basically the same way. If the initial instance with all hard
constraints and no soft constraints completes, this instance will be used as
the new base for adding the soft constraints and the accumulation loop will
be cancelled. If the accumulation loop should finish first (somewhat unlikely,
but possible), the initial instance with all hard constraints may be
terminated.

When everything has been added that can be in the given time, the latest
`Result` is saved as a JSON file, "Result.json". This includes details of the
last successful run and diagnostic information – at least an indication of
which constraints were dropped and any error messages for them which may have
been produced by the generator back-end).
*/

func (attdata *AutoTtData) StartGeneration(TIMEOUT int) {
	attdata.Running = true
	logger := attdata.BaseData.Logger
	attdata.lastResult = nil
	attdata.ConstraintErrors = map[ConstraintIndex]string{}
	attdata.BlockConstraint = map[ConstraintIndex]bool{}
	attdata.instanceCounter = 0
	attdata.current_instance = nil

	// Catch termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	runqueue := &RunQueue{
		AutoTtData: attdata,
		Queue:      nil,
		Active:     map[*TtInstance]struct{}{},
		MaxRunning: attdata.Parameters.MAXPROCESSES,
		Next:       0,
	}

	// Global data
	attdata.Ticks = 0

	// First run: all constraints enabled.
	// On successful completion, all other instances should be stopped.
	// If it fails, just this instance should be wound up. Otherwise it
	// should run until it times out, at which point any other active
	// instances should be stopped and the "best" solution at this point
	// chosen.
	enabled := make([]bool, attdata.NConstraints)
	for i := range attdata.NConstraints {
		enabled[i] = true
	}
	attdata.full_instance = &TtInstance{
		Tag:               "COMPLETE",
		Timeout:           0,
		ConstraintEnabled: enabled,
	}
	// Add to run queue
	runqueue.add(attdata.full_instance)

	// Instance without soft constraints (if any, otherwise same as full
	// instance) – enable only the hard constraints.
	enabled = make([]bool, attdata.NConstraints)
	for _, ilist := range attdata.HardConstraintMap {
		for _, i := range ilist {
			enabled[i] = true
		}
	}
	attdata.hard_instance = &TtInstance{
		Tag:               "HARD_ONLY",
		Timeout:           0,
		ConstraintEnabled: enabled,
	}
	// Add to run queue
	runqueue.add(attdata.hard_instance)

	// Unconstrained instance
	attdata.cycle_timeout = 0
	if attdata.Parameters.SKIP_HARD {
		attdata.phase = 2
		if len(attdata.SoftConstraintMap) == 0 {
			logger.Warning("-h- Option -h: no soft constraints")
		}
	} else {
		enabled = make([]bool, attdata.NConstraints)
		attdata.null_instance = &TtInstance{
			Tag:               "UNCONSTRAINED",
			Timeout:           attdata.cycle_timeout,
			ConstraintEnabled: enabled,
		}
		// Add to run queue
		runqueue.add(attdata.null_instance)

		// Start phase 0
		logger.Info(
			"[%d] Phase 0 ...\n",
			attdata.Ticks)
		attdata.phase = 0
	}

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
			logger.Bug("[%d] !!! RECOVER !!!\n=== %v\n+++\n%s\n---\n",
				attdata.Ticks, r, debug.Stack())
			base.Report("!!! ERROR: see log\n")
		}
		for {
			// Wait for active instances to finish, stopping them if
			// necessary.
			count := 0
			for instance := range runqueue.Active {
				if instance.RunState == 0 {
					instance.Backend.Tick(attdata, instance)
					count++
					attdata.abort_instance(instance)
				}
			}
			if count == 0 {
				break
			}
			<-ticker.C
		}
		if !attdata.Parameters.DEBUG {
			// Remove all remaining temporary files
			attdata.BackendInterface.Tidy()
		}
		if attdata.lastResult != nil {

			attdata.lastResult.ConstraintErrors = attdata.ConstraintErrors

			// Save result of last successful instance.
			//b, err := json.Marshal(LastResult)
			b, err := json.MarshalIndent(attdata.lastResult, "", "  ")
			if err != nil {
				panic(err)
			}
			fpath := filepath.Join(attdata.BaseData.SourceDir, "Result.json")
			f, err := os.Create(fpath)
			if err != nil {
				panic("Couldn't open output file: " + fpath)
			}
			defer f.Close()
			_, err = f.Write(b)
			if err != nil {
				panic("Couldn't write result to: " + fpath)
			}
			attdata.current_instance.Backend.FinalizeResult(attdata)
		}
	}()

tickloop:
	// Start queued instances if there are free processors.
	for {
		if runqueue.update_queue() == 0 {
			logger.Info(
				"[%d] Run-queue empty\n",
				attdata.Ticks)
		}

		select {
		case <-ticker.C:
		case <-sigChan:
			logger.Info("[%d] !!! INTERRUPTED !!!\n",
				attdata.Ticks)
			break tickloop
		}

		attdata.Ticks++
		logger.Info(
			"[%d] +TICK\n",
			attdata.Ticks)

		runqueue.update_instances()

		if attdata.full_instance.ProcessingState == 1 {
			// Cancel all other runs and return this instance as result.
			attdata.current_instance = attdata.full_instance
			attdata.new_current_instance(attdata.current_instance)
			logger.Info("[%d] +A+ All constraints OK +++\n",
				attdata.Ticks)
			break
		} else {
			p := attdata.full_instance.Progress
			if p > full_progress {
				full_progress = p
				full_progress_ticks = attdata.Ticks
				logger.Info(
					"[%d] ? %s (%d @ %d)\n",
					attdata.Ticks,
					attdata.full_instance.Tag,
					full_progress,
					full_progress_ticks,
				)
			}
		}

		if attdata.phase != 2 {
			if attdata.hard_instance.ProcessingState == 1 {
				// Set as current and start processing soft constraints.
				attdata.current_instance = attdata.hard_instance
				attdata.new_current_instance(attdata.current_instance)
				logger.Info(
					"[%d] +H+ All hard constraints OK +++\n",
					attdata.Ticks)
				// Cancel everything except full instance.
				if attdata.null_instance.ProcessingState == 0 {
					attdata.abort_instance(attdata.null_instance)
				}
				for _, instance := range attdata.constraint_list {
					if instance.ProcessingState == 0 {
						attdata.abort_instance(instance)
					}
					// Indicate that a queued instance is not to be started
					instance.ProcessingState = 3
				}
				attdata.constraint_list = nil
				attdata.phase = 1
			} else {
				p := attdata.hard_instance.Progress
				if p > hard_progress {
					hard_progress = p
					hard_progress_ticks = attdata.Ticks
					logger.Info(
						"[%d] ? %s (%d @ %d)\n",
						attdata.Ticks,
						attdata.hard_instance.Tag,
						hard_progress,
						hard_progress_ticks,
					)
				}
			}
		} else if attdata.Parameters.SKIP_HARD {
			if attdata.current_instance == nil {
				if attdata.hard_instance.ProcessingState == 1 {
					// First successful instance.
					attdata.current_instance = attdata.hard_instance
				}
			} else if attdata.hard_instance.ProcessingState == 0 {
				attdata.abort_instance(attdata.hard_instance)
			}

			if attdata.hard_instance.ProcessingState == 1 &&
				attdata.current_instance == nil {
				// First successful instance.
				attdata.current_instance = attdata.hard_instance
			}
		}

		if attdata.Ticks == TIMEOUT {
			logger.Info(
				"[%d] !!! TIMEOUT !!!\n + %s: %d @ %d\n + %s: %d @ %d\n",
				attdata.Ticks,
				attdata.full_instance.Tag,
				full_progress,
				full_progress_ticks,
				attdata.hard_instance.Tag,
				hard_progress,
				hard_progress_ticks,
			)
			break
		}

		if attdata.phase == 0 {
			// During phase 0 only `full_instance`, `hard_instance` and
			// `null_instance` are running.
			switch runqueue.phase0() {
			case 0:
				continue

			case 1:
				//TODO-- attdata.phase = 1
				//TODO-- base.Message.Printf(
				//TODO-- 	"[%d] Phase 1 ...\n",
				//TODO-- 	attdata.Ticks)

			case -1:
				logger.Info(
					"[%d] Couldn't process input data!\n",
					attdata.Ticks)
				base.Report("!!! Couldn't process input data!\n")
				return

			default:
				panic("attdata.phase0() -> invalid return value")
			}
		}

		if runqueue.mainphase() {
			break
		}
	} // tickloop: end
	logger.Info(
		"[%d] Phase 3 ... finalizing ...\n",
		attdata.Ticks)

	hnn := 0
	hnall := 0
	snn := 0
	snall := 0
	result := attdata.current_instance
	type constraintinfo struct {
		c string
		n int
		N int
	}
	infolist := []constraintinfo{}
	for c, clist := range attdata.HardConstraintMap {
		n := 0
		for _, cix := range clist {
			if result != nil && result.ConstraintEnabled[cix] {
				n++
			}
		}
		if len(clist) != 0 {
			infolist = append(infolist,
				constraintinfo{string(c) + " (HARD)", n, len(clist)})
			hnn += n
			hnall += len(clist)
		}
	}

	for c, clist := range attdata.SoftConstraintMap {
		n := 0
		for _, cix := range clist {
			if result != nil && result.ConstraintEnabled[cix] {
				n++
			}
		}
		if len(clist) != 0 {
			infolist = append(infolist,
				constraintinfo{string(c) + " (SOFT)", n, len(clist)})
			snn += n
			snall += len(clist)
		}
	}
	slices.SortFunc(infolist, func(a, b constraintinfo) int {
		return strings.Compare(a.c, b.c)
	})
	for _, info := range infolist {
		logger.Info("$ %s: %d / %d\n", info.c, info.n, info.N)
	}

	report := fmt.Sprintf(
		"::: ALL CONSTRAINTS: (hard) %d / %d  (soft) %d / %d\n",
		hnn, hnall, snn, snall)
	logger.Info("%s", report)
	base.Report(report)

	if result != nil {
		logger.Info("Result: %s\n", result.Tag)
	}
	logger.Result("TT_DONE", "")
	attdata.Running = false
}

func (attdata *AutoTtData) abort_instance(instance *TtInstance) {
	if !instance.Stopped {
		instance.Backend.Abort()
		instance.Stopped = true
	}
}

func (attdata *AutoTtData) new_instance(
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
	// Single-constraint instances always have no timeout
	if len(constraint_indexes) == 1 {
		timeout = 0
	}

	// Make a new `TtInstance`
	attdata.instanceCounter++
	instance := &TtInstance{
		Tag: fmt.Sprintf("z%05d~%s",
			attdata.instanceCounter, constraint_type),
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
