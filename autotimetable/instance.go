package autotimetable

import (
	"fmt"
	"slices"
)

func (basic_data *BasicData) abort_instance(instance *TtInstance) {
	if !instance.Stopped {
		instance.Backend.Abort()
		instance.Stopped = true
	}
}

func (basic_data *BasicData) new_instance(
	instance_0 *TtInstance,
	constraint_type ConstraintType,
	constraint_indexes []ConstraintIndex,
	timeout int,
) *TtInstance {
	enabled := slices.Clone(instance_0.ConstraintEnabled)
	// Add the new constraints
	for _, c := range constraint_indexes {
		enabled[c] = true
	}

	// Make a new `TtInstance`
	basic_data.instanceCounter++
	instance := &TtInstance{
		Tag: fmt.Sprintf("z%05d~%s",
			basic_data.instanceCounter, constraint_type),
		Timeout:      timeout,
		BaseInstance: instance_0,

		ConstraintEnabled: enabled,
		ConstraintType:    constraint_type,
		Constraints:       constraint_indexes,

		// Run time
		//BackEndData     any
		Ticks:   0,
		Stopped: false,
		//ProcessingState int
	}
	return instance
}
