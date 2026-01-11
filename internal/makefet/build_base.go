package makefet

import (
    "fetrunner/internal/autotimetable"
    "fetrunner/internal/base"
    "fetrunner/internal/fet"
    "fetrunner/internal/timetable"

    "github.com/beevik/etree"
)

const fet_version = "7.5.5"

type IdPair = autotimetable.IdPair
type Constraint = autotimetable.Constraint

type FetBuild struct {
    basedata               *base.BaseData
    ttdata                 *timetable.TtData
    rundata                *fet.TtRunDataFet // the structure to be built
    fetroot                *etree.Element
    room_list              *etree.Element // needed for adding virtual rooms
    activity_tag_list      *etree.Element // in case these are needed
    time_constraints_list  *etree.Element
    space_constraints_list *etree.Element

    constraint_counter int // for tagging constraints

    // Cache for FET virtual rooms, "hash" -> FET-virtual-room tag
    fet_virtual_rooms  map[string]string
    fet_virtual_room_n map[string]int // FET-virtual-room tag -> number of room sets

    real_soft bool // if false, give soft constraints weight 100
}
