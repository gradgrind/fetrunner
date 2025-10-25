package autotimetable

import (
	"fetrunner/base"
	"slices"
)

// During phase 0 only `full_instance`, `hard_instance` and
// `null_instance` are running.
func (basic_data *BasicData) phase0() int {
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
		base.Message.Printf(
			"(TODO) [%d] Unconstrained instance failed",
			basic_data.Ticks)

		base.Error.Println(" ... " + basic_data.null_instance.Message)

		//TODO: Seek problems in the unconstrained data.
		panic("TODO")
		return -10
	}
}

/*
	Main processing phase(s), accumulating constraints.

In this phase instances are run which try to add the (as yet not included)
constraints of a single type, with a given timeout. A certain number of
these can be run in parallel. If one completes successfully, it is removed
from the constraint list. All the other instances are stopped and the
successful instance is used as the base for a new cycle. Depending on the
time this instance took to complete, the timeout may be increased.

There is some flexibility around the timeouts. If an instance seems to be
progressing too slowly, it can be halted immediately. On the other hand,
if the instance looks like it might complete if given a little more time,
the timeout is delayed.

When an instance times out, it is removed from the constraint list. It is
split into two, each with half of the constraints, the new instances being
added to the constraint list and to the end of the run-queue. If the
instance has only one constraint to add, no new instance is started – until
the next cycle.

When the constraint list is empty, the cycle ends. For the next cycle, the
as yet unincluded constraints are collected again and the timeout is
increased somewhat.

Should it come to pass that all the constraints have been added successfully
(unlikely, because the overall timeout will have been reached or the
hard-only or full instance will have completed already), phase 3 is entered,
indicating that there are no more constraints to add.
*/
func (basic_data *BasicData) mainphase(runqueue *RunQueue) bool {

	next_timeout := 0 // non-zero => "restart with new base"

	// See if an instance has completed successfully.
	for i, instance := range basic_data.constraint_list {
		if instance.ProcessingState == 1 {
			// Completed successfully, make this instance the new base.
			basic_data.current_instance = instance
			basic_data.new_current_instance(instance)
			next_timeout = max(
				instance.Ticks*basic_data.Parameters.NEW_BASE_TIMEOUT_FACTOR/10,
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
		// Start trials of remaining constraints, hard then soft.
		basic_data.cycle_timeout = max(basic_data.cycle_timeout,
			(basic_data.current_instance.Ticks)*
				basic_data.Parameters.NEW_CYCLE_TIMEOUT_FACTOR) / 10
		var n int
	rpt:
		basic_data.constraint_list, n = basic_data.get_basic_constraints(
			basic_data.current_instance, basic_data.phase == 2)
		if n == 0 {
			if basic_data.phase == 2 {
				// The fully constrained instance is no longer needed.
				if basic_data.full_instance.ProcessingState == 0 {
					basic_data.abort_instance(basic_data.full_instance)
				}
				return true // solution found
			} else {
				base.Message.Printf(
					"(TODO) [%d] Phase 2 based on accumulated instance",
					basic_data.Ticks)
				basic_data.phase = 2
				// The hard-only instance is no longer needed.
				if basic_data.hard_instance.ProcessingState == 0 {
					basic_data.abort_instance(basic_data.hard_instance)
				}
				goto rpt
			}
		}
		basic_data.cycle++
		// Queue instances for running
		for _, bc := range basic_data.constraint_list {
			runqueue.add(bc)
		}
		hs := "hard"
		if basic_data.phase == 2 {
			hs = "soft"
		}
		base.Message.Printf(
			"(TODO) [%d] Cycle %d (%s): %d (timeout %d)\n",
			basic_data.Ticks, basic_data.cycle, hs, n, basic_data.cycle_timeout)
		return false // still processing
	}

	// Seek failed instances, which should be split.
	// If there is a new base, stop the old instances and
	// restart them accordingly.
	split_instances := []*TtInstance{}
	new_constraint_list := []*TtInstance{}
	for _, instance := range basic_data.constraint_list {
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
					basic_data.new_instance(
						basic_data.current_instance,
						instance.ConstraintType,
						instance.Constraints[:nhalf],
						timeout))
				split_instances = append(split_instances,
					basic_data.new_instance(
						basic_data.current_instance,
						instance.ConstraintType,
						instance.Constraints[nhalf:],
						timeout))
			} else if len(instance.Constraints) == 0 {
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
					basic_data.current_instance,
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
