package autotimetable

import (
	"cmp"
	"slices"
	"strconv"
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
// Use this list as the new run queue.
// Return the list of instances and the total number of individual constraints.
func (attdata *AutoTtData) get_basic_constraints() {
	instances := []*TtInstance{} // one instance per constraint type
	//nconstraints := 0            // count constraints
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
			return strings.Compare(b.weight, a.weight)
		})

	case PHASE_FINISHED:
		panic("Bug: get_basic_constraints ... FINISHED!")

	case PHASE_BASIC, PHASE_HARD:
		natypes := attdata.Source.GetPhase0ConstraintTypes()
		emap := attdata.HardConstraintMap
	nexttype:
		for _, ctype := range attdata.Constraint_Types {
			for _, ct := range natypes {
				if ct == ctype {
					// It is a "not-available" constraint.
					if p == PHASE_BASIC {
						wlist = append(wlist,
							weighted_constraint_list{"", ctype, emap[ctype]})
					}
					continue nexttype
				}
			}
			// It is not a "not-available" constraint.
			if p == PHASE_HARD {
				wlist = append(wlist,
					weighted_constraint_list{"", ctype, emap[ctype]})
			}
		}

	default:
		panic("Bug: Unexpected PHASE: " + strconv.Itoa(p))
	}
	instance0 := attdata.current_instance
	for _, wcl := range wlist {
		cixlist := []ConstraintIndex{}
		for _, i := range wcl.indexes {
			if !instance0.ConstraintEnabled[i] {
				cixlist = append(cixlist, i)
			}
		}
		if len(cixlist) == 0 {
			continue
		}
		//nconstraints += len(cixlist)
		instances = append(instances, &TtInstance{
			// Make a new `TtInstance`
			Timeout:      attdata.cycle_timeout,
			BaseInstance: instance0,

			ConstraintType: wcl.ctype,
			Constraints:    cixlist,
			Weight:         wcl.weight,
		})
	}
	attdata.set_runqueue(instances)
}

func SortConstraintTypes(
	constraint_types []ConstraintType,
	priority map[string]int,
) []ConstraintType {
	slices.Sort(constraint_types)
	l := slices.Compact(constraint_types)
	slices.SortStableFunc(l,
		func(a, b ConstraintType) int {
			return cmp.Compare(priority[b], priority[a])
		})
	return l
}
