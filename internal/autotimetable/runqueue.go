package autotimetable

import (
	"fetrunner/internal/base"
	"fmt"
	"slices"
)

// TODO: Would a simple slice be good enough?
// This is using a linked list for the run-queue instances. The links are in the
// instances themselves.
type run_queue struct {
	first         *TtInstance
	last          *TtInstance
	next_instance *TtInstance // next instance in `constraint_instance_list`
}

func (rq *run_queue) add_end(i *TtInstance) {
	l := rq.last
	if l == nil {
		rq.first = i
	} else {
		rq.last.list_next = i
	}
	rq.last = i
	i.list_previous = l
	i.list_next = nil
}

func (rq *run_queue) remove(i *TtInstance) {
	p := i.list_previous
	n := i.list_next
	if p == nil {
		// This should be the first in the list.
		if i != rq.first {
			panic("Attempt to remove TtInstance from list – instance not in list")
		}
		rq.first = n
	} else {
		p.list_next = n
	}
	if n == nil {
		// This is the last in the list
		rq.last = p
	}
	i.list_next = nil
	i.list_previous = nil
}

func (rq *run_queue) get_next() *TtInstance {
	n := rq.next_instance
	if n != nil {
		rq.next_instance = n.list_next
		rq.remove(n)
	}
	return n
}

func (rq *run_queue) add_multiple(instances []*TtInstance) {
	for _, i := range instances {
		rq.add_end(i)
	}
}

func (rq *run_queue) clear() {
	n := rq.first
	for n != nil {
		n.list_next = nil
		n.list_previous = nil
	}
	rq.first = nil
	rq.last = nil
	rq.next_instance = nil
}

// ---------------------------------

type active_instance_set struct {
	instances []*TtInstance
	blocked   bool // not accepting new instances
}

func (ais *active_instance_set) get_instances() []*TtInstance {
	return ais.instances
}

func (ais *active_instance_set) number() int {
	return len(ais.instances)
}

func (ais *active_instance_set) remove(i *TtInstance) {
	ais.instances = slices.DeleteFunc(ais.instances, func(tti *TtInstance) bool {
		return tti == i
	})
}

func (ais *active_instance_set) add(i *TtInstance) {
	ais.instances = append(ais.instances, i)
}

//

/* TODO--
A new instance is not started immediately. First it is placed at the end the the queue
(the `Queue` field). This can, in principle, grow indefinitely, but when an instance is
started it is removed from the front of the queue, leaving a gap. The first entry in
the queue is at the index stored in the `Next` field. When `Next` reaches a certain
value, all the queue entries are moved up to the beginning of the `Queue` slice. Maybe
a circular buffer would be better, but that would need a fixed size, or a more
complicated growing algorithm.
* /
type RunQueue struct {
	AutoTtData   *AutoTtData              // convenient access to autotimetable data
	Queue        []*TtInstance            // pre-start buffer
	Active       map[*TtInstance]struct{} // set of running instances
	MaxProcesses int                      // maximum number of running processes
	NextInstance int                      // index of next instance in `Queue`
}

// Add an instance to the queue, compacting the queue buffer if necessary.
func (rq *RunQueue) add(instance *TtInstance) {
	if rq.NextInstance >= 100 {
		// Reclaim space
		vec2 := rq.Queue[rq.NextInstance:]
		n := len(vec2)
		copy(rq.Queue, vec2)
		rq.Queue = rq.Queue[:n]
		rq.NextInstance = 0
	}
	instance.RunState = 0 // not started yet
	rq.Queue = append(rq.Queue, instance)
}
*/

// `update_instances` is called – in the tick-loop – just after receiving a
// tick.
// The `RunState` field is initially 0, which indicates "not started".
// Unstarted instances in the queue are started in `update_queue`, which
// also sets `RunState` to -1. `RunState` is set to a "finished" value – 1
// (successful, 100%) or 2 (not successful) – in the back-end tick handler
// `DoTick()`, called at the beginning of this method, and thus also in the
// tick-loop thread.
func (attdata *AutoTtData) update_instances() {
	bdata := attdata.BaseData
	logger := bdata.Logger
	// First increment the ticks of active instances.
	for _, instance := range attdata.active_instances.get_instances() {
		if instance.RunState < 0 {
			instance.Ticks++
			// Among other things, update the state:
			instance.InstanceBackend.DoTick(bdata, attdata, instance)
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
					attdata.abort_instance(instance)
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
				attdata.abort_instance(instance)
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

func (attdata *AutoTtData) BlockSingleConstraint(
	instance *TtInstance,
	logger *base.Logger,
) {
	if len(instance.Constraints) == 1 {
		c := instance.Constraints[0]
		attdata.BlockConstraint[c] = true
		logger.Result(
			".ELIMINATE", fmt.Sprintf("%s.%d.%s",
				attdata.Backend.ConstraintName(instance),
				c,
				instance.Message))
	}
}

// `update_queue` is called – in the tick-loop – just before waiting for a tick.
// Clean up finished/discarded instances, try to start new ones.
func (attdata *AutoTtData) update_queue() int {
	bdata := attdata.BaseData
	logger := bdata.Logger

	// Count running instances, remove others
	running := 0
	for _, instance := range attdata.active_instances.get_instances() {
		if instance.RunState < 0 {
			running++
		} else {
			// This is the final end of this instance. The FET run must have
			// finished already, and all back-end data from the run must have been
			// collected.
			attdata.active_instances.remove(instance)
			if !attdata.Parameters.DEBUG {
				instance.InstanceBackend.Clear()
			}
		}
	}

	// Try to start queued instances
	if attdata.active_instances.blocked {
		return running
	}
	maxprocesses := attdata.Parameters.MAXPROCESSES
	for running < maxprocesses {
		instance := attdata.constraint_instance_list.get_next()
		if instance == nil {
			goto split
		}
		attdata.Backend.RunBackend(attdata, instance)
		instance.RunState = -1 // indicate started/running
		attdata.active_instances.add(instance)
		running++
	}
	return running

split:
	// If not all processors are being used, split one or more instances.
	//TODO: This is not terribly neat, it also had a couple of bugs, and may
	// still have some. It should perhaps be replaced by something cleaner.
	// Rapidly progressing instances should perhaps not be split (yet)?
	for _, instance := range attdata.active_instances.get_instances() {
		np := maxprocesses - running
		if np <= 0 {
			break
		}
		if instance.RunState == -1 { // instance running, not (yet) split
			if instance.Stopped {
				continue
			}
			n := len(instance.Constraints)
			if n <= 1 {
				continue
			}
			instance.RunState = -2 // mark as split
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
				attdata.constraint_instance_list.add_end(inew)
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
}
