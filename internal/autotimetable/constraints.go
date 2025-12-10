package autotimetable

// Collect the individual constraints by type in the order given by the
// `ConstraintTypes` list, inclkuding only those which are disabled.
func (attdata *AutoTtData) get_basic_constraints(
	instance0 *TtInstance, soft bool,
) ([]*TtInstance, int) {
	instances := []*TtInstance{} // one instance per constraint type
	nconstraints := 0            // count constraints
	var emap map[ConstraintType][]ConstraintIndex
	if soft {
		emap = attdata.SoftConstraintMap
	} else {
		emap = attdata.HardConstraintMap
	}
	for _, ctype := range attdata.ConstraintTypes {
		blist := emap[ctype]
		cixlist := []ConstraintIndex{}
		for _, i := range blist {
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
			ctype,
			cixlist,
			attdata.cycle_timeout)
		instances = append(instances, instance)
	}
	return instances, nconstraints
}
