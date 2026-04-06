package autotimetable

import (
	"fmt"
)

// Handling the instance queue ...

func (attdata *AutoTtData) set_runqueue(instances []*TtInstance) {
	attdata.run_queue = instances
	attdata.run_queue_next = 0
}

func (attdata *AutoTtData) unqueue_instance() *TtInstance {
	if attdata.run_queue_next < len(attdata.run_queue) {
		tti := attdata.run_queue[attdata.run_queue_next]
		attdata.run_queue_next++
		return tti
	}
	return nil
}

func (attdata *AutoTtData) queue_instance(instance *TtInstance) {
	if attdata.run_queue_next >= 100 {
		// Reclaim space
		vec2 := attdata.run_queue[attdata.run_queue_next:]
		n := len(vec2)
		copy(attdata.run_queue, vec2)
		attdata.run_queue = attdata.run_queue[:n]
		attdata.run_queue_next = 0
	}
	instance.RunState = 0 // not started yet
	attdata.run_queue = append(attdata.run_queue, instance)
}

func (attdata *AutoTtData) n_queued() int {
	return len(attdata.run_queue) - attdata.run_queue_next
}

// ... end of instance queue handling

// Clean up finished/discarded instances, try to start new ones.
// This is called – in the tick-loop – just before waiting for a tick.
func (attdata *AutoTtData) update_queue() int {
	// Count running, still active, instances; remove completed ones.
	running := 0
	insert_index := 0
	for _, instance := range attdata.active_instances {
		if instance.RunState > 0 {
			// This is the final end of this instance. The back-end run must
			// have finished already, and all data from the run must have been
			// collected.
			if !attdata.Parameters.DEBUG {
				instance.InstanceBackend.Clear()
			}
		} else {
			if instance.RunState == INSTANCE_RUNNING {
				running++
			}
			attdata.active_instances[insert_index] = instance
			insert_index++
		}
	}
	attdata.active_instances = attdata.active_instances[0:insert_index]

	// Try to start queued instances
	maxprocesses := attdata.Parameters.MAXPROCESSES
	for running < maxprocesses {
		// Get next pending instance.
		instance := attdata.unqueue_instance()
		if instance == nil {
			goto split
		}
		attdata.start_instance(instance)
		running++
	}
	return running

split:
	//TODO: Consider further splitting. For the moment it is not being done.
	return running

	/*
		// If not all processors are being used, split one or more instances.
		//TODO: This is not terribly neat, it also had a couple of bugs, and may
		// still have some. It should perhaps be replaced by something cleaner.
		// Rapidly progressing instances should perhaps not be split (yet)?

		bdata := attdata.BaseData
		logger := bdata.Logger

		for _, instance := range attdata.active_instances.get_instances() {
			np := maxprocesses - running
			if np <= 0 {
				break
			}
			if instance.RunState == INSTANCE_RUNNING { // instance running, not (yet) split
				if instance.Stopped {
					continue
				}
				n := len(instance.Constraints)
				if n <= 1 {
					continue
				}
				instance.RunState = -2 // mark as split
				//TODO: Split only when starting?
				attdata.abort_instance(instance)
				//TODO: Remove it from constraint list? Is it still in there?
				//attdata.constraint_instance_list.remove(instance)

				// Always assume one more processor, so that one instance
				// will be available in the queue, if possible.
				np++
				// Limit the number of divisions to at most the number of
				// constraints.
				if n < np {
					np = n
				}
				rem := n % np
				ni := n / np
				tags := []string{}
				for range np {
					nx := n
					n -= ni
					if rem != 0 {
						n--
						rem--
					}
					inew := attdata.new_instance(
						instance.BaseInstance,
						instance.ConstraintType,
						instance.Weight,
						instance.Constraints[n:nx],
						instance.Timeout)
					attdata.run_queue.add_end(inew)
					tags = append(tags,
						fmt.Sprintf("%d:%s", inew.Index, inew.ConstraintType))
					running++
				}
				if n != 0 {
					panic("Bug: wrong constraint division ...")
				}
				logger.Info("(NSPLIT) %d:%s -> %v",
					instance.Index, instance.ConstraintType, tags)
			}
		}

		return attdata.active_instances.number()
	*/
}

// TODO ... ??? Much is now in the phase handler
// `update_instances` is called – in the tick-loop – just after receiving a
// tick.
// The `RunState` field is initially 0, which indicates "not started".
// Unstarted instances in the queue are started in `update_queue`, which
// also sets `RunState` to INSTANCE_RUNNING. `RunState` is set to a "finished"
// value – 1 (successful, 100%) or 2 (not successful) TODO? – in the back-end
// tick handler `DoTick()`, called at the beginning of this method, and thus
// also in the tick-loop thread.
func (attdata *AutoTtData) update_instances() {
	bdata := attdata.BaseData
	logger := bdata.Logger
	// First increment the ticks of active instances.
	for _, instance := range attdata.active_instances {
		if instance.RunState < 0 {
			instance.Ticks++
			// Among other things, update the state:
			instance.InstanceBackend.DoTick(attdata, instance)
		}
		switch instance.RunState {

		case -2: // running, awaiting completion after "abort", split

		case -1: // running, not finished
			if instance.Progress == 100 {
				continue // the state will be changed next time round
			}
			// Check for timeout or getting "stuck"

			//TODO: Should this rather be in the back-end as it could be engine-dependent?

			t := instance.Timeout
			if t == 0 {
				// Check for lack of progress for instances with no timeout
				if instance.LastTime < attdata.Parameters.LAST_TIME_0 &&
					instance.Ticks >= attdata.Parameters.LAST_TIME_1 {
					// Stop instance
					logger.Info(
						"Stop (too slow) %d:%s @ %d, p: %d n: %d",
						instance.Index,
						instance.ConstraintType,
						instance.Ticks,
						instance.Progress,
						len(instance.Constraints))
					attdata.abort_instance(instance, ABORT_TIMED_OUT)
					attdata.BlockSingleConstraint(instance, logger)
				}
				continue
			}

			limit := (instance.Ticks * 50) / t
			//TODO: This is not really a timeout! And the multiplier is highly experimental.
			// It's more of a "progress on course" criterion.
			if instance.Progress < limit {
				// Progress is too slow ...
				logger.Info("Timeout %d @ %d, %d%%",
					instance.Index,
					instance.Ticks,
					instance.Progress)
				attdata.abort_instance(instance, ABORT_TIMED_OUT)
				continue
			}

		case 1: // completed successfully

		case 2: // completed unsuccessfully

		case 3: // don't start

		default: // shouldn't be possible
			panic("Impossible instance RunState")
		}
	}
}

func (attdata *AutoTtData) EliminateSingleConstraint(instance *TtInstance) {
	if len(instance.Constraints) == 1 {
		attdata.BaseData.Logger.Result(
			".ELIMINATE", fmt.Sprintf("%s.%d.%s",
				attdata.Backend.ConstraintName(instance),
				instance.Constraints[0],
				instance.Message))
	}
}
