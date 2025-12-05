package autotimetable

import (
	"encoding/json"
	"fetrunner/base"
	"fmt"
	"os"
	"path/filepath"
)

type Result struct {
	Time                       int
	Days                       []IdPair
	Hours                      []IdPair
	Activities                 []IdPair
	Constraints                []Constraint
	ConstraintErrors           map[ConstraintIndex]string
	Rooms                      []IdPair
	Placements                 []TtActivityPlacement
	UnfulfilledHardConstraints map[ConstraintType][]ConstraintIndex
	TotalHardConstraints       int
	UnfulfilledSoftConstraints map[ConstraintType][]ConstraintIndex
	TotalSoftConstraints       int
}

// Get the result of the current instance as a `Result` structure.
// Save as JSON if debugging.
func (attdata *AutoTtData) new_current_instance(
	bdata *base.BaseData, instance *TtInstance,
) {
	bdata.Logger.Info("[%d] <<< %d:%s @ %d, n: %d\n",
		attdata.Ticks, instance.Index, instance.ConstraintType,
		instance.Ticks, len(instance.Constraints))

	// Read placements
	alist := instance.Backend.Results(bdata, attdata, instance)

	// The discarded hard constraints ...
	hnall := 0 // count all hard constraints
	// Gather constraint indexes:
	hunfulfilled := map[ConstraintType][]ConstraintIndex{}
	for ctype, clist := range attdata.HardConstraintMap {
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
	for ctype, clist := range attdata.SoftConstraintMap {
		ulist := []ConstraintIndex{}
		for _, i := range clist {
			if !instance.ConstraintEnabled[i] {
				ulist = append(ulist, i)
			}
		}
		sunfulfilled[ctype] = ulist
		snall += len(clist)
	}
	clist := attdata.Source.GetConstraints()
	rlist := attdata.Source.GetRooms()
	attdata.lastResult = &Result{
		Time:        instance.Ticks,
		Days:        attdata.Source.GetDays(),
		Hours:       attdata.Source.GetHours(),
		Activities:  attdata.Source.GetActivities(),
		Constraints: clist,
		// ConstraintErrors can be updated after this Result is constructed.
		// This allows constraint errors which are detected later to be
		// included, but there may also be spurious timeout messages about
		// constraints which are enabled.
		ConstraintErrors:           attdata.ConstraintErrors,
		Rooms:                      rlist,
		Placements:                 alist,
		UnfulfilledHardConstraints: hunfulfilled,
		TotalHardConstraints:       hnall,
		UnfulfilledSoftConstraints: sunfulfilled,
		TotalSoftConstraints:       snall,
	}
	if attdata.Parameters.DEBUG {
		//b, err := json.Marshal(LastResult)
		b, err := json.MarshalIndent(attdata.lastResult, "", "  ")
		if err != nil {
			panic(err)
		}
		fpath := filepath.Join(bdata.SourceDir,
			fmt.Sprintf("%s_%d.json", instance.ConstraintType, instance.Index))
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

// Get the "result" of the last successful instance as JSON.
func (attdata *AutoTtData) GetLastResult() []byte {
	if attdata.lastResult == nil {
		return nil
	}

	// This will include entries added after the "last result" was recorded
	attdata.lastResult.ConstraintErrors = attdata.ConstraintErrors

	//b, err := json.Marshal(LastResult)
	b, err := json.MarshalIndent(attdata.lastResult, "", "  ")
	if err != nil {
		panic(err)
	}
	return b
}
