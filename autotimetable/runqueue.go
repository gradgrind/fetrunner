package autotimetable

import (
	"fetrunner/base"
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
			instance.Backend.Tick(rq.BasicData, instance)
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
				if instance.LastTime < rq.BasicData.Parameters.LAST_TIME_0 &&
					instance.Ticks >= rq.BasicData.Parameters.LAST_TIME_1 {
					// Stop instance
					base.Message.Printf(
						"[%d] Stop (too slow) %s @ %d, p: %d n: %d\n",
						rq.BasicData.Ticks,
						instance.Tag,
						instance.Ticks,
						instance.Progress,
						len(instance.Constraints))
					rq.BasicData.abort_instance(instance)
					if len(instance.Constraints) == 1 {
						rq.BasicData.BlockConstraint[instance.Constraints[0]] = true
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

				//TODO-- instance.TimedOut = true
				if len(instance.Constraints) == 1 {
					instance.Timeout = 0
				} else {
					base.Message.Printf("[%d] Timeout %s @ %d, p: %d n: %d\n",
						rq.BasicData.Ticks,
						instance.Tag,
						instance.Ticks,
						instance.Progress,
						len(instance.Constraints))
					rq.BasicData.abort_instance(instance)
				}
				continue
			}

		case 1: // completed successfully
			base.Message.Printf("[%d] <<+ %s @ %d (%d)\n + %v\n",
				rq.BasicData.Ticks, instance.Tag, instance.Ticks,
				len(instance.Constraints), instance.Constraints)
			instance.ProcessingState = 1

		default: // completed unsuccessfully
			base.Message.Printf("[%d] <<- %s @ %d (%d)\n + %v\n",
				rq.BasicData.Ticks, instance.Tag, instance.Ticks,
				len(instance.Constraints), instance.Constraints)
			instance.ProcessingState = 2
		}
	}
}

func (rq *RunQueue) update_queue() int {
	// Try to start queued instances
	running := 0
	timed_out := []*TtInstance{}
	for instance := range rq.Active {
		if instance.RunState != 0 {
			delete(rq.Active, instance)
			if !rq.BasicData.Parameters.DEBUG {
				instance.Backend.Clear()
			}
			continue
		}
		if instance.ProcessingState == 0 {

			/* TODO--
			if instance.TimedOut {
				if len(instance.Constraints) == 1 {
					timed_out = append(timed_out, instance)
				} else {
					base.Message.Printf("[%d] Timeout %s @ %d, p: %d n: %d\n",
						rq.BasicData.Ticks,
						instance.Tag,
						instance.Ticks,
						instance.Progress,
						len(instance.Constraints))
					rq.BasicData.abort_instance(instance)
				}
			} else {
				running++
			}
			*/

			running++
		}
	}

	for rq.Next < len(rq.Queue) && running < rq.MaxRunning {
		instance := rq.Queue[rq.Next]
		rq.Queue[rq.Next] = nil
		rq.Next++

		if instance.ProcessingState < 0 {
			base.Message.Printf("[%d] >> %s n: %d t: %d\n + %v\n",
				rq.BasicData.Ticks,
				instance.Tag,
				len(instance.Constraints),
				instance.Timeout,
				instance.Constraints)
			instance.Backend =
				rq.BasicData.BackendInterface.RunBackend(instance)
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

	// If not all processors are being used, allow one or more timed-out
	// single-constraint instances to continue running.
	for _, instance := range timed_out {
		if running < rq.MaxRunning {
			//if false {
			running++
		} else {
			base.Message.Printf("[%d] Timeout %s @ %d, p:%d n: 1\n",
				rq.BasicData.Ticks,
				instance.Tag,
				instance.Ticks,
				instance.Progress)
			rq.BasicData.abort_instance(instance)
		}
	}

	// If still not all processors are being used, split one or more
	// instances.
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
			rq.BasicData.abort_instance(instance)
			// Remove it from constraint list.
			rq.BasicData.constraint_list = slices.DeleteFunc(
				rq.BasicData.constraint_list, func(i *TtInstance) bool {
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
				inew := rq.BasicData.new_instance(
					instance.BaseInstance,
					instance.ConstraintType,
					instance.Constraints[n:nx],
					instance.Timeout)
				rq.BasicData.constraint_list = append(
					rq.BasicData.constraint_list, inew)
				rq.add(inew)
				tags = append(tags, inew.Tag)
				running++
			}
			if n != 0 {
				panic("Bug: wrong constraint division ...")
			}
			base.Message.Printf("??? NSPLIT %s -> %v\n",
				instance.Tag, tags)
		}
	}

	return len(rq.Active)
}
