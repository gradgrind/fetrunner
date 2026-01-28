package autotimetable

import (
	"fetrunner/internal/base"
	"fmt"
	"slices"
)

type RunQueue struct {
	BData      *base.BaseData
	AutoTtData *AutoTtData
	Queue      []*TtInstance
	Pending    []*TtInstance
	Active     map[*TtInstance]struct{}
	MaxRunning int
	Next       int
}

func (rq *RunQueue) add(instance *TtInstance) {
	if rq.Next >= 100 {
		// Reclaim space
		vec2 := rq.Queue[rq.Next:]
		n := len(vec2)
		copy(rq.Queue, vec2)
		rq.Queue = rq.Queue[:n]
		rq.Next = 0
	}
	//instance.ProcessingState = 0 // not started yet
	rq.Queue = append(rq.Queue, instance)
}

// `update_instances` is called – in the tick-loop – just after receiving a
// tick.
// The `RunState` field is initially 0, which indicates "not started".
// Unstarted instances in the queue are started in `update_queue`, which
// also sets `RunState` to -1. `RunState` is set to a "finished" value, 1
// (successful, 100%) or 2 (not successful), in the back-end tick handler
// `DoTick()`, called at the beginning of this method, and thus also in the
// tick-loop thread.
func (rq *RunQueue) update_instances() {
	attdata := rq.AutoTtData
	logger := rq.BData.Logger
	// First increment the ticks of active instances.
	for instance := range rq.Active {

		//TODO--
		logger.Info("update_instances...%d: %d", instance.Index, instance.RunState)

		if instance.RunState < 0 {
			instance.Ticks++
			// Among other things, update the state:
			instance.Backend.DoTick(rq.BData, attdata, instance)
		}
		switch instance.RunState {

		case -2: // running, awaiting completion after "abort", split

		case -1: // running, not finished
			if instance.Progress == 100 {
				continue // the state will be changed next time round
			}
			// Check for timeout or getting "stuck"
			t := instance.Timeout
			if t == 0 {
				// Check for lack of progress when there is no timeout
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

			limit := (instance.Ticks * 100) / t
			if instance.Progress < limit {
				// Progress is too slow ...
				if instance.Progress*2 > limit {
					// ... but stretch the rule a bit
					continue
				}
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
				attdata.BackendInterface.ConstraintName(instance),
				c,
				instance.Message))
	}
}

// `update_queue` is called – in the tick-loop – just before waiting for a tick.
// Clean up finished/discarded instances, try to start new ones.
func (rq *RunQueue) update_queue() int {
	attdata := rq.AutoTtData
	logger := rq.BData.Logger

	// Count running instances, remove others
	running := 0
	for instance := range rq.Active {
		if instance.RunState < 0 {
			running++
		} else {
			delete(rq.Active, instance)
			if !attdata.Parameters.DEBUG {
				instance.Backend.Clear()
			}
		}
	}

	// Try to start queued instances
	for rq.Next < len(rq.Queue) && running < rq.MaxRunning {
		instance := rq.Queue[rq.Next]
		rq.Queue[rq.Next] = nil
		rq.Next++

		if instance.RunState == 0 {
			instance.Backend =
				attdata.BackendInterface.RunBackend(rq.BData, instance)
			if instance.Backend == nil {
				instance.RunState = 3
			} else {
				instance.RunState = -1 // indicate started/running
				rq.Active[instance] = struct{}{}
				running++
			}
		} else {
			if instance.RunState != 3 {
				panic("Bug")
			}
			// Cancelled before starting, skip it
		}
	}

	// If not all processors are being used, split one or more instances.
	//TODO: This is not terribly neat, it also had a couple of bugs, and may
	// still have some. It should perhaps be replaced by something cleaner.
	for instance := range rq.Active {
		np := rq.MaxRunning - running
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
			// Remove it from constraint list.
			attdata.constraint_list = slices.DeleteFunc(
				attdata.constraint_list, func(i *TtInstance) bool {
					return i == instance
				})

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
				attdata.constraint_list = append(
					attdata.constraint_list, inew)
				rq.add(inew)
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

	return len(rq.Active)
}
