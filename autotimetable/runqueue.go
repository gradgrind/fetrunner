package autotimetable

import (
	"fetrunner/base"
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
	instance.ProcessingState = -1 // not started yet
	rq.Queue = append(rq.Queue, instance)
}

func (rq *RunQueue) update_instances() {
	attdata := rq.AutoTtData
	logger := rq.BData.Logger
	// First increment the ticks of active instances.
	for instance := range rq.Active {
		if instance.RunState != 0 && instance.ProcessingState < 2 {
			// This should only be possible after the call to
			// the back-end tick method below.
			panic(fmt.Sprintf("Bug, State = %d", instance.RunState))
		}
		if instance.RunState == 0 {
			instance.Ticks++
			// Among other things, update the state:
			instance.Backend.Tick(rq.BData, attdata, instance)
		} else if instance.ProcessingState < 2 {
			// This should only be possible after the call to
			// the back-end tick method.
			panic(fmt.Sprintf("Bug, State = %d", instance.RunState))
		}

		if instance.ProcessingState == 3 {
			// Await completion of the goroutine
			continue
		}
		switch instance.RunState {
		case 0: // running, not finished
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
						"[%d] Stop (too slow) %d:%s @ %d, p: %d n: %d\n",
						attdata.Ticks,
						instance.Index,
						instance.ConstraintType,
						instance.Ticks,
						instance.Progress,
						len(instance.Constraints))
					attdata.abort_instance(instance)
					if len(instance.Constraints) == 1 {
						attdata.BlockConstraint[instance.Constraints[0]] = true
					}
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
				logger.Info("[%d] Timeout %d:%s @ %d, p: %d n: %d\n",
					attdata.Ticks,
					instance.Index,
					instance.ConstraintType,
					instance.Ticks,
					instance.Progress,
					len(instance.Constraints))
				attdata.abort_instance(instance)
				continue
			}

		case 1: // completed successfully
			//logger.Info("[%d] <<+ %s @ %d (%d)\n + %v\n",
			//	basic_data.Ticks, instance.Tag, instance.Ticks,
			//	len(instance.Constraints), instance.Constraints)
			logger.Info("[%d] <<+ %d:%s @ %d (%d)\n",
				attdata.Ticks,
				instance.Index,
				instance.ConstraintType,
				instance.Ticks,
				len(instance.Constraints))
			instance.ProcessingState = 1

		default: // completed unsuccessfully
			//logger.Info("[%d] <<- %s @ %d (%d)\n + %v\n",
			//	basic_data.Ticks, instance.Tag, instance.Ticks,
			//	len(instance.Constraints), instance.Constraints)
			logger.Info("[%d] <<- %d:%s @ %d (%d)\n",
				attdata.Ticks,
				instance.Index,
				instance.ConstraintType,
				instance.Ticks,
				len(instance.Constraints))
			instance.ProcessingState = 2
		}
	}
}

func (rq *RunQueue) update_queue() int {
	attdata := rq.AutoTtData
	logger := rq.BData.Logger
	// Try to start queued instances
	running := 0
	for instance := range rq.Active {
		if instance.RunState != 0 {
			delete(rq.Active, instance)
			if !attdata.Parameters.DEBUG {
				instance.Backend.Clear()
			}
			continue
		}
		if instance.ProcessingState == 0 {
			running++
		}
	}

	for rq.Next < len(rq.Queue) && running < rq.MaxRunning {
		instance := rq.Queue[rq.Next]
		rq.Queue[rq.Next] = nil
		rq.Next++

		if instance.ProcessingState < 0 {
			//logger.Info("[%d] >> %s n: %d t: %d\n + %v\n",
			//	basic_data.Ticks,
			//	instance.Tag,
			//	len(instance.Constraints),
			//	instance.Timeout,
			//	instance.Constraints)
			logger.Info("[%d] >> %d:%s n: %d t: %d\n",
				attdata.Ticks,
				instance.Index,
				instance.ConstraintType,
				len(instance.Constraints),
				instance.Timeout)
			instance.Backend =
				attdata.BackendInterface.RunBackend(rq.BData, instance)
			instance.ProcessingState = 0 // indicate started/running
			rq.Active[instance] = struct{}{}
			running++
		} else {
			if instance.ProcessingState != 3 {
				panic("Bug")
			}
			// Cancelled before starting, skip it
		}
	}

	// If not all processors are being used, split one or more instances.
	for instance := range rq.Active {
		np := rq.MaxRunning - running
		if np <= 0 {
			break
		}
		if instance.ProcessingState == 0 {
			n := len(instance.Constraints)
			if n <= 1 {
				continue
			}
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
			logger.Info("[%d] (NSPLIT) %d:%s -> %v\n",
				attdata.Ticks, instance.Index, instance.ConstraintType, tags)
		}
	}

	return len(rq.Active)
}
