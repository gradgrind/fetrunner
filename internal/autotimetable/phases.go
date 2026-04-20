package autotimetable

import (
	"fetrunner/internal/base"
	"fmt"
	"strconv"
	"strings"
)

const (
	PHASE_BASIC    = 0
	PHASE_HARD     = 1
	PHASE_SOFT     = 2
	PHASE_FINISHED = 3
)

const NEARLY_FINISHED = 80 // (progress, %) TODO: experimental, what is a good value?

/* On entering each phase, all running instances except the special
instances appropriate to the new phase are aborted.

Initially there is no `current_instance`. Entering at PHASE_BASIC (the normal case) sets
it to `null_instance`, which is initially running. Entering at PHASE_SOFT (with
SKIP_HARD set), it is set to `hard_instance`, which is initially running. After an
instance has completed successfully, `current_instance` will always be an
instance which has completed successfully and having run-state INSTANCE_ACCEPTED. This
run-state can be used to determine whether there is a final result.

## In PHASE_BASIC all the special instances are running initially:

- _UNCONSTRAINED
- _PRIORITY (if there are any "priority" constraints)
- _HARD_ONLY
- _COMPLETE

Hard constraints restricting the availablity of classes, teachers and rooms,
and fixed activity placements are added (see `GetPhase0ConstraintTypes()`).
These are regarded as especially important constraints.

If _UNCONSTRAINED completes successfully, at least this is available as a result, but
if there are "priority" constraints the phase is not yet changed. If this instance
fails, the whole process ends unsuccessfully.

If _PRIORITY completes successfully PHASE_HARD is entered.

If _HARD_ONLY completes successfully PHASE_SOFT is entered.

If _COMPLETE completes successfully PHASE_FINISHED is entered.

If no more constraint-addition instances are running, PHASE_HARD is entered
if there is an "accepted" current instance – otherwise the whole process finishes
unsuccessfuly.

If the "Parameter" SKIP_HARD is set, processing starts in PHASE_SOFT, so
PHASE_BASIC will not be entered.

## In PHASE_HARD the only special constraints which may be running are:

- _HARD_ONLY
- _COMPLETE

If _HARD_ONLY completes successfully PHASE_SOFT is entered.

If _COMPLETE completes successfully PHASE_FINISHED is entered.

Also if there are no remaining constraint-addition instances running,
PHASE_SOFT is entered.

If SKIP_HARD is set, processing starts in PHASE_SOFT, so PHASE_HARD
will not be entered.

## In PHASE_SOFT the only special constraint which may be running is:

- _COMPLETE

If _COMPLETE completes successfully PHASE_FINISHED is entered.

Also if there are no remaining constraint-addition processes running,
PHASE_FINISHED is entered.

## In PHASE_FINISHED no instances should be running. After any instances
which are still running on entry have completed the whole process
is finished.
*/

// Enter new phase.
func (attdata *AutoTtData) enter_phase(p int) {
	// Adjust the initial time-out guideline.
	if attdata.current_instance == nil {
		attdata.cycle_timeout = MIN_TIMEOUT
	} else {
		attdata.cycle_timeout = (max(
			MIN_TIMEOUT,
			attdata.cycle_timeout,
			attdata.current_instance.Ticks) * TtParameters.NEW_PHASE_TIMEOUT_FACTOR) / 10
	}

	attdata.phase = p
	base.LogResult(".PHASE", strconv.Itoa(p))

	// Abort special instances which are no longer relevant.
	// Note that anything halted here will not be restarted, because this
	// is a transition to a new phase. ABORT_NEW_CYCLE is used because no
	// error message should arise for the constraints.
	switch p {
	case PHASE_BASIC:
		attdata.current_instance = attdata.null_instance
		attdata.get_basic_constraints()
		return
	case PHASE_HARD:
		attdata.abort_instance(attdata.null_instance)
		attdata.abort_instance(attdata.priority_instance)
	case PHASE_SOFT:
		if attdata.current_instance == nil {
			// SKIP_HARD
			attdata.current_instance = attdata.hard_instance
		} else {
			attdata.abort_instance(attdata.hard_instance)
		}
	case PHASE_FINISHED:
		attdata.abort_instance(attdata.null_instance)
		attdata.abort_instance(attdata.priority_instance)
		attdata.abort_instance(attdata.hard_instance)
		attdata.abort_instance(attdata.full_instance)
		return
	}
	// From here only in PHASE_HARD and PHASE_SOFT ...

	// Abort all non-special processes.
	for _, instance := range attdata.active_instances {
		if len(instance.Constraints) != 0 {
			attdata.abort_instance(instance)
		}
	}

	// Initialize constraint-instance list.
	attdata.get_basic_constraints()
}

