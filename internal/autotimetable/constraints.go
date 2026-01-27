package autotimetable

import (
	"slices"
	"strings"
)

type weighted_constraint_list struct {
	weight  string
	ctype   ConstraintType
	indexes []ConstraintIndex
}

// Collect the individual constraints by type, including only those which
// are disabled. Hard constraints are ordered as in the `ConstraintTypes`
// list. Soft constraints are ordered according to weight.
func (attdata *AutoTtData) get_basic_constraints(
	instance0 *TtInstance,
) ([]*TtInstance, int) {
	instances := []*TtInstance{} // one instance per constraint type
	nconstraints := 0            // count constraints
	wlist := []weighted_constraint_list{}
	p := attdata.phase
	switch p {
	case PHASE_SOFT:

		for k, v := range attdata.SoftConstraintMap {
			w, c, ok := strings.Cut(k, ":")
			if !ok {
				panic("Bug: Soft constraint has no weight tag: " + k)
			}
			wlist = append(wlist, weighted_constraint_list{w, c, v})
		}
		slices.SortFunc(wlist, func(a, b weighted_constraint_list) int {
			return strings.Compare(a.weight, b.weight)
		})

	case PHASE_FINISHED:

		panic("Bug: get_basic_constraints ... FINISHED!")

	default:

		emap := attdata.HardConstraintMap
		for _, ctype := range attdata.ConstraintTypes {
			if strings.Contains(ctype, "NotAvailable") {
				if p != PHASE_BASIC {
					continue
				}
			} else if p != PHASE_HARD {
				continue
			}
			wlist = append(wlist,
				weighted_constraint_list{"", ctype, emap[ctype]})
		}

	}
	for _, wcl := range wlist {
		cixlist := []ConstraintIndex{}
		for _, i := range wcl.indexes {
			if !instance0.ConstraintEnabled[i] &&
				!attdata.BlockConstraint[i] {
				cixlist = append(cixlist, i)
			}
		}
		if len(cixlist) == 0 {
			continue
		}
		nconstraints += len(cixlist)
		instance := attdata.new_instance(
			instance0,
			wcl.ctype,
			wcl.weight,
			cixlist,
			attdata.cycle_timeout)
		instances = append(instances, instance)
	}
	return instances, nconstraints
}
