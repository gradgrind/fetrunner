package autotimetable

import (
	"encoding/json"
	"fetrunner/internal/base"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
)

type Result struct {
	Time                       int
	Days                       []base.ElementBase
	Hours                      []base.ElementBase
	Teachers                   []base.ElementBase
	Classes                    []base.ElementBase
	Rooms                      []base.ElementBase
	Activities                 []*TtActivity
	Constraints                []*TtConstraint
	ConstraintErrors           map[ConstraintIndex]string
	Placements                 []TtActivityPlacement
	UnfulfilledHardConstraints map[ConstraintType][]ConstraintIndex
	TotalHardConstraints       int
	UnfulfilledSoftConstraints map[ConstraintType][]ConstraintIndex
	TotalSoftConstraints       int
}

// Get the result of the current instance as a `Result` structure.
// Save as JSON if debugging.
func (attdata *AutoTtData) new_current_instance() {
	bdata := attdata.BaseData
	logger := bdata.Logger
	instance := attdata.current_instance
	instance.RunState = INSTANCE_ACCEPTED
	logger.Result(".ACCEPT", strconv.Itoa(instance.Index))
	alist := attdata.Backend.Results(logger, instance) // read placements
	clist := attdata.Source.GetConstraints()
	cl_list0 := attdata.Source.GetClasses()
	cl_list := make([]base.ElementBase, len(cl_list0))
	for i, cl := range cl_list0 {
		cl_list[i] = base.ElementBase{Id: cl.Id, Tag: cl.Tag}
	}
	attdata.lastResult = &Result{
		Time:        instance.Ticks,
		Days:        attdata.Source.GetDays(),
		Hours:       attdata.Source.GetHours(),
		Teachers:    attdata.Source.GetTeachers(),
		Classes:     cl_list,
		Rooms:       attdata.Source.GetRooms(),
		Activities:  attdata.Source.GetActivities(),
		Constraints: clist,
		// ConstraintErrors can be updated after this Result is constructed.
		// This allows constraint errors which are detected later to be
		// included, but there may also be spurious timeout messages about
		// constraints which are enabled.
		ConstraintErrors: attdata.ConstraintErrors,
		Placements:       alist,
	}

	attdata.log_nconstraints(instance.ConstraintEnabled)

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

// Count the constraints (fulfilled and total), placing infos
// including the unfulfilled indexes in the "last" result item.
func (attdata *AutoTtData) log_nconstraints(enabled []bool) {
	// Collect the unfulfilled hard constraints ...
	hnall := 0 // count all hard constraints
	hn := 0    // count fulfilled hard constraints
	// Gather constraint indexes:
	hunfulfilled := map[ConstraintType][]ConstraintIndex{}
	for ctype, clist := range attdata.HardConstraintMap {
		ulist := []ConstraintIndex{}
		for _, i := range clist {
			if enabled[i] {
				hn++
			} else {
				ulist = append(ulist, i)
			}
		}
		hunfulfilled[ctype] = ulist
		hnall += len(clist)
	}
	// Collect the unfulfilled soft constraints ...
	snall := 0 // count all soft constraints
	sn := 0    // count fulfilled soft constraints
	// Gather constraint indexes:
	sunfulfilled := map[ConstraintType][]ConstraintIndex{}
	for ctype, clist := range attdata.SoftConstraintMap {
		ulist := []ConstraintIndex{}
		for _, i := range clist {
			if enabled[i] {
				sn++
			} else {
				ulist = append(ulist, i)
			}
		}
		sunfulfilled[ctype] = ulist
		snall += len(clist)
	}
	attdata.BaseData.Logger.Result(".NCONSTRAINTS", fmt.Sprintf("%d.%d.%d.%d",
		hn, hnall, sn, snall))

	if attdata.lastResult != nil {
		attdata.lastResult.UnfulfilledHardConstraints = hunfulfilled
		attdata.lastResult.TotalHardConstraints = hnall
		attdata.lastResult.UnfulfilledSoftConstraints = sunfulfilled
		attdata.lastResult.TotalSoftConstraints = snall
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
