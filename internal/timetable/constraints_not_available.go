package timetable

import (
	"fetrunner/internal/base"
)

// Collect the "not-available" constraints, keeping an additional list of the
// hard ones.
// Although the data structures support weights of less than 100%, the input
// data may not.
func (tt_data *TtData) get_blocked_slots(constraint_map map[string][]*base.Constraint) {
	// Rooms
	for _, c0 := range constraint_map[base.C_RoomNotAvailable] {
		i := len(tt_data.constraints)
		srcdata := c0.Data.(base.ResourceNotAvailable)
		r := tt_data.Room2Index[srcdata.Resource]
		tt_data.constraints = append(tt_data.constraints, &constraint{
			Id:     string(c0.Id),
			CType:  c0.CType,
			Weight: c0.Weight,
			Data: map[string]any{
				"Room": r, "Times": srcdata.NotAvailable},
		})
		if c0.Weight == base.MAXWEIGHT {
			tt_data.hard_not_available = append(tt_data.hard_not_available, i)
		}
	}
	delete(constraint_map, base.C_RoomNotAvailable)

	// Teachers
	for _, c0 := range constraint_map[base.C_TeacherNotAvailable] {
		i := len(tt_data.constraints)
		srcdata := c0.Data.(base.ResourceNotAvailable)
		t := tt_data.Teacher2Index[srcdata.Resource]
		tt_data.constraints = append(tt_data.constraints, &constraint{
			Id:     string(c0.Id),
			CType:  c0.CType,
			Weight: c0.Weight,
			Data: map[string]any{
				"Teacher": t, "Times": srcdata.NotAvailable},
		})
		if c0.Weight == base.MAXWEIGHT {
			tt_data.hard_not_available = append(tt_data.hard_not_available, i)
		}
	}
	delete(constraint_map, base.C_TeacherNotAvailable)

	// Classes
	for _, c0 := range constraint_map[base.C_ClassNotAvailable] {
		i := len(tt_data.constraints)
		srcdata := c0.Data.(base.ResourceNotAvailable)
		c := tt_data.Class2Index[srcdata.Resource]
		tt_data.constraints = append(tt_data.constraints, &constraint{
			Id:     string(c0.Id),
			CType:  c0.CType,
			Weight: c0.Weight,
			Data: map[string]any{
				"Class": c, "Times": srcdata.NotAvailable},
		})
		if c0.Weight == base.MAXWEIGHT {
			tt_data.hard_not_available = append(tt_data.hard_not_available, i)
		}
	}
	delete(constraint_map, base.C_ClassNotAvailable)
}
