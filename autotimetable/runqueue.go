package autotimetable

import (
	"fetrunner/base"
	"fmt"
)

type RunQueue struct {
	BasicData  *BasicData
	Queue      []*TtInstance
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
	//base.Message.Printf("(TODO) [%d] Queue %s\n",
	//	Ticks, instance.TtData.Description)
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
					rq.BasicData.abort_instance(instance)
				}
				continue
			}

			limit := (instance.Ticks * 100) / t
			if instance.Progress < limit {
				// Progress is too slow ...
				if instance.Progress*2 > limit {
					continue
				}
				base.Message.Printf("(TODO) [%d] Trap %s @ %d (%d): %d\n",
					rq.BasicData.Ticks,
					instance.Tag,
					instance.Ticks,
					instance.Progress,
					len(instance.Constraints))
				rq.BasicData.abort_instance(instance)
			}

		case 1: // completed successfully
			base.Message.Printf("(TODO) [%d] <<+ %s @ %d\n",
				rq.BasicData.Ticks, instance.Tag, instance.Ticks)
			instance.ProcessingState = 1

		default: // completed unsuccessfully
			base.Message.Printf("(TODO) [%d] <<- %s @ %d\n",
				rq.BasicData.Ticks, instance.Tag, instance.Ticks)
			instance.ProcessingState = 2
		}
	}
}

func (rq *RunQueue) update_queue() int {
	// Try to start queued instances
	running := 0
	for instance := range rq.Active {
		if instance.RunState != 0 {
			delete(rq.Active, instance)
			if !rq.BasicData.Parameters.DEBUG {
				instance.Backend.Clear()
			}
			continue
		}
		if instance.ProcessingState == 0 || instance.ProcessingState == 3 {
			running++
		}
	}
	for rq.Next < len(rq.Queue) && running < rq.MaxRunning {
		instance := rq.Queue[rq.Next]
		rq.Queue[rq.Next] = nil
		rq.Next++

		if instance.ProcessingState < 0 {
			base.Message.Printf("(TODO) [%d] >> %s {%d}\n",
				rq.BasicData.Ticks, instance.Tag, instance.Timeout)
			instance.Backend = rq.BasicData.RunBackend(rq.BasicData, instance)
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
	return len(rq.Active)
}
