package timetable

import (
	"fetrunner/internal/base"
)

// Collect the "not-available" constraints.
// Although the data structures support weights of less than 100%, the input
// data may not.
func (tt_data *TtData) get_blocked_slots(constraint_map map[string][]*base.BaseConstraint) {
	// Rooms
	for _, c0 := range constraint_map[base.C_RoomNotAvailable] {
		srcdata := c0.Data.(base.ResourceNotAvailable)
		r := tt_data.room2Index[srcdata.Resource]
		tt_data.constraints = append(tt_data.constraints, &ttConstraint{
			Id:     string(c0.Id),
			CType:  c0.CType,
			Weight: c0.Weight,
			Data: map[string]any{
				"Room": r, "Times": srcdata.NotAvailable},
		})
	}
	delete(constraint_map, base.C_RoomNotAvailable)

	// Teachers
	for _, c0 := range constraint_map[base.C_TeacherNotAvailable] {
		srcdata := c0.Data.(base.ResourceNotAvailable)
		t := tt_data.teacher2Index[srcdata.Resource]
		tt_data.constraints = append(tt_data.constraints, &ttConstraint{
			Id:     string(c0.Id),
			CType:  c0.CType,
			Weight: c0.Weight,
			Data: map[string]any{
				"Teacher": t, "Times": srcdata.NotAvailable},
		})
	}
	delete(constraint_map, base.C_TeacherNotAvailable)

	// Classes
	for _, c0 := range constraint_map[base.C_ClassNotAvailable] {
		srcdata := c0.Data.(base.ResourceNotAvailable)
		c := tt_data.class2Index[srcdata.Resource]
		tt_data.constraints = append(tt_data.constraints, &ttConstraint{
			Id:     string(c0.Id),
			CType:  c0.CType,
			Weight: c0.Weight,
			Data: map[string]any{
				"Class": c, "Times": srcdata.NotAvailable},
		})
	}
	delete(constraint_map, base.C_ClassNotAvailable)
}

func (tt_data *TtData) GetResourceUnavailableConstraintTypes() []constraintType {
	return []constraintType{
		base.C_ClassNotAvailable,
		base.C_TeacherNotAvailable,
		base.C_RoomNotAvailable,
	}
}
