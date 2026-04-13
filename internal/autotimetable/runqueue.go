package autotimetable

import (
	"fmt"
)

// Handling the instance queue ...

func (attdata *AutoTtData) set_runqueue(instances []*TtInstance) {
	attdata.run_queue = instances
	attdata.run_queue_next = 0
}

func (attdata *AutoTtData) get_runqueue() []*TtInstance {
	return attdata.run_queue[attdata.run_queue_next:]
}

func (attdata *AutoTtData) queue_instance(instance *TtInstance) {
	if attdata.run_queue_next >= 50 {
		// Reclaim space
		vec2 := attdata.run_queue[attdata.run_queue_next:]
		n := len(vec2)
		copy(attdata.run_queue, vec2)
		attdata.run_queue = attdata.run_queue[:n]
		attdata.run_queue_next = 0
	}
	attdata.run_queue = append(attdata.run_queue, instance)
}

func (attdata *AutoTtData) n_queued() int {
	return len(attdata.run_queue) - attdata.run_queue_next
}

// ... end of instance queue handling

// Clean up finished/discarded instances, try to start new ones.
// This is called – in the tick-loop – just before waiting for a tick.
func (attdata *AutoTtData) update_queue() int {
	// Count running, still active, instances. Remove completed ones.
	running := 0
	insert_index := 0
	for _, instance := range attdata.active_instances {
		if instance.Finished {
			// This is the final end of this instance, it is removed from
			// the "active" list.
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
		// Start next pending instance.
		if !attdata.start_queued_instance() {
			goto split
		}
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

// Get an instance from the start of the queue, use `attdata.current_instance`
// for the base. Rebase if necessary and decide whether to split.
func (attdata *AutoTtData) start_queued_instance() bool {
	if attdata.run_queue_next >= len(attdata.run_queue) {
		return false
	}
	instance_0 := attdata.run_queue[attdata.run_queue_next]
	// It is possible for `attdata.current_instance` to be `nil`, at the
	// beginning of a run. In this case only "fresh" instances should be in the queue.
	instance_base := attdata.current_instance
	// Single-constraint instances always have no timeout
	if len(instance_0.Constraints) > 1 {
		switch instance_0.RunState {
		case 0, INSTANCE_SUCCESSFUL: // just rebase, if necessary
		case INSTANCE_FAILED, ABORT_TIMED_OUT:
			//TODO-- attdata.BaseData.Logger.Info("§1 %d %d", instance_0.Index, instance_0.RunState)

			attdata.split()
			return true
		case ABORT_NEW_CYCLE:
			//TODO: Do I really want to split here if progress was slow?
			// Is this the right criterion?
			// Split if running long enough and not got very far.
			if instance_0.Ticks >= 10 && instance_0.Progress < NEARLY_FINISHED {
				//TODO-- attdata.BaseData.Logger.Info("§2 %d %d", instance_0.Index, instance_0.RunState)

				attdata.split()
				return true
			}
			// otherwise just rebase
		default:
			panic(fmt.Sprintf("Unexpected RunState, instance %d: %d",
				instance_0.Index, instance_0.RunState))
		}
	}
	if instance_base != nil && instance_0.BaseInstance != instance_base {
		attdata.start_instance(&TtInstance{
			// Make a new `TtInstance`
			Timeout:      attdata.cycle_timeout,
			BaseInstance: instance_base,

			ConstraintType: instance_0.ConstraintType,
			Constraints:    instance_0.Constraints,
			Weight:         instance_0.Weight,
		})
	} else {
		attdata.start_instance(instance_0)
	}
	attdata.run_queue_next++
	return true
}

// Split the first instance in the run queue, starting the first half.
func (attdata *AutoTtData) split() {
	instance := attdata.run_queue[attdata.run_queue_next]
	attdata.BaseData.Logger.Info("(SPLIT) %d:%s (%d)",
		instance.Index, instance.ConstraintType, instance.RunState)

	/*/TODO--
	ilist := []string{}
	for _, ii := range attdata.get_runqueue() {
		ilist = append(ilist, fmt.Sprintf("%d:%d", ii.Index, ii.RunState))
	}
	attdata.BaseData.Logger.Info("§QUEUE1 %+v\n", strings.Join(ilist, ", "))
	alist := []string{}
	for _, ii := range attdata.active_instances {
		alist = append(alist, fmt.Sprintf("%d:%d", ii.Index, ii.RunState))
	}
	attdata.BaseData.Logger.Info("§ACTIVE1 %+v\n", strings.Join(alist, ", "))
	*/

	instance.RunState = INSTANCE_ABANDONED
	// Split, and start the first half
	nhalf := len(instance.Constraints) / 2
	attdata.start_instance(&TtInstance{
		// Make a new `TtInstance`
		Timeout:      attdata.cycle_timeout,
		BaseInstance: attdata.current_instance,

		ConstraintType: instance.ConstraintType,
		Constraints:    instance.Constraints[:nhalf],
		Weight:         instance.Weight,
	})
	// The second half replaces the original queue item.
	attdata.run_queue[attdata.run_queue_next] = &TtInstance{
		// Make a new `TtInstance`
		Timeout:      attdata.cycle_timeout,
		BaseInstance: attdata.current_instance,

		ConstraintType: instance.ConstraintType,
		Constraints:    instance.Constraints[nhalf:],
		Weight:         instance.Weight,
	}

	/*/TODO--
	ilist = []string{}
	for _, ii := range attdata.get_runqueue() {
		ilist = append(ilist, fmt.Sprintf("%d:%d", ii.Index, ii.RunState))
	}
	attdata.BaseData.Logger.Info("§QUEUE2 %+v\n", strings.Join(ilist, ", "))
	alist = []string{}
	for _, ii := range attdata.active_instances {
		alist = append(alist, fmt.Sprintf("%d:%d", ii.Index, ii.RunState))
	}
	attdata.BaseData.Logger.Info("§ACTIVE2 %+v\n", strings.Join(alist, ", "))
	*/
}