func (attdata *AutoTtData) tick_phase() bool {
	p := attdata.phase
	if p >= PHASE_FINISHED {
		panic("Bug, tick_phase in PHASE_FINISHED+")
	}
	if attdata.full_instance.RunState == INSTANCE_SUCCESSFUL {
		// Set as current and prepare to wind up process.
		attdata.current_instance = attdata.full_instance
		base.LogResult(".ALL_OK", "All constraints OK")
		attdata.new_current_instance()
		attdata.enter_phase(PHASE_FINISHED)
		return true
	}
	if attdata.hard_instance.RunState == INSTANCE_SUCCESSFUL {
		base.LogResult(".HARD_OK", "All hard constraints OK")
		if p == PHASE_SOFT {
			// If `hard_instance` is no longer `current_instance` it should have been halted,
			// and thus have a different run state.
			if attdata.hard_instance != attdata.current_instance {
				panic("current_instance should be hard_instance")
			}
		} else {
			// Set as current and prepare for processing soft constraints.
			attdata.current_instance = attdata.hard_instance
			attdata.new_current_instance()
			attdata.enter_phase(PHASE_SOFT)
			return true
		}
		attdata.new_current_instance()
		// Don't change phase.
	}
	if p == PHASE_BASIC {
		if attdata.priority_instance != nil && attdata.priority_instance.RunState == INSTANCE_SUCCESSFUL {
			// Set as current and prepare for processing remaining hard constraints.
			attdata.current_instance = attdata.priority_instance
			base.LogResult(".PRIORITY_OK", "All priority constraints OK")
			attdata.new_current_instance()
			attdata.enter_phase(PHASE_HARD)
			return true
		}
		switch attdata.null_instance.RunState {
		case INSTANCE_SUCCESSFUL:
			// If `null_instance` is no longer `current_instance` it should have been halted,
			// and thus have a different run state.
			if attdata.null_instance != attdata.current_instance {
				panic("current_instance should be null_instance")
			}
			base.LogResult(".NULL_OK", "Without constraints OK")
			attdata.new_current_instance()
			// Don't change phase.
		case INSTANCE_FAILED:
			base.LogError("--UNCONSTRAINED_FAILED\n%s\n ***",
				strings.TrimSpace(attdata.null_instance.Message))
			attdata.enter_phase(PHASE_FINISHED)
			return true
		}
	}

	// Handle the currently active constraint-adding instances.
	// Go to next phase if no remaining constraint-adding instances and `current_instance`
	// no longer running – it can only be running before any successful completion.
	if attdata.phase_main() && attdata.current_instance.RunState != INSTANCE_RUNNING {
		attdata.enter_phase(p + 1)
		return true
	}
	return false
}

