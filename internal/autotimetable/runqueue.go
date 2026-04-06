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

func (attdata *AutoTtData) EliminateSingleConstraint(instance *TtInstance) {
	if len(instance.Constraints) == 1 {
		attdata.BaseData.Logger.Result(
			".ELIMINATE", fmt.Sprintf("%s.%d.%s",
				attdata.Backend.ConstraintName(instance),
				instance.Constraints[0],
				instance.Message))
	}
}
