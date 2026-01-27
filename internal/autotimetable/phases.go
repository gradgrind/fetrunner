package autotimetable

import (
	"fetrunner/internal/base"
	"fmt"
	"slices"
	"strconv"
)

// Enter new phase.
func (rq *RunQueue) enter_phase(p int) {
	attdata := rq.AutoTtData

	base_instance := attdata.current_instance
	if base_instance == nil {
		base_instance = attdata.hard_instance
	} else {
		attdata.cycle_timeout = (max(attdata.cycle_timeout,
			attdata.current_instance.Ticks) *
			attdata.Parameters.NEW_PHASE_TIMEOUT_FACTOR) / 10
	}

	attdata.phase = p
	rq.BData.Logger.Result(".PHASE", strconv.Itoa(p))
	var n int
new_phase:
	if p == PHASE_FINISHED {
		return
	}
	// Initialize constraint list.
	attdata.constraint_list, n = attdata.get_basic_constraints(
		base_instance)
	if n == 0 {
		if p == PHASE_SOFT && attdata.full_instance.RunState < 0 {
			// The fully constrained instance is no longer required
			attdata.abort_instance(attdata.full_instance)
		}
		if p == PHASE_HARD && attdata.hard_instance.RunState < 0 {
			// The hard-only instance is no longer required
			attdata.abort_instance(attdata.hard_instance)
		}
		if p == PHASE_BASIC {
			if attdata.null_instance.RunState < 0 {
				// The unconstrained instance is no longer required
				attdata.abort_instance(attdata.null_instance)
			}
			if attdata.na_instance.RunState < 0 {
				// The "na" instance is no longer required
				attdata.abort_instance(attdata.na_instance)
			}
		}
		// Skip to next phase
		p++
		attdata.phase = p
		rq.BData.Logger.Result(".PHASE", strconv.Itoa(p))
		goto new_phase
	}
	// Queue instances for running
	for _, bc := range attdata.constraint_list {
		rq.add(bc)
	}
}

// During the basic phase only `full_instance`, `hard_instance` and
// `null_instance` are running. Return `true` if the phase is changed,
// otherwise `false`.
func (rq *RunQueue) phase_basic() bool {
	attdata := rq.AutoTtData
	if attdata.ticked_hard_only(rq.BData) {
		// All hard constraints OK, skip to trials of soft constraints.
		rq.enter_phase(PHASE_SOFT)
		return true
	}
	if attdata.ticked_na_only(rq.BData) {
		// All "na" constraints OK, skip to trials of hard constraints.
		rq.enter_phase(PHASE_HARD)
		return true
	}
	switch attdata.null_instance.RunState {
	case -1:
		if attdata.null_instance.Ticks ==
			attdata.null_instance.Timeout {
			attdata.abort_instance(attdata.null_instance)
			// The failure will be caught next time round ...
		}
	case 1:
		// The null instance completed successfully.
		attdata.current_instance = attdata.null_instance
		attdata.new_current_instance(rq.BData, attdata.current_instance)
	default:
		// The null instance failed.
		rq.BData.Logger.Error(
			"Unconstrained instance failed:\n:::+\n%s\n:::-",
			attdata.null_instance.Message)
		rq.enter_phase(PHASE_FINISHED)
		return true
	}
	return false
}

// During the "hard" phases, `full_instance`, `hard_instance` and various
// instances adding individual hard constraint types are running. Return
// `true` if the phase is changed, otherwise `false`.
func (rq *RunQueue) phase_hard() bool {
	attdata := rq.AutoTtData
	if attdata.ticked_hard_only(rq.BData) {
		// All hard constraints OK, skip to trials of soft constraints.
		rq.enter_phase(PHASE_SOFT)
		return true
	}
	if rq.mainphase() {
		rq.enter_phase(PHASE_SOFT)
		return true
	}
	return false
}

// During the "soft" phase, `full_instance` and various instances
// adding individual soft constraint types are running. Return `true` if
// the phase is changed, otherwise `false`.
func (rq *RunQueue) phase_soft() bool {
	if rq.mainphase() {
		rq.enter_phase(PHASE_FINISHED)
		return true
	}
	return false
}

