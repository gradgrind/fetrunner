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
	tt_data.teacher_hard_blocked = make([][][]bool, len(tt_data.teachers))
	for _, c0 := range constraint_map[base.C_TeacherNotAvailable] {
		srcdata := c0.Data.(base.ResourceNotAvailable)
		tix := tt_data.teacher2Index[srcdata.Resource]
		tt_data.constraints = append(tt_data.constraints, &ttConstraint{
			Id:     string(c0.Id),
			CType:  c0.CType,
			Weight: c0.Weight,
			Data: map[string]any{
				"Teacher": tix, "Times": srcdata.NotAvailable},
		})
		blocked_slots := make([][]bool, tt_data.ndays)
		for d := range tt_data.ndays {
			blocked_slots[d] = make([]bool, tt_data.nhours)
		}
		for _, nas := range srcdata.NotAvailable {
			blocked_slots[nas.Day][nas.Hour] = true
		}
		tt_data.teacher_hard_blocked[tix] = blocked_slots
	}
	delete(constraint_map, base.C_TeacherNotAvailable)

	// Classes
	tt_data.class_hard_blocked = make([][][]bool, len(tt_data.classDivisions))
	for _, c0 := range constraint_map[base.C_ClassNotAvailable] {
		srcdata := c0.Data.(base.ResourceNotAvailable)
		cix := tt_data.class2Index[srcdata.Resource]
		tt_data.constraints = append(tt_data.constraints, &ttConstraint{
			Id:     string(c0.Id),
			CType:  c0.CType,
			Weight: c0.Weight,
			Data: map[string]any{
				"Class": cix, "Times": srcdata.NotAvailable},
		})
		blocked_slots := make([][]bool, tt_data.ndays)
		for d := range tt_data.ndays {
			blocked_slots[d] = make([]bool, tt_data.nhours)
		}
		for _, nas := range srcdata.NotAvailable {
			blocked_slots[nas.Day][nas.Hour] = true
		}
		tt_data.class_hard_blocked[cix] = blocked_slots
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
