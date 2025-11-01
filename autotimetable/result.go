package autotimetable

import (
	"encoding/json"
	"fetrunner/base"
	"os"
	"path/filepath"
)

type Result struct {
	Time                       int
	Days                       []string
	Hours                      []string
	Activities                 []ActivityId
	Constraints                []TtItem
	ConstraintErrors           map[ConstraintIndex]string
	Rooms                      []TtItem
	Placements                 []ActivityPlacement
	UnfulfilledHardConstraints map[ConstraintType][]ConstraintIndex
	TotalHardConstraints       int
	UnfulfilledSoftConstraints map[ConstraintType][]ConstraintIndex
	TotalSoftConstraints       int
}

type TtItem struct {
	Key  string
	Text string
}

// Get the result of the current instance as a `Result` structure.
// Save as JSON if debugging.
func (basic_data *BasicData) new_current_instance(instance *TtInstance) {
	base.Message.Printf("[%d] <<< %s @ %d, n: %d\n",
		basic_data.Ticks, instance.Tag,
		instance.Ticks, len(instance.Constraints))

	/* TODO--
	cerrors := map[ConstraintIndex]string{}
	for k, v := range basic_data.ConstraintErrors {
		if len(v) != 0 {
			cerrors[k] = v
		}
	}
	*/

	// Read placements
	alist := instance.Backend.Results(basic_data, instance)

	// The discarded hard constraints ...
	hnall := 0 // count all hard constraints
	// Gather constraint indexes:
	hunfulfilled := map[ConstraintType][]ConstraintIndex{}
	for ctype, clist := range basic_data.HardConstraintMap {
		ulist := []ConstraintIndex{}
		for _, i := range clist {
			if !instance.ConstraintEnabled[i] {
				ulist = append(ulist, i)
			}
		}
		hunfulfilled[ctype] = ulist
		hnall += len(clist)
	}
	// The discarded soft constraints ...
	snall := 0 // count all soft constraints
	// Gather constraint indexes:
	sunfulfilled := map[ConstraintType][]ConstraintIndex{}
	for ctype, clist := range basic_data.SoftConstraintMap {
		ulist := []ConstraintIndex{}
		for _, i := range clist {
			if !instance.ConstraintEnabled[i] {
				ulist = append(ulist, i)
			}
		}
		sunfulfilled[ctype] = ulist
		snall += len(clist)
	}
	clist := basic_data.Source.GetConstraintItems()
	rlist := basic_data.Source.GetRooms()
	basic_data.lastResult = &Result{
		Time:        instance.Ticks,
		Days:        basic_data.Source.GetDayTags(),
		Hours:       basic_data.Source.GetHourTags(),
		Activities:  basic_data.Source.GetActivityIds(),
		Constraints: clist,
		//TODO-- ConstraintErrors:           cerrors,
		Rooms:                      rlist,
		Placements:                 alist,
		UnfulfilledHardConstraints: hunfulfilled,
		TotalHardConstraints:       hnall,
		UnfulfilledSoftConstraints: sunfulfilled,
		TotalSoftConstraints:       snall,
	}
	if basic_data.Parameters.DEBUG {
		//b, err := json.Marshal(LastResult)
		b, err := json.MarshalIndent(basic_data.lastResult, "", "  ")
		if err != nil {
			panic(err)
		}
		fpath := filepath.Join(basic_data.WorkingDir, instance.Tag+".json")
		f, err := os.Create(fpath)
		if err != nil {
			panic("Couldn't open output file: " + fpath)
		}
		defer f.Close()
		_, err = f.Write(b)
		if err != nil {
			panic("Couldn't write result to: " + fpath)
		}
	}
}