/*
	Main processing phase(s), accumulating constraints.

`mainphase()` is run in phases PHASE_NA, PHASE_EXTRAHARD and PHASE_HARD
(adding hard constraints) and also PHASE_SOFT (adding soft constraints).
Generator instances are run which try to add the (as yet not included)
constraints of a single type, with a timeout. A certain number of these can
be run in parallel. If one completes successfully, it is removed from the
constraint list, all the other instances are stopped (including any others
that might have completed successfully) and this successful instance is used
as the base for a new cycle. Depending on the time this instance took to
complete, the timeout may be increased.

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

Should it come to pass that all the constraints (hard and soft) have been
added successfully (unlikely, because it is more likely that the overall
timeout will have been reached or the hard-only or full instance will have
completed already), return `true`, indicating that there are no more
constraints to add. Otherwise (the normal case) return `false`.
*/
func (rq *RunQueue) mainphase() bool {
	attdata := rq.AutoTtData
	logger := rq.BData.Logger
	next_timeout := 0 // non-zero => "restart with new base"
	base_instance := attdata.current_instance
	if base_instance == nil {
		// Possible only with SKIP_HARD option, in which case the instance
		// won't be running, let alone finished!
		base_instance = attdata.hard_instance
	}

	// See if an instance has completed successfully, setting `next_timeout`
	// to a non-zero value if one has.
	for i, instance := range attdata.constraint_list {
		if instance.RunState == 1 {
			// Completed successfully, make this instance the new base.
			attdata.current_instance = instance
			base_instance = instance
			attdata.new_current_instance(rq.BData, instance)
			next_timeout = max(
				(instance.Ticks*attdata.Parameters.NEW_BASE_TIMEOUT_FACTOR)/10,
				attdata.cycle_timeout)
			// Remove it from constraint list.
			attdata.constraint_list = slices.Delete(
				attdata.constraint_list, i, i+1)

			// next_timeout != 0 and base_instance = current_instance is new
			break
		}
	}

	if len(attdata.constraint_list) == 0 {
		// all current constraint trials finished.
		return true
	}

	// Seek failed instances, which should be split.
	// If there is a new base, stop the old instances and
	// restart them accordingly.
	split_instances := []*TtInstance{}
	new_constraint_list := []*TtInstance{}
	for _, instance := range attdata.constraint_list {
		if instance.RunState == 2 { // timed out / failed
			// Split if more than one instance in list
			if len(instance.Constraints) > 1 {
				timeout := next_timeout
				if timeout == 0 {
					timeout = instance.Timeout
				}

				sit := []string{}
				for _, si := range rq.split_instance(
					instance, base_instance, timeout) {
					split_instances = append(split_instances, si)
					sit = append(sit,
						fmt.Sprintf("%d:%s", si.Index, si.ConstraintType))
				}
				logger.Info("(SPLIT) %d:%s -> %v",
					instance.Index, instance.ConstraintType, sit)

				//split_instances = append(split_instances,
				//	runqueue.split_instance(
				//		instance, base_instance, timeout)...)

			} else if len(instance.Constraints) == 1 {
				if len(instance.Message) != 0 {
					attdata.ConstraintErrors[instance.Constraints[0]] =
						instance.Message
				}
			} else {
				panic("Bug, expected constraint(s)")
			}
		} else {
			if next_timeout != 0 {
				// There is a new base instance ...
				// Cancel existing instance
				switch instance.RunState {
				case -2: // already split and aborted, don't build a new instance
					continue
				case -1: // running
					attdata.abort_instance(instance)
				case 0: // queued, mark it "don't start"
					instance.RunState = 3
				}
				// Build new instance
				instance = attdata.new_instance(
					base_instance,
					instance.ConstraintType,
					instance.Weight,
					instance.Constraints,
					next_timeout)
				rq.add(instance)
			}
			new_constraint_list = append(
				new_constraint_list, instance)
		}
	}
	attdata.constraint_list = append(new_constraint_list,
		split_instances...)
	for _, instance := range split_instances {
		rq.add(instance)
	}
	return false // still processing
}

func (rq *RunQueue) split_instance(
	instance *TtInstance, base_instance *TtInstance, timeout int,
) []*TtInstance {
	attdata := rq.AutoTtData
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

// Handle "tick-updates" for the special HARD_ONLY instance.
// This is called in basic and hard phases. Return `true` if the
// HARD_ONLY instance has completed successfully, otherwise `false`.
func (attdata *AutoTtData) ticked_hard_only(bdata *base.BaseData) bool {
	if attdata.hard_instance.RunState == 1 {
		// Set as current and prepare for processing soft constraints.
		attdata.current_instance = attdata.hard_instance
		bdata.Logger.Result(".HARD_OK", "All hard constraints OK")
		attdata.new_current_instance(bdata, attdata.current_instance)
		// Cancel everything except full instance.
		if attdata.null_instance.RunState < 0 {
			attdata.abort_instance(attdata.null_instance)
		}
		if attdata.na_instance != nil && attdata.na_instance.RunState < 0 {
			attdata.abort_instance(attdata.na_instance)
		}
		for _, instance := range attdata.constraint_list {
			if instance.RunState < 0 {
				attdata.abort_instance(instance)
			} else if instance.RunState == 0 {
				// Indicate that a queued instance is not to be started
				instance.RunState = 3
			}
		}
		attdata.constraint_list = nil
		return true
	}
	return false
}

// Handle "tick-updates" for the special NA_ONLY instance.
// This is called in the PHASE_BASIC. Return `true` if the
// NA_ONLY instance has completed successfully, otherwise `false`.
func (attdata *AutoTtData) ticked_na_only(bdata *base.BaseData) bool {
	if attdata.na_instance.RunState == 1 {
		// Set as current.
		attdata.current_instance = attdata.na_instance
		bdata.Logger.Result(".NA_OK", "All hard NotAvailable constraints OK")
		attdata.new_current_instance(bdata, attdata.current_instance)
		// Cancel everything except full and hard-only instances.
		if attdata.null_instance.RunState < 0 {
			attdata.abort_instance(attdata.null_instance)
		}
		for _, instance := range attdata.constraint_list {
			if instance.RunState < 0 {
				attdata.abort_instance(instance)
			} else if instance.RunState == 0 {
				// Indicate that a queued instance is not to be started
				instance.RunState = 3
			}
		}
		attdata.constraint_list = nil
		return true
	}
	return false
}