/*
	Tick processing for non-special instances, accumulating constraints.

`phase_main()` is run in all phases except PHASE_FINISHED.
Generator instances are run which try to add the (as yet not included)
constraints of a single type, with a "timeout". A certain number of these can
be run in parallel. If one completes successfully, it is removed from the
constraint list, all the other instances are stopped (including any others
that might have completed successfully) and this successful instance is used
as the base for a new cycle. Depending on the time this instance took to
complete, the timeout may be increased.

TODO: There might be useful tweaks to the ordering of the constraint
instances (and splitting?) in the next cycle, based on their progress in
the current one.

If an instance seems to be progressing too slowly, it will be halted. If it
adds more than a single constraint, it is added to the back of the queue.
Otherwise it is added to a "timed-out" list, which is currently unused (TODO?).
As an active process is thus halted, this allows another instance to be started.
Before the halted process is started again (from the queue), it is split into
two, each with half of the constraints.

If an instance has only one constraint to add, it is run without a "timeout",
which means that different criteria may apply for checking whether it is stuck.

If there are no more instances to run, return `true`, so that the next phase
can be entered. Otherwise return `false`.
*/
func (attdata *AutoTtData) phase_main() bool {
	var (
		// Gather all unsuccessfully ended constraints here, whether with a single
		// or >1 constraints, timed out or with error.
		failed      []*TtInstance = nil
		to_continue []*TtInstance = nil   // only relevant at the end of a "cycle"
		new_cycle   bool          = false // set to true at the end of a cycle
		n_active    int           = 0     // number of running instances
	)

	// Seek failed instances, which should be retried with a longer timeout or split.
	// Also check for a successful completion. Collect instances which would need
	// restarting in a new cycle.
	for _, instance := range attdata.active_instances {
		if instance.Processed || len(instance.Constraints) == 0 { // a special instance
			// Handled in `tick_phase()`
			continue
		}
		switch instance.RunState {
		case INSTANCE_FAILED, ABORT_TIMED_OUT:
			instance.Processed = true
			failed = append(failed, instance)
		case INSTANCE_SUCCESSFUL:
			instance.Processed = true
			if !new_cycle {
				// Start a new cycle, with this instance as the new base.
				attdata.current_instance = instance
				attdata.new_current_instance()
				new_cycle = true
				attdata.cycle_timeout = max(
					(instance.Ticks*TtParameters.NEW_BASE_TIMEOUT_FACTOR)/10,
					attdata.cycle_timeout)
			} else {
				// This will need restarting in the new cycle.
				to_continue = append(to_continue, instance)
			}
		case INSTANCE_RUNNING:
			n_active++
			to_continue = append(to_continue, instance)
		case ABORT_NEW_CYCLE, INSTANCE_ABANDONED: // awaiting completion only, no action
		case INSTANCE_ACCEPTED: // nothing to do, already processed
		default:
			panic(fmt.Sprintf("Unexpected RunState, instance %d: %d",
				instance.Index, instance.RunState))
		}
	}

	// Rebuild or extend run queue, according to whether a new cycle is beginning.
	if new_cycle {
		// There is a new base, stop the old instances and queue them for restarting.
		new_queue := []*TtInstance{} // restart run queue
		old_queue := attdata.get_runqueue()

		/*/TODO--
		ilist := []string{}
		for _, ii := range old_queue {
			ilist = append(ilist, fmt.Sprintf("%d:%d", ii.Index, ii.RunState))
		}
		logger.Info("§OLDQUEUE %+v\n", strings.Join(ilist, ", "))
		*/

		// Stop superfluous base instances in PHASE_BASIC and PHASE_SOFT
		switch attdata.phase {
		case PHASE_BASIC:
			attdata.abort_instance(attdata.null_instance)
		case PHASE_SOFT:
			attdata.abort_instance(attdata.hard_instance)
		}

		for _, instance := range to_continue {
			attdata.abort_instance(instance)
			if instance.Progress >= NEARLY_FINISHED {
				// Add to the new run queue.
				new_queue = append(new_queue, instance)
			} else {
				// Add to the old run queue.
				old_queue = append(old_queue, instance)
			}
		}
		// Append old_queue to new_queue.
		new_queue = append(new_queue, old_queue...)
		attdata.set_runqueue(new_queue)
	}

	// Queue failed instances if they have more than one constraint.
	for _, instance := range failed {
		if len(instance.Constraints) > 1 {
			attdata.queue_instance(instance)
		} else if len(instance.Constraints) == 1 {
			// Only a single constraint
			if instance.RunState == INSTANCE_RUNNING {
				if len(instance.Message) != 0 {
					attdata.ConstraintErrors[instance.Constraints[0]] = instance.Message
				} else {
					attdata.ConstraintErrors[instance.Constraints[0]] = "UnknownFailure"
				}
				base.LogInfo("InstanceFailed: %d\n:::+\n%s\n:::-", instance.Index, instance.Message)
				base.LogResult(
					".ELIMINATE", fmt.Sprintf("%s.%d",
						attdata.Backend.ConstraintName(instance),
						instance.Constraints[0]))
			} else {
				// Collect timed-out single-constraint instances.
				attdata.timed_out_instances = append(attdata.timed_out_instances, instance)
				//TODO: At present these are not used, but there may be circumstances
				// under which they should be tried again?
				base.LogResult(".TIMED_OUT", fmt.Sprintf("%s.%d.%d.%d",
					attdata.Backend.ConstraintName(instance),
					instance.Constraints[0],
					instance.Progress, instance.Ticks))
			}
		} else {
			panic("Bug, expected constraint(s)")
		}
	}

	// Return `true` if all the instances have been processed.
	return n_active == 0 && attdata.n_queued() == 0
}
