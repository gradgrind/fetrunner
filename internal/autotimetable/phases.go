package autotimetable

import (
	"fmt"
	"strconv"
)

const (
	PHASE_BASIC = iota
	PHASE_HARD
	PHASE_SOFT
	PHASE_FINISHED
)

const NEARLY_FINISHED = 80 // (progress, %) TODO: experimental, what is a good value?

/* On entering each phase, all running instances except the special
processes appropriate to the new phase are aborted.

## In PHASE_BASIC all the special constraints are running initially:

- _UNCONSTRAINED
- _NA_ONLY (if there are any hard not-available constraints)
- _HARD_ONLY
- _COMPLETE

There is no `current_instance` (it is `nil`) until a run has completed
successfully. Hard constraints restricting the availablity of classes, teachers
and rooms are added (see `GetResourceUnavailableConstraintTypes()`).
These are regarded as especially important constraints.

If _UNCONSTRAINED completes successfully and `current_instance` is still unset,
this becomes the current instance.

If _NA_ONLY completes successfully the PHASE_HARD is entered.

If _HARD_ONLY completes successfully PHASE_SOFT is entered.

If _COMPLETE completes successfully PHASE_FINISHED is entered.

If no more instances are running, PHASE_HARD is entered if there is a
current instance – otherwise the whole process finishes unsuccessfuly.

If SKIP_HARD is set, processing starts in PHASE_SOFT, so PHASE_BASIC
will not be entered.

## In PHASE_HARD the only special constraints which may be running are:

- _HARD_ONLY
- _COMPLETE

If _HARD_ONLY completes successfully PHASE_SOFT is entered.

If _COMPLETE completes successfully PHASE_FINISHED is entered.

Also if there are no remaining constraint-addition processes running,
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
	bdata := attdata.BaseData
	if attdata.current_instance != nil {
		// Adjust the initial time-out guideline.
		attdata.cycle_timeout = (max(attdata.cycle_timeout,
			attdata.current_instance.Ticks) *
			attdata.Parameters.NEW_PHASE_TIMEOUT_FACTOR) / 10
	}
new_phase:
	attdata.phase = p
	bdata.Logger.Result(".PHASE", strconv.Itoa(p))

	// Abort special instances which are no longer relevant.
	// Note that anything halted here will not be restarted, because this
	// is a transition to a new phase. ABORT_NEW_CYCLE is used because no
	// error message should arise for the constraints.
	if p == PHASE_BASIC {
		// no current instance
		new_instance_list, _ := attdata.get_basic_constraints(attdata.null_instance)
		attdata.set_runqueue(new_instance_list)
		return
	}
	if p == PHASE_HARD && attdata.current_instance == nil {
		bdata.Logger.Error("Unconstrained instance failed:\n:::+\n%s\n:::-",
			attdata.null_instance.Message)
		p = PHASE_FINISHED
		goto new_phase
	}
	attdata.abort_instance(attdata.null_instance, ABORT_NEW_CYCLE)
	attdata.abort_instance(attdata.na_instance, ABORT_NEW_CYCLE)
	if p >= PHASE_SOFT {
		attdata.abort_instance(attdata.hard_instance, ABORT_NEW_CYCLE)
	}
	if p == PHASE_FINISHED {
		attdata.abort_instance(attdata.full_instance, ABORT_NEW_CYCLE)
		return
	}

	// Abort all non-special processes.
	for _, instance := range attdata.active_instances {
		if len(instance.Constraints) != 0 {
			attdata.abort_instance(instance, ABORT_NEW_CYCLE)
		}
	}

	// Initialize constraint-instance list, only in PHASE_HARD and PHASE_SOFT.
	if new_instance_list, n := attdata.get_basic_constraints(attdata.current_instance); n == 0 {
		// Skip to next phase
		p++
		goto new_phase
	} else {
		attdata.set_runqueue(new_instance_list)
	}
}

func (attdata *AutoTtData) tick_phase() bool {
	bdata := attdata.BaseData
	logger := bdata.Logger
	p := attdata.phase
	if p >= PHASE_FINISHED {
		panic("Bug, tick_phase in PHASE_FINISHED+")
	}
	if attdata.full_instance.RunState == INSTANCE_SUCCESSFUL {
		// Set as current and prepare to wind up process.
		attdata.current_instance = attdata.full_instance
		logger.Result(".ALL_OK", "All constraints OK")
		attdata.new_current_instance(bdata, attdata.current_instance)
		attdata.enter_phase(PHASE_FINISHED)
		return true
	}
	if p <= PHASE_HARD && attdata.hard_instance.RunState == INSTANCE_SUCCESSFUL {
		// Set as current and prepare for processing soft constraints.
		attdata.current_instance = attdata.hard_instance
		logger.Result(".HARD_OK", "All hard constraints OK")
		attdata.new_current_instance(bdata, attdata.current_instance)
		attdata.enter_phase(PHASE_SOFT)
		return true
	}
	if p == PHASE_BASIC {
		if attdata.na_instance != nil && attdata.na_instance.RunState == INSTANCE_SUCCESSFUL {
			// Set as current.
			attdata.current_instance = attdata.na_instance
			bdata.Logger.Result(".NA_OK", "All hard NotAvailable constraints OK")
			attdata.new_current_instance(bdata, attdata.current_instance)
			attdata.enter_phase(PHASE_HARD)
			return true
		}
		if attdata.null_instance.RunState == INSTANCE_SUCCESSFUL && attdata.current_instance == nil {
			attdata.current_instance = attdata.null_instance
			logger.Result(".NULL_OK", "Without constraints OK")
			attdata.new_current_instance(bdata, attdata.current_instance)
			// Don't change phase.
		}
	}

	// Handle the currently active constraint-adding instances.
	// Go to next phase if no remaining constraint-adding instances.
	if attdata.phase_main() {
		attdata.enter_phase(p + 1)
		return true
	}
	return false
}

/*
    Main processing phase(s), accumulating constraints.

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

If an instance seems to be progressing too slowly, it will be halted, and
removed from the instance list. This allows another instance to be started.
The halted process is split into two, each with half of the constraints,
the new instances being added to the end of the run-queue.

If an instance has only one constraint to add, it is run without a "timeout",
which means that different criteria may apply for checking whether it is stuck.
Once this instance is deemed to be stuck or too slow, it is discarded completely
(its constraint is judged to be "impossible").
TODO: Might there be circumstances under which a further attempt is made
to include this constraint?

If there are no more instances to run, return `true`, so that the next phase
can be entered. Otherwise return `false`.
*/

