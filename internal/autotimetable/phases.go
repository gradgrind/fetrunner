package autotimetable

import (
	"fmt"
	"slices"
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
successfully, no new instances are built in this phase.

If _UNCONSTRAINED completes successfully it becomes the current instance.

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
	if p == PHASE_BASIC {
		// Only special instances, no current instance
		return
	} else {
		attdata.abort_instance(attdata.null_instance, ABORT_NEW_CYCLE)
		attdata.abort_instance(attdata.na_instance, ABORT_NEW_CYCLE)
	}
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

	// Initialize constraint-instance list.
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
		if attdata.null_instance != nil {
			if attdata.null_instance.RunState == INSTANCE_SUCCESSFUL {
				// Set as current.
				attdata.current_instance = attdata.null_instance
				logger.Result(".NULL_OK", "Without constraints OK")
				attdata.new_current_instance(bdata, attdata.current_instance)
				attdata.null_instance = nil
			} else if attdata.null_instance.RunState > INSTANCE_SUCCESSFUL {
				// The null instance failed.
				logger.Error(
					"UnconstrainedInstanceFailed:\n:::+\n%s\n:::-",
					attdata.null_instance.Message)
				attdata.enter_phase(PHASE_FINISHED)
				return true
			}
		}
		if attdata.na_instance == nil {
			if attdata.null_instance == nil {
				attdata.enter_phase(PHASE_HARD)
				return true
			}
		} else if attdata.na_instance.RunState == INSTANCE_SUCCESSFUL {
			// Set as current.
			attdata.current_instance = attdata.na_instance
			bdata.Logger.Result(".NA_OK", "All hard NotAvailable constraints OK")
			attdata.new_current_instance(bdata, attdata.current_instance)
			attdata.enter_phase(PHASE_HARD)
			return true
		}
		return false // in this phase there are only special instances running
	}

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
	// Also check for a successful completion.
	var (
		failed       []*TtInstance = nil
		to_continue  []*TtInstance = nil // only relevant at the end of a "cycle"
		next_timeout int           = 0   // set non-zero at the end of a cycle
	)
	insertion_index := 0
	for _, instance := range attdata.active_instances {
		if len(instance.Constraints) == 0 { // a special instance
			if instance.RunState < 0 {
				// Retain instance in active list.
				attdata.active_instances[insertion_index] = instance
				insertion_index++
			}
			continue
		}

		//TODO: Do I want to distinguish between failed and timed out?

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
				// next_timeout != 0 and base_instance = current_instance is new
				continue
			}
			fallthrough
		case INSTANCE_RUNNING: // running
			to_continue = append(to_continue, instance)
		case ABORT_NEW_CYCLE, ABORT_TIMED_OUT: // aborted, but still running
			// Retain until completed //TODO???
			attdata.active_instances[insertion_index] = instance
			insertion_index++
		case INSTANCE_CANCELLED, INSTANCE_TIMED_OUT, INSTANCE_FAILED:
			//TODO???
			// Gather all unsuccessfully ended constraints here, whether with >1
			// constraints, a single constraint, abandoned, timed out or with error.
			failed = append(failed, instance)
		default:
			panic("Invalid RunState: " + strconv.Itoa(instance.RunState))
		}
	}
	attdata.active_instances = attdata.active_instances[0:insertion_index]

	//TODO

	timeout := next_timeout
	if timeout == 0 {
		// No new base yet, add instances which are still running.
		attdata.active_instances = append(attdata.active_instances, to_continue...)
		timeout = attdata.cycle_timeout // for new split instances
	} else {
		// There is a new base, stop the old instances and queue them for restarting.
		new_queue := []*TtInstance{}
		old_queue := attdata.run_queue
		for _, instance := range to_continue {
			if instance.RunState == INSTANCE_RUNNING {
				// Cancel existing instance
				attdata.abort_instance(instance, ABORT_NEW_CYCLE)
			}
			// Get progress because if well advanced, it should be prioritized.
			progress := instance.Progress

			//TODO: Split it if progress was slow (but not slow enough to
			// trigger an Abort)?
			//TODO: Adjust placement in queue according to progress rate?

			if progress > NEARLY_FINISHED {
				// Build new instance and queue it.
				new_queue = append(new_queue, attdata.new_instance(
					base_instance,
					instance.ConstraintType,
					instance.Weight,
					instance.Constraints,
					timeout))
			} else {
				// Add to the old run queue.
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
		attdata.run_queue = new_queue
	}
	//TODO: Add split failed instances.

	//TODO?

	split_instances := []*TtInstance{}
	for _, instance := range failed {
		if len(instance.Constraints) > 1 {
			sit := []string{}
			for _, si := range attdata.split_instance(
				instance, base_instance, timeout) {
				split_instances = append(split_instances, si)
				sit = append(sit,
					fmt.Sprintf("%d:%s", si.Index, si.ConstraintType))
			}
			logger.Info("(SPLIT) %d:%s -> %v",
				instance.Index, instance.ConstraintType, sit)
		} else if len(instance.Constraints) == 1 {
			// Only a single constraint
			//TODO: Save the constraint index with some measure of its
			// progress rate, to allow possible later reinclusion?
			if len(instance.Message) != 0 {
				attdata.ConstraintErrors[instance.Constraints[0]] =
					instance.Message
			}

			//TODO: I may well want to distinguish between the various
			// reasons for stopping.
			// If there was an actual error, the constraint should be
			// immediately blocked, this instance dropped.
			// If there was a timeout, at least one that doesn't look like
			// being completely stuck, it is possible that a (later) rerun
			// might be considered.
			// If the instance was aborted because of a completed instance,
			// the handling might depend on the progress, if the instance
			// is not too fresh.
		} else {
			panic("Bug, expected constraint(s)")
		}

	}

	//TODO: Here is the old code, bits of which will still be needed ...

	for _, instance := range attdata.active_instances.get_instances() {
		if instance.RunState == 2 { // timed out / failed
			// Split if more than one instance in list
			if len(instance.Constraints) > 1 {
				timeout := next_timeout
				if timeout == 0 {
					// If not rebasing, keep the old "timeout"
					timeout = instance.Timeout
				}
				sit := []string{}
				for _, si := range attdata.split_instance(
					instance, base_instance, timeout) {
					split_instances = append(split_instances, si)
					sit = append(sit,
						fmt.Sprintf("%d:%s", si.Index, si.ConstraintType))
				}
				logger.Info("(SPLIT) %d:%s -> %v",
					instance.Index, instance.ConstraintType, sit)
			} else if len(instance.Constraints) == 1 {
				// Only a single constraint
				//TODO: Save the constraint index with some measure of its
				// progress rate, to allow possible later reinclusion?
				if len(instance.Message) != 0 {
					attdata.ConstraintErrors[instance.Constraints[0]] =
						instance.Message
				}
			} else {
				panic("Bug, expected constraint(s)")
			}
		} else { // completed successfully or still running
			if next_timeout != 0 {
				// There is a new base instance ...
				// Cancel existing instance
				progress := 0
				switch instance.RunState {
				case -2: // already split and aborted, don't build a new instance
					continue
				case -1: // running
					attdata.abort_instance(instance)
					// Get progress because if well advanced, it should be prioritized
					progress = instance.Progress
				case 0: // queued, mark it "don't start"
					instance.RunState = 3
				}
				// Build new instance

				//TODO: Split it if progress was slow (but not slow enough to
				// trigger an Abort)?
				//TODO: Adjust placement in queue according to progress rate?

				instance = attdata.new_instance(
					base_instance,
					instance.ConstraintType,
					instance.Weight,
					instance.Constraints,
					next_timeout)

				if progress > NEARLY_FINISHED {
					renewed_instances = append(renewed_instances,
						weighted_instance{progress, instance})
				} else {
					//TODO???
					rq.add(instance)
				}
			}
			new_constraint_list = append(
				new_constraint_list, instance)
		}
	}
	//TODO: the constraint_instance_list needs rebasing
	attdata.run_queue = nil
	if len(renewed_instances) != 0 {
		slices.SortStableFunc(renewed_instances, func(a, b weighted_instance) int {
			return b.progress - a.progress // highest progress first
		})
		for _, ri := range renewed_instances {
			attdata.run_queue = append(attdata.run_queue,
				ri.instance)
		}
	}

	//

	// See if an instance has completed successfully, setting `next_timeout`
	// to a non-zero value if one has.
	next_timeout := 0 // non-zero => "restart with new base"
	for _, instance := range attdata.active_instances.get_instances() {
		if instance.RunState == 1 {
			// Completed successfully, make this instance the new base.
			attdata.current_instance = instance
			base_instance = instance
			attdata.new_current_instance(bdata, instance)
			next_timeout = max(
				(instance.Ticks*attdata.Parameters.NEW_BASE_TIMEOUT_FACTOR)/10,
				attdata.cycle_timeout)
			// next_timeout != 0 and base_instance = current_instance is new

			//TODO? Is this really needed?
			attdata.active_instances.blocked = true
			break
		}
	}

	//TODO: Shouldn't this be testing the active instances?
	// Or does this test belong somewhere else, say in update_queue, after removing
	// the completed instances, etc.?
	if len(attdata.run_queue) == 0 {
		// all current constraint trials finished.
		return true
	}

	// Seek failed instances, which should be split.
	// If there is a new base, stop the old instances and
	// restart them accordingly.
	split_instances := []*TtInstance{}
	new_constraint_list := []*TtInstance{}
	type weighted_instance struct {
		progress int
		instance *TtInstance
	}
	renewed_instances := []weighted_instance{}
	for _, instance := range attdata.active_instances.get_instances() {
		if instance.RunState == 2 { // timed out / failed
			// Split if more than one instance in list
			if len(instance.Constraints) > 1 {
				timeout := next_timeout
				if timeout == 0 {
					// If not rebasing, keep the old "timeout"
					timeout = instance.Timeout
				}
				sit := []string{}
				for _, si := range attdata.split_instance(
					instance, base_instance, timeout) {
					split_instances = append(split_instances, si)
					sit = append(sit,
						fmt.Sprintf("%d:%s", si.Index, si.ConstraintType))
				}
				logger.Info("(SPLIT) %d:%s -> %v",
					instance.Index, instance.ConstraintType, sit)
			} else if len(instance.Constraints) == 1 {
				// Only a single constraint
				//TODO: Save the constraint index with some measure of its
				// progress rate, to allow possible later reinclusion?
				if len(instance.Message) != 0 {
					attdata.ConstraintErrors[instance.Constraints[0]] =
						instance.Message
				}
			} else {
				panic("Bug, expected constraint(s)")
			}
		} else { // completed successfully or still running
			if next_timeout != 0 {
				// There is a new base instance ...
				// Cancel existing instance
				progress := 0
				switch instance.RunState {
				case -2: // already split and aborted, don't build a new instance
					continue
				case -1: // running
					attdata.abort_instance(instance)
					// Get progress because if well advanced, it should be prioritized
					progress = instance.Progress
				case 0: // queued, mark it "don't start"
					instance.RunState = 3
				}
				// Build new instance

				//TODO: Split it if progress was slow (but not slow enough to
				// trigger an Abort)?
				//TODO: Adjust placement in queue according to progress rate?

				instance = attdata.new_instance(
					base_instance,
					instance.ConstraintType,
					instance.Weight,
					instance.Constraints,
					next_timeout)

				if progress > NEARLY_FINISHED {
					renewed_instances = append(renewed_instances,
						weighted_instance{progress, instance})
				} else {
					//TODO???
					rq.add(instance)
				}
			}
			new_constraint_list = append(
				new_constraint_list, instance)
		}
	}
	//TODO: the constraint_instance_list needs rebasing
	attdata.run_queue = nil
	if len(renewed_instances) != 0 {
		slices.SortStableFunc(renewed_instances, func(a, b weighted_instance) int {
			return b.progress - a.progress // highest progress first
		})
		for _, ri := range renewed_instances {
			attdata.run_queue = append(attdata.run_queue,
				ri.instance)
		}
	}
	attdata.run_queue = append(attdata.run_queue,
		new_constraint_list...)
	attdata.run_queue = append(attdata.run_queue,
		split_instances...)

	//TODO???
	for _, instance := range split_instances {
		rq.add(instance)
	}
	return false // still processing
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
