package autotimetable

import (
	"fmt"
	"slices"
)

// During phase 0 only `full_instance`, `hard_instance` and
// `null_instance` are running.
func (runqueue *RunQueue) phase0() int {
	basic_data := runqueue.BasicData
	switch basic_data.null_instance.ProcessingState {
	case 0:
		if basic_data.null_instance.Ticks ==
			basic_data.null_instance.Timeout {
			basic_data.abort_instance(basic_data.null_instance)
		}
		return 0

	case 1:
		// The null instance completed successfully.
		basic_data.current_instance = basic_data.null_instance
		basic_data.new_current_instance(basic_data.current_instance)
		// Start trials of single constraint types.
		return 1

	default:
		// The null instance failed.
		basic_data.Logger.Error(
			"[%d] --- Unconstrained instance failed:\n+++\n%s\n---\n",
			basic_data.Ticks, basic_data.null_instance.Message)
		return -1
	}
}

/*
	Main processing phase(s), accumulating constraints.

In this phase generator instances are run which try to add the (as yet not
included) constraints of a single type, with a timeout. A certain number of
these can be run in parallel. If one completes successfully, it is removed
from the constraint list, all the other instances are stopped and the
successful instance is used as the base for a new cycle. Depending on the
time this instance took to complete, the timeout may be increased.

There is some flexibility around the timeouts. If an instance seems to be
progressing too slowly, it can be halted immediately. On the other hand,
if the instance looks like it might complete if given a little more time,
the timeout termination is delayed.

When an instance times out, it is removed from the constraint list. It is
split into two, each with half of the constraints, the new instances being
added to the constraint list and to the end of the run-queue. If the
instance has only one constraint to add, no new instance is started – until
the next cycle with a new base.

When the constraint list is empty, the cycle ends. For the next cycle, the
as yet unincluded constraints are collected again and the timeout is
adjusted if necessary.

Should it come to pass that all the constraints have been added successfully
(unlikely, because it is more likely that the overall timeout will have been
reached or the hard-only or full instance will have completed already), return
`true`, indicating that there are no more constraints to add.
*/
func (runqueue *RunQueue) mainphase() bool {
	basic_data := runqueue.BasicData
	logger := basic_data.Logger
	next_timeout := 0 // non-zero => "restart with new base"
	base_instance := basic_data.current_instance
	if base_instance == nil {
		// Possible with SKIP_HARD option.
		// Note that this instance may not have finished!
		base_instance = basic_data.hard_instance
	}

	// See if an instance has completed successfully.
	for i, instance := range basic_data.constraint_list {
		if instance.ProcessingState == 1 {
			// Completed successfully, make this instance the new base.
			basic_data.current_instance = instance
			base_instance = instance
			basic_data.new_current_instance(instance)
			next_timeout = max(
				(instance.Ticks*basic_data.Parameters.NEW_BASE_TIMEOUT_FACTOR)/10,
				basic_data.cycle_timeout)
			// Remove it from constraint list.
			basic_data.constraint_list = slices.Delete(
				basic_data.constraint_list, i, i+1)

			// next_timeout != 0 and current_instance is new
			break
		}
	}

	if len(basic_data.constraint_list) == 0 {
		// ... all current constraint trials finished.
		// Start trials of remaining constraints, hard then soft,
		// forcing a longer timeout.
		old_ticks := base_instance.Ticks
		if base_instance.ProcessingState != 1 {
			old_ticks = 0
		}
		basic_data.cycle_timeout = (max(basic_data.cycle_timeout,
			old_ticks) * basic_data.Parameters.NEW_CYCLE_TIMEOUT_FACTOR) / 10
		var n int
	rpt:
		switch basic_data.phase {

		case 0:
			logger.Info(
				"[%d] Phase 1 ...\n",
				basic_data.Ticks)
			basic_data.phase = 1
			basic_data.constraint_list, n = basic_data.get_basic_constraints(
				base_instance, false)
			if n == 0 {
				logger.Warning("--HARD: No hard constraints")
				goto rpt
			}

		case 1:
			if basic_data.hard_instance.ProcessingState == 1 {
				logger.Info(
					"[%d] Phase 2 ... <- %s\n",
					basic_data.Ticks, basic_data.hard_instance.Tag)
			} else {
				logger.Info(
					"[%d] Phase 2 ... <- (accumulated instance)\n",
					basic_data.Ticks)
				// The hard-only instance is no longer needed.
				if basic_data.hard_instance.ProcessingState == 0 {
					basic_data.abort_instance(basic_data.hard_instance)
				}
			}
			basic_data.phase = 2
			basic_data.constraint_list, n = basic_data.get_basic_constraints(
				base_instance, true)
			if n == 0 {
				logger.Info("--SOFT: No soft constraints")
				goto rpt
			}

		case 2:
			return true // end of process

		default:
			panic(fmt.Sprintf("Bug, invalid phase: %d", basic_data.phase))
		}
		// Queue instances for running
		for _, bc := range basic_data.constraint_list {
			runqueue.add(bc)
		}
	}

	// Seek failed instances, which should be split.
	// If there is a new base, stop the old instances and
	// restart them accordingly.
	split_instances := []*TtInstance{}
	new_constraint_list := []*TtInstance{}
	for _, instance := range basic_data.constraint_list {
		if instance.ProcessingState == 2 { // timed out / failed
			// Split if more than one instance in list
			if len(instance.Constraints) > 1 {
				timeout := next_timeout
				if timeout == 0 {
					timeout = instance.Timeout
				}

				sit := []string{}
				for _, si := range runqueue.split_instance(
					instance, base_instance, timeout) {
					split_instances = append(split_instances, si)
					sit = append(sit, si.Tag)
				}
				logger.Info("[%d] (SPLIT) %s -> %v\n",
					basic_data.Ticks, instance.Tag, sit)

				//split_instances = append(split_instances,
				//	runqueue.split_instance(
				//		instance, base_instance, timeout)...)

			} else if len(instance.Constraints) == 1 {
				if len(instance.Message) != 0 {
					basic_data.ConstraintErrors[instance.Constraints[0]] =
						instance.Message
				}
			} else {
				panic("Bug, expected constraint(s)")
			}
		} else {
			if next_timeout != 0 {
				// Cancel existing instance
				if instance.ProcessingState == 0 {
					basic_data.abort_instance(instance)
				}
				// Indicate that a queued instance is not to be started
				instance.ProcessingState = 3
				// Build new instance
				instance = basic_data.new_instance(
					base_instance,
					instance.ConstraintType,
					instance.Constraints,
					next_timeout)
				runqueue.add(instance)
			}
			new_constraint_list = append(
				new_constraint_list, instance)
		}
	}
	basic_data.constraint_list = append(new_constraint_list,
		split_instances...)
	for _, instance := range split_instances {
		runqueue.add(instance)
	}
	return false // still processing
}

func (runqueue *RunQueue) split_instance(
	instance *TtInstance, base_instance *TtInstance, timeout int,
) []*TtInstance {
	basic_data := runqueue.BasicData
	nhalf := len(instance.Constraints) / 2
	return []*TtInstance{
		basic_data.new_instance(
			base_instance,
			instance.ConstraintType,
			instance.Constraints[:nhalf],
			timeout),
		basic_data.new_instance(
			base_instance,
			instance.ConstraintType,
			instance.Constraints[nhalf:],
			timeout),
	}
}