func (attdata *AutoTtData) phase_main() bool {
	bdata := attdata.BaseData
	logger := bdata.Logger
	base_instance := attdata.current_instance

	// Seek failed instances, which should be retried with a longer timeout or split.
	// Also check for a successful completion. Collect instances which would need
	// restarting in a new cycle.
	var (
		failed       []*TtInstance = nil
		to_continue  []*TtInstance = nil // only relevant at the end of a "cycle"
		next_timeout int           = 0   // set non-zero at the end of a cycle
		n_active     int           = 0   // number of running instances
	)
	for _, instance := range attdata.active_instances {
		if len(instance.Constraints) == 0 { // a special instance
			// Handled in `tick_phase()`
			continue
		}
		switch instance.RunState {
		case INSTANCE_SUCCESSFUL: // completed successfully
			if next_timeout == 0 {
				// This instance will be the new base.
				attdata.current_instance = instance
				base_instance = instance
				attdata.new_current_instance(bdata, instance)
				next_timeout = max(
					(instance.Ticks*attdata.Parameters.NEW_BASE_TIMEOUT_FACTOR)/10,
					attdata.cycle_timeout)
				/* for debugging
				   // next_timeout != 0 and base_instance = current_instance is new
				   fmt.Printf("$ %s  n: %d  t: %d (%d, %d)\n",
				       instance.ConstraintType, len(instance.Constraints),
				       next_timeout, instance.Ticks, attdata.cycle_timeout)
				*/
			} else {
				to_continue = append(to_continue, instance)
			}
		case INSTANCE_RUNNING: // running
			n_active++
			to_continue = append(to_continue, instance)
		//case ABORT_NEW_CYCLE, INSTANCE_TIMED_OUT, INSTANCE_CANCELLED: //TODO?
		case ABORT_TIMED_OUT, INSTANCE_FAILED:
			//TODO???
			// Gather all unsuccessfully ended constraints here, whether with >1
			// constraints, a single constraint, timed out or with error.
			failed = append(failed, instance)
		}
	}

	// Rebuild or extend run queue, according to whether a new cycle is beginning.
	timeout := next_timeout
	if timeout == 0 {
		// No new base yet
		timeout = attdata.cycle_timeout // for new split instances
	} else {
		// There is a new base, stop the old instances and queue them for restarting.
		new_queue := []*TtInstance{} // restart run queue
		old_queue := attdata.get_runqueue()
		for _, instance := range to_continue {
			attdata.abort_instance(instance, ABORT_NEW_CYCLE)

			//TODO: Split it if progress was slow (but not slow enough to
			// trigger an Abort)?

			if instance.Progress >= NEARLY_FINISHED {
				// Build new instance and queue it.
				new_queue = append(new_queue, attdata.new_instance(
					base_instance,
					instance.ConstraintType,
					instance.Weight,
					instance.Constraints,
					timeout))
			} else {
				// Add to the back of the old run queue, here still with the old base instance.
				old_queue = append(old_queue, instance)
			}
		}
		// Add rebased old_queue to new_queue.
		for _, instance := range old_queue {
			new_queue = append(new_queue, attdata.new_instance(
				base_instance,
				instance.ConstraintType,
				instance.Weight,
				instance.Constraints,
				timeout))
		}
		attdata.set_runqueue(new_queue)
	}

	// Split and add failed instances.
	for _, instance := range failed {
		if len(instance.Constraints) > 1 {
			sit := []string{}
			for _, si := range attdata.split_instance(
				instance, base_instance, timeout) {
				attdata.queue_instance(si)
				sit = append(sit,
					fmt.Sprintf("%d:%s", si.Index, si.ConstraintType))
			}
			logger.Info("(SPLIT) %d:%s -> %v",
				instance.Index, instance.ConstraintType, sit)
		} else if len(instance.Constraints) == 1 {
			// Only a single constraint
			switch instance.RunState {
			case INSTANCE_FAILED:
				if len(instance.Message) != 0 {
					attdata.ConstraintErrors[instance.Constraints[0]] = instance.Message
				} else {
					attdata.ConstraintErrors[instance.Constraints[0]] = "UnknownFailure"
				}
				logger.Info("InstanceFailed: %d\n:::+\n%s\n:::-", instance.Index, instance.Message)
				logger.Result(
					".ELIMINATE", fmt.Sprintf("%s.%d",
						attdata.Backend.ConstraintName(instance),
						instance.Constraints[0]))

			case ABORT_TIMED_OUT:
				attdata.timed_out_instances = append(attdata.timed_out_instances, instance)
				logger.Result(".TIMED_OUT", fmt.Sprintf("%s.%d.%d.%d",
					attdata.Backend.ConstraintName(instance),
					instance.Constraints[0],
					instance.Progress, instance.Ticks))
			case INSTANCE_CANCELLED:
				// No error messages
			default:
				panic("Unexpected RunState: " + strconv.Itoa(instance.RunState))
			}
		} else {
			panic("Bug, expected constraint(s)")
		}
	}
	// Return `true` if all the instances have been processed.
	return n_active == 0 && attdata.n_queued() == 0
}

func (attdata *AutoTtData) split_instance(
	instance *TtInstance, base_instance *TtInstance, timeout int,
) []*TtInstance {
	nhalf := len(instance.Constraints) / 2
	return []*TtInstance{
		attdata.new_instance(
			base_instance,
			instance.ConstraintType,
			instance.Weight,
			instance.Constraints[:nhalf],
			timeout),
		attdata.new_instance(
			base_instance,
			instance.ConstraintType,
			instance.Weight,
			instance.Constraints[nhalf:],
			timeout),
	}
}
