package autotimetable

import (
	"fetrunner/base"
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
		base.Message.Printf(
			"(TODO) [%d] Unconstrained instance failed",
			basic_data.Ticks)
		base.Error.Println(" ... " + basic_data.null_instance.Message)
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
func (runqueue *RunQueue) mainphase() {
	basic_data := runqueue.BasicData

	//TODO--
	nc := 0
	for _, c := range basic_data.constraint_list {
		nc += len(c.Constraints)
	}
	base.Message.Printf(
		"??? [%d] Cycle %d: %d\n",
		basic_data.Ticks, basic_data.cycle, nc)

	// See if an instance has completed successfully, to be used as the new
	// base.
	// Also look for failed instances: those with only one constraint are
	// placed in the `Runqueue.Pending` list of instances. The others are
	// added to the `to_split_instances` list – these will later be split
	// into two halves and added to the end of the run queue.
	var new_base *TtInstance = nil
	to_split_instances := []*TtInstance{}
	for i, instance := range basic_data.constraint_list {
		switch instance.ProcessingState {

		case 1:
			if new_base == nil {
				new_base = instance
				// Mark for removal from constraint list.
				basic_data.constraint_list[i] = nil
			}

		case 2:
			// Timed out / failed, mark for removal from constraint list.
			basic_data.constraint_list[i] = nil

			if len(instance.Constraints) > 1 {
				to_split_instances = append(to_split_instances, instance)
			} else if len(instance.Constraints) == 1 {
				runqueue.Pending = append(runqueue.Pending, instance)
				c := instance.Constraints[0]
				basic_data.ConstraintErrors[c] = instance.Message
			} else {
				panic("Bug, expected constraint(s)")
			}
		}
	}
	new_constraint_list := []*TtInstance{}
	for _, c := range basic_data.constraint_list {
		if c != nil {
			new_constraint_list = append(new_constraint_list, c)
		}
	}
	basic_data.constraint_list = new_constraint_list

	// If there is a new base instance, all instances in `constraint_list`
	// which are still running need to be stopped. They are duplicated and
	// added to a new constraint list and run queue.
	next_timeout := 0
	if new_base != nil {
		basic_data.current_instance = new_base
		basic_data.new_current_instance(new_base)
		next_timeout = (new_base.Ticks *
			basic_data.Parameters.NEW_BASE_TIMEOUT_FACTOR) / 10
		new_constraint_list = []*TtInstance{}
		for _, instance := range basic_data.constraint_list {
			if instance.ProcessingState == 0 {
				basic_data.abort_instance(instance)
				// Build new instance
				new := basic_data.new_instance(
					basic_data.current_instance,
					instance.ConstraintType,
					instance.Constraints,
					next_timeout)
				new_constraint_list = append(new_constraint_list, new)
				runqueue.add(new)
			}
		}
		basic_data.constraint_list = new_constraint_list
	}

	// Split the instances in `split_instances`, add the new instances to
	// the constraint list and run queue.
	for _, instance := range to_split_instances {
		// Split into two. The new timeout is the same as the old one
		// if `next_timeout` is 0 (not starting with a new base).
		timeout := next_timeout
		if timeout == 0 {
			timeout = instance.Timeout
		}
		nhalf := len(instance.Constraints) / 2
		new := basic_data.new_instance(
			basic_data.current_instance,
			instance.ConstraintType,
			instance.Constraints[:nhalf],
			timeout)
		basic_data.constraint_list = append(basic_data.constraint_list, new)
		runqueue.add(new)
		new = basic_data.new_instance(
			basic_data.current_instance,
			instance.ConstraintType,
			instance.Constraints[nhalf:],
			timeout)
		basic_data.constraint_list = append(basic_data.constraint_list, new)
		runqueue.add(new)
	}

	// If a new cycle is being started, also reactivate the single-constraint
	// instances in `runqueue.Pending`.
	if next_timeout != 0 {
		for _, instance := range runqueue.Pending {
			new := basic_data.new_instance(
				basic_data.current_instance,
				instance.ConstraintType,
				instance.Constraints,
				next_timeout)
			basic_data.constraint_list = append(
				basic_data.constraint_list, new)
			runqueue.add(new)
		}
		runqueue.Pending = nil

		basic_data.cycle++
		hs := "hard"
		if basic_data.phase == 2 {
			hs = "soft"
		}
		nc := 0
		for _, c := range basic_data.constraint_list {
			nc += len(c.Constraints)
		}
		base.Message.Printf(
			"(TODO) [%d] Cycle %d (%s): %d (timeout %d)\n",
			basic_data.Ticks, basic_data.cycle, hs, nc, next_timeout)
	}

	//TODO: Handle the case where all instances time out without any successes.
	// It may be that the extension of timeouts when there are no other
	// processes is often enough, but if there are non-active pending
	// instances there may be better solutions?
}

func (basic_data *BasicData) mainphase_0(runqueue *RunQueue) bool {

	next_timeout := 0 // non-zero => "restart with new base"

	// See if an instance has completed successfully.
	for i, instance := range basic_data.constraint_list {
		if instance.ProcessingState == 1 {
			// Completed successfully, make this instance the new base.
			basic_data.current_instance = instance
			basic_data.new_current_instance(instance)

			next_timeout = (instance.Ticks *
				basic_data.Parameters.NEW_BASE_TIMEOUT_FACTOR) / 10

			if next_timeout < basic_data.cycle_timeout {
				next_timeout = basic_data.cycle_timeout
			}

			//--next_timeout = max(
			//--	instance.Ticks*basic_data.Parameters.NEW_BASE_TIMEOUT_FACTOR/10,
			//--	basic_data.cycle_timeout)
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
		basic_data.cycle_timeout = (max(basic_data.cycle_timeout,
			basic_data.current_instance.Ticks) *
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
