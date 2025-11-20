package autotimetable

import (
	"fmt"
	"slices"
)

type RunQueue struct {
	BasicData  *BasicData
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
	basic_data := rq.BasicData
	logger := basic_data.Logger
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
			instance.Backend.Tick(basic_data, instance)
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
				if instance.LastTime < basic_data.Parameters.LAST_TIME_0 &&
					instance.Ticks >= basic_data.Parameters.LAST_TIME_1 {
					// Stop instance
					logger.Info(
						"[%d] Stop (too slow) %s @ %d, p: %d n: %d\n",
						basic_data.Ticks,
						instance.Tag,
						instance.Ticks,
						instance.Progress,
						len(instance.Constraints))
					basic_data.abort_instance(instance)
					if len(instance.Constraints) == 1 {
						basic_data.BlockConstraint[instance.Constraints[0]] = true
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
				logger.Info("[%d] Timeout %s @ %d, p: %d n: %d\n",
					basic_data.Ticks,
					instance.Tag,
					instance.Ticks,
					instance.Progress,
					len(instance.Constraints))
				basic_data.abort_instance(instance)
				continue
			}

		case 1: // completed successfully
			//logger.Info("[%d] <<+ %s @ %d (%d)\n + %v\n",
			//	basic_data.Ticks, instance.Tag, instance.Ticks,
			//	len(instance.Constraints), instance.Constraints)
			logger.Info("[%d] <<+ %s @ %d (%d)\n",
				basic_data.Ticks, instance.Tag, instance.Ticks,
				len(instance.Constraints))
			instance.ProcessingState = 1

		default: // completed unsuccessfully
			//logger.Info("[%d] <<- %s @ %d (%d)\n + %v\n",
			//	basic_data.Ticks, instance.Tag, instance.Ticks,
			//	len(instance.Constraints), instance.Constraints)
			logger.Info("[%d] <<- %s @ %d (%d)\n",
				basic_data.Ticks, instance.Tag, instance.Ticks,
				len(instance.Constraints))
			instance.ProcessingState = 2
		}
	}
}

func (rq *RunQueue) update_queue() int {
	basic_data := rq.BasicData
	logger := basic_data.Logger
	// Try to start queued instances
	running := 0
	for instance := range rq.Active {
		if instance.RunState != 0 {
			delete(rq.Active, instance)
			if !basic_data.Parameters.DEBUG {
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
			logger.Info("[%d] >> %s n: %d t: %d\n",
				basic_data.Ticks,
				instance.Tag,
				len(instance.Constraints),
				instance.Timeout)
			instance.Backend =
				basic_data.BackendInterface.RunBackend(instance)
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
			basic_data.abort_instance(instance)
			// Remove it from constraint list.
			basic_data.constraint_list = slices.DeleteFunc(
				basic_data.constraint_list, func(i *TtInstance) bool {
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
				inew := basic_data.new_instance(
					instance.BaseInstance,
					instance.ConstraintType,
					instance.Constraints[n:nx],
					instance.Timeout)
				basic_data.constraint_list = append(
					basic_data.constraint_list, inew)
				rq.add(inew)
				tags = append(tags, inew.Tag)
				running++
			}
			if n != 0 {
				panic("Bug: wrong constraint division ...")
			}
			logger.Info("[%d] (NSPLIT) %s -> %v\n",
				basic_data.Ticks, instance.Tag, tags)
		}
	}

	return len(rq.Active)
}
