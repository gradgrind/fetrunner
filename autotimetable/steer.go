package autotimetable

import (
	"encoding/json"
	"fetrunner/base"
	"fetrunner/timetable"
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

var (
	// The behaviour of the TESTING flag depends on the back-end. It might,
	// for example, use fixed seeds for random number generators so as to
	// produce reproduceable runs.
	TESTING bool
	// This approach relies on parallel processing. If there are too few real
	// processors it will be inefficient:
	MAXPROCESSES int

	NEW_BASE_TIMEOUT_FACTOR  int // factor * 10
	STAGE_TIMEOUT_MIN        int
	CYCLE_TIMEOUT            int
	NEW_STAGE_TIMEOUT_FACTOR int // factor * 10

	DEBUG bool

	InstanceCounter int = 0
	LastResult      *Result
	// Tick count limits for testing whether an instance with no timeout
	// has got stuck. See `(*RunQueue).update_instances()` method.
	LAST_TIME_0 int
	LAST_TIME_1 int
)

func SetParameterDefault() {
	MAXPROCESSES = min(max(runtime.NumCPU(), 4), 6)

	NEW_BASE_TIMEOUT_FACTOR = 15 // => 1.5
	STAGE_TIMEOUT_MIN = 5
	NEW_STAGE_TIMEOUT_FACTOR = 15 // => 1.5
	LAST_TIME_0 = 5
	LAST_TIME_1 = 50

	DEBUG = false
}

func init() {
	SetParameterDefault()
}

/*
Various strategies are used to try to achieve a – possibly imperfect –
timetable within a specified time. It is impossible to guarantee that all
constraints will be satisfied within a given time, so in order to place
all the activities within this time it may be necessary to drop some of
them.

A certain degree of parallel processing is assumed – too few (less than four?)
processor cores is likely to result in a very significant slowdown.

The main function (`StartGeneration`) starts a run with the fully constrained
data and a second run with all the "non-basic" constraints removed. Fixed
activity placements and blocked time-slots (for teachers, classes, and rooms)
are regarded as basic, non-negotiable.

A `TtInstance` structure is constructed to manage the data for each
timetable generation run, each run having its own goroutine. Each instance
has its own individual timeout. There is also a global timeout to stop
all instances which are still running.

Once these initial instances have been started, a "tick-loop" (which is
triggered every second) is entered. This monitors the progress of each active
instance and handles the actions resulting from their completion, whether
successful or not.

Should a fully constrained instance complete successfully within the
allotted time, all other instances are terminated and its result will be
saved.

When the unconstrained instance completes successfully, a series of further
instances is queued for running, each specifying the addition of a list of
(hard) constraints of a single type. Thus for each type of constraint an
instance is constructed. Using timeouts leading to binary divisions of these
lists an attempt is made to find individual "difficult" constraints, which can
then be disabled in order to get full activity placement within a reasonable
time. Parallel processing can be of some assistance here.

TODO: Should the unconstrained instance fail to complete successfully within
its allotted time, further steps may be taken to trace difficulties within the
activity collection, perhaps identifying "difficult" classes or teachers.

When a single-constraint-type instance completes successfully, it is used as
a new base (`current_instance`) for the addition of further constraints. All
the remaining constraint-type instances are stopped and restarted with this
new base. If a constraint-type instance is timed out, it is stopped and split
into two halves, which then run in its place. If there are no halves (only
one constraint being added) there is no successor, the constraint is dropped.

When an instance completes successfully within the allotted time, its result
is saved as a JSON file, so that the best result so far gradually encompasses
more of the constraints. When all the constraints have been tested with a
certain timeout, the rejected ones are tried again, but with longer timeouts.

The results include diagnostic information (at least an indication of which
constraints were dropped).

TODO: There is probably no general "optimum" value for the various timeout
parameters, that is likely to depend on the data. But perhaps values can be
found which are frequently useful. It might be helpful to use shorter overall
timeouts during the initial phases of testing the data, to identify potential
problem areas without long processing delays. For later phases longer times
may be necessary (depending on the difficulty of the data).
*/

func StartGeneration(cdata *ConstraintData, TIMEOUT int) {
	constraint_data = cdata
	LastResult = nil

	// Catch termination signal
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	runqueue := &RunQueue{
		Queue:      nil,
		Active:     map[*TtInstance]struct{}{},
		MaxRunning: MAXPROCESSES,
		Next:       0,
	}

	// Global data
	Ticks = 0

	// First run: all constraints enabled.
	// On successful completion, all other instances should be stopped.
	// If it fails, just this instance should be wound up. Otherwise it
	// should run until it times out, at which point any other active
	// instances should be stopped and the "best" solution at this point
	// chosen.
	enabled := make([]bool, cdata.Constraints)
	for i := range cdata.Constraints {
		enabled[i] = true
	}
	full_instance := &TtInstance{
		Tag:               "COMPLETE",
		Timeout:           0,
		ConstraintEnabled: enabled,
	}
	// Add to run queue
	runqueue.add(full_instance)

	// Instance without soft constraints (if any, otherwise same as full
	// instance) – enable only the hard constraints.
	enabled = make([]bool, cdata.Constraints)
	for _, ilist := range cdata.HardConstraintMap {
		for _, i := range ilist {
			enabled[i] = true
		}
	}
	hard_instance := &TtInstance{
		Tag:               "HARD_ONLY",
		Timeout:           0,
		ConstraintEnabled: enabled,
	}
	// Add to run queue
	runqueue.add(hard_instance)

	// Unconstrained instance
	CYCLE_TIMEOUT = STAGE_TIMEOUT_MIN
	enabled = make([]bool, cdata.Constraints)
	null_instance := &TtInstance{
		Tag:               "ONLY_BLOCKED_SLOTS",
		Timeout:           CYCLE_TIMEOUT,
		ConstraintEnabled: enabled,
	}
	// Add to run queue
	runqueue.add(null_instance)

	// Start stage 0
	stage := 0
	soft := false
	full_progress := 0       // current percentage
	full_progress_ticks := 0 // time of last increment
	hard_progress := 0       // current percentage
	hard_progress_ticks := 0 // time of last increment

	// *** Ticker loop ***
	var constraint_list []*TtInstance
	// The (successful) instance on which current trials are based:
	var current_instance *TtInstance
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	defer func() {
		// Tidy up
		r := recover()
		if r != nil {
			base.Message.Println("(TODO) *** RECOVER ***", r)
			debug.PrintStack()
		}
		for {
			// Wait for active instances to finish, stopping them if
			// necessary.
			count := 0
			for instance := range runqueue.Active {
				if instance.RunState == 0 {
					Backend.Tick(instance)
					count++
					abort_instance(instance)
				}
			}
			if count == 0 {
				break
			}
			<-ticker.C
		}
		if !DEBUG {
			// Remove all remaining temporary files
			Backend.Tidy(WorkingDir)
		}
		if LastResult != nil {
			// Save result of last successful instance.
			//b, err := json.Marshal(LastResult)
			b, err := json.MarshalIndent(LastResult, "", "  ")
			if err != nil {
				panic(err)
			}
			fpath := filepath.Join(WorkingDir, "Result.json")
			f, err := os.Create(fpath)
			if err != nil {
				panic("Couldn't open output file: " + fpath)
			}
			defer f.Close()
			_, err = f.Write(b)
			if err != nil {
				panic("Couldn't write result to: " + fpath)
			}
		}
	}()

tickloop:
	// Start queued instances if there are free processors.
	// Quit the loop if there are no instances left and no constraints
	// pending (stage -1).
	for runqueue.update_queue() != 0 || stage >= 0 {
		select {
		case <-ticker.C:
		case <-sigChan:
			base.Message.Printf("(TODO) *** INTERRUPTED @ %d ***\n", Ticks)
			break tickloop
		}

		Ticks++
		runqueue.update_instances()

		if full_instance.ProcessingState == 1 {
			// Cancel all other runs and return this instance as result.
			current_instance = full_instance
			new_current_instance(current_instance)
			base.Message.Printf("(TODO) *** All constraints OK @ %d ***\n", Ticks)
			break
		} else {
			p := full_instance.Progress
			if p > full_progress {
				full_progress = p
				full_progress_ticks = Ticks
				base.Message.Printf(
					"(TODO) [%d] ? %s (%d @ %d)\n",
					Ticks,
					full_instance.Tag,
					full_progress,
					full_progress_ticks,
				)
			}
		}

		if !soft {
			if hard_instance.ProcessingState == 1 {
				// Set as current and start processing soft constraints.
				current_instance = hard_instance
				new_current_instance(current_instance)
				base.Message.Printf(
					"(TODO) *** All hard constraints OK @ %d ***\n", Ticks)
				// Cancel everything except full instance.
				if null_instance.ProcessingState == 0 {
					abort_instance(null_instance)
				}
				for _, instance := range constraint_list {
					if instance.ProcessingState == 0 {
						abort_instance(instance)
					}
					// Indicate that a queued instance is not to be started
					instance.ProcessingState = 3
				}
				constraint_list = nil
				soft = true
				base.Message.Printf(
					"(TODO) [%d] Soft constraints based on hard-only instance", Ticks)
			} else {
				p := hard_instance.Progress
				if p > hard_progress {
					hard_progress = p
					hard_progress_ticks = Ticks
					base.Message.Printf(
						"(TODO) [%d] ? %s (%d @ %d)\n",
						Ticks,
						hard_instance.Tag,
						hard_progress,
						hard_progress_ticks,
					)
				}
			}
		}

		if Ticks == TIMEOUT {
			base.Message.Printf(
				"(TODO) [%d] TIMEOUT (%d @ %d) (%d @ %d) \n",
				Ticks,
				full_progress,
				full_progress_ticks,
				hard_progress,
				hard_progress_ticks,
			)
			break
		}

		if stage < 0 {
			continue
		}

		// There should be no problem if there are no constraints to add.

		if stage == 0 {
			// During stage 0 only `full_instance`, `hard_instance` and
			// `null_instance` are running.
			switch null_instance.ProcessingState {
			case 0:
				if null_instance.Ticks == null_instance.Timeout {
					abort_instance(null_instance)
				}
				continue
			case 1:
				// The null instance completed successfully.
				current_instance = null_instance
				new_current_instance(current_instance)
				// Start trials of single constraint types.
				base.Message.Printf("(TODO) [%d] INITIAL CONSTRAINT-TYPES: %d\n",
					Ticks, len(constraint_list))
				// not continue!
			default:
				// The null instance failed.
				stage = -10
				base.Message.Printf(
					"(TODO) [%d] Unconstrained instance failed", Ticks)

				base.Error.Println(" ... " + null_instance.Message)

				//TODO: Seek problems in the unconstrained data.
				panic("TODO")
				continue
			}
		}

		// This is the main phase, during which instances are run which
		// try to add the (as yet not included) constraints of a single
		// type, with a given timeout. A certain number of these can be
		// run in parallel. If one completes successfully, it is removed
		// from the constraint list. All the other instances are stopped
		// and the successful instance is used as the base for a new
		// cycle. Depending on the time this instance took to complete,
		// the timeout may be increased.
		// There is some flexibility around the timeouts. If an instance
		// seems to be progressing too slowly, it can be halted
		// immediately. On the other hand, if the instance looks like it
		// might complete if given a little more time, the timeout is
		// delayed.
		// When an instance times out, it is removed from the constraint
		// list. It is split into two, each with half of the constraints,
		// the new instances being added to the constraint list and to the
		// end of the run-queue. If the instance has only one constraint
		// to add, no new instance is started – until the next stage.
		// When the constraint list is empty, the stage ends. For the next
		// stage, the as yet unincluded constraints are collected again
		// and the timeout is increased somewhat.

		// Should it come to pass that all the constraints have been
		// added successfully (unlikely, because the overall timeout
		// will have been reached or the hard-only or full instance
		// will have completed already), stage -1 is entered, which will
		// cause the loop to be exited.

		next_timeout := 0 // non-zero => "restart with new base"

		// See if an instance has completed successfully.
		for i, instance := range constraint_list {
			if instance.ProcessingState == 1 {
				// Completed successfully, make this instance the new base.
				current_instance = instance
				new_current_instance(current_instance)
				next_timeout = max(
					instance.Ticks*NEW_BASE_TIMEOUT_FACTOR/10,
					CYCLE_TIMEOUT)
				// Remove it from constraint list.
				constraint_list = slices.Delete(
					constraint_list, i, i+1)

				// next_timeout != 0 and current_instance is new
				break
			}
		}
		if len(constraint_list) == 0 {
			// ... all current constraint trials finished.
			// Start trials of remaining constraints, hard then soft.
			CYCLE_TIMEOUT = max(CYCLE_TIMEOUT,
				current_instance.Ticks) * NEW_STAGE_TIMEOUT_FACTOR / 10
			var n int
			constraint_list, n = get_basic_constraints(
				current_instance, soft)
			if n == 0 {
				if soft {
					break // solution found
				} else {
					base.Message.Printf(
						"(TODO) [%d] Soft constraints based on accumulated instance", Ticks)
					soft = true
					constraint_list, n = get_basic_constraints(
						current_instance, soft)
					if n == 0 {
						break // solution found
					} else {
						// The hard-only instance is no longer needed.
						if hard_instance.ProcessingState == 0 {
							abort_instance(hard_instance)
						}
					}
				}
			}
			stage++
			// Queue instances for running
			for _, bc := range constraint_list {
				runqueue.add(bc)
			}
			hs := "hard"
			if soft {
				hs = "soft"
			}
			base.Message.Printf(
				"(TODO) [%d] Stage %d (%s): %d (timeout %d)\n",
				Ticks, stage, hs, n, CYCLE_TIMEOUT)
			continue
		}

		// Seek failed instances, which should be split.
		// If there is a new base, stop the old instances and
		// restart them accordingly.
		split_instances := []*TtInstance{}
		new_constraint_list := []*TtInstance{}
		for _, instance := range constraint_list {
			if instance.ProcessingState == 2 {
				// timed out / failed

				// Split if more than one instance in list
				if len(instance.Constraints) > 1 {
					timeout := next_timeout
					if timeout == 0 {
						timeout = instance.Timeout
					}
					nhalf := len(instance.Constraints) / 2
					split_instances = append(split_instances,
						new_instance(
							current_instance,
							instance.Tag,
							instance.ConstraintType,
							instance.Constraints[:nhalf],
							timeout,
							soft))
					split_instances = append(split_instances,
						new_instance(
							current_instance,
							instance.Tag,
							instance.ConstraintType,
							instance.Constraints[nhalf:],
							timeout,
							soft))
				} else if len(instance.Constraints) == 0 {
					panic("Bug, expected constraint(s)")
				}
			} else {
				if next_timeout != 0 {
					// Cancel existing instance
					if instance.ProcessingState == 0 {
						abort_instance(instance)
					}
					// Indicate that a queued instance is not to be started
					instance.ProcessingState = 3
					// Build new instance
					instance = new_instance(
						current_instance,
						instance.Tag,
						instance.ConstraintType,
						instance.Constraints,
						next_timeout,
						soft)
					runqueue.add(instance)
				}
				new_constraint_list = append(
					new_constraint_list, instance)
			}
		}
		constraint_list = append(new_constraint_list,
			split_instances...)
		for _, instance := range split_instances {
			runqueue.add(instance)
		}
	} // tickloop: end

	result := current_instance

	hnn := 0
	hnall := 0
	for c, clist := range cdata.HardConstraintMap {
		n := 0
		for _, cix := range clist {
			if result.ConstraintEnabled[cix] {
				n++
			}
		}
		if len(clist) != 0 {
			fmt.Printf("$ (HARD) %s: %d / %d\n", c, n, len(clist))
			hnn += n
			hnall += len(clist)
		}
	}
	snn := 0
	snall := 0
	for c, clist := range cdata.SoftConstraintMap {
		n := 0
		for _, cix := range clist {
			if result.ConstraintEnabled[cix] {
				n++
			}
		}
		if len(clist) != 0 {
			fmt.Printf("$ (SOFT) %s: %d / %d\n", c, n, len(clist))
			snn += n
			snall += len(clist)
		}
	}
	fmt.Printf("$ ALL CONSTRAINTS: (hard) %d / %d  (soft) %d / %d\n",
		hnn, hnall, snn, snall)

	//TODO
	base.Message.Printf("(TODO) RESULT: %s\n", result.Tag)
}

func abort_instance(instance *TtInstance) {
	if !instance.Stopped {
		timetable.BACKEND.Abort(instance)
		instance.Stopped = true
	}
}

func new_instance(
	instance_0 *TtInstance,
	tag string,
	constraint_type timetable.ConstraintType,
	constraint_indexes []ConstraintIndex,
	timeout int,
	soft bool,
) *TtInstance {
	// Prepare instance "name"
	InstanceCounter++
	if i := strings.Index(tag, "~"); i >= 0 {
		tag = tag[i+1:]
	}
	tag = fmt.Sprintf("z%05d~%s", InstanceCounter, tag)

	enabled := slices.Clone(instance_0.ConstraintEnabled)
	// Add the new constraints
	for _, c := range constraint_indexes {
		enabled[c] = true
	}

	// Make a new `TtInstance`
	instance := &TtInstance{

		Tag: tag,

		//TODO--? InstanceDir string // working space for this instance

		Timeout: timeout,

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
