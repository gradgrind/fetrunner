package autotimetable

// Collect the individual constraints by type in the order given by the
// `ConstraintTypes` list, inclkuding only those which are disabled.
func (basic_data *BasicData) get_basic_constraints(
	instance0 *TtInstance, soft bool,
) ([]*TtInstance, int) {
	instances := []*TtInstance{} // one instance per constraint type
	nconstraints := 0            // count constraints
	var emap map[ConstraintType][]ConstraintIndex
	if soft {
		emap = basic_data.SoftConstraintMap
	} else {
		emap = basic_data.HardConstraintMap
	}
	for _, ctype := range basic_data.ConstraintTypes {
		blist := emap[ctype]
		cixlist := []ConstraintIndex{}
		for _, i := range blist {
			if !instance0.ConstraintEnabled[i] {
				cixlist = append(cixlist, i)
			}
		}
		if len(cixlist) == 0 {
			continue
		}
		nconstraints += len(cixlist)
		instance := basic_data.new_instance(
			instance0,
			ctype,
			cixlist,
			basic_data.CYCLE_TIMEOUT,
			soft)
		instances = append(instances, instance)
	}
	return instances, nconstraints
}
