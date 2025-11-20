package w365tt

import (
	"fetrunner/db"
	"fmt"
	"strconv"
	"strings"
)

func (dbi *W365TopLevel) readRooms(newdb *db.DbTopLevel) {
	logger := newdb.Logger
	dbi.RealRooms = map[NodeRef]*db.Room{}
	dbi.RoomTags = map[string]NodeRef{}
	dbi.RoomChoiceNames = map[string]NodeRef{}
	for _, e := range dbi.Rooms {
		// Perform some checks and add to the RoomTags map.
	rloop:
		_, nok := dbi.RoomTags[e.Tag]
		if nok {
			logger.Error(
				"Room Tag (Shortcut) defined twice: %s\n",
				e.Tag)
			e.Tag += "$"
			goto rloop
		}
		dbi.RoomTags[e.Tag] = e.Id
		// Copy to base db.
		tsl := dbi.handleZeroAfternoons(e.NotAvailable, 1)
		r := newdb.NewRoom(e.Id)
		r.Tag = e.Tag
		r.Name = e.Name
		if len(tsl) != 0 {
			// Add a constraint
			newdb.NewRoomNotAvailable("", db.MAXWEIGHT, r.Id, tsl)
		}
		dbi.RealRooms[e.Id] = r
	}
}

// In the case of RoomGroups, cater for empty Tags (Shortcuts).
func (dbi *W365TopLevel) readRoomGroups(newdb *db.DbTopLevel) {
	logger := newdb.Logger
	dbi.RoomGroupMap = map[NodeRef]*db.RoomGroup{}
	for _, e := range dbi.RoomGroups {
		if e.Tag != "" {
		rloop:
			_, nok := dbi.RoomTags[e.Tag]
			if nok {
				logger.Error(
					"Room Tag (Shortcut) defined twice: %s\n",
					e.Tag)
				e.Tag += "$"
				goto rloop
			}
			dbi.RoomTags[e.Tag] = e.Id
		}
		// Copy to base db.
		r := newdb.NewRoomGroup(e.Id)
		r.Tag = e.Tag
		r.Name = e.Name
		r.Rooms = e.Rooms
		dbi.RoomGroupMap[e.Id] = r
	}
}

// Call this after all room types have been "read".
func (dbi *W365TopLevel) checkRoomGroups(newdb *db.DbTopLevel) {
	logger := newdb.Logger
	for _, e := range newdb.RoomGroups {
		// Collect the Ids and Tags of the component rooms.
		taglist := []string{}
		reflist := []NodeRef{}
		for _, rref := range e.Rooms {
			r, ok := dbi.RealRooms[rref]
			if ok {
				reflist = append(reflist, rref)
				taglist = append(taglist, r.Tag)
				continue

			}
			logger.Error(
				"Invalid Room in RoomGroup %s:\n  %s\n",
				e.Tag, rref)
		}
		if e.Tag == "" {
			// Make a new Tag
			var tag string
			i := 0
			for {
				i++
				tag = "{" + strconv.Itoa(i) + "}"
				_, nok := dbi.RoomTags[tag]
				if !nok {
					break
				}
			}
			e.Tag = tag
			dbi.RoomTags[tag] = e.Id
			// Also extend the name
			if e.Name == "" {
				e.Name = strings.Join(taglist, ",")
			} else {
				e.Name = strings.Join(taglist, ",") + ":: " + e.Name
			}
		} else if e.Name == "" {
			e.Name = strings.Join(taglist, ",")
		}
		e.Rooms = reflist
	}
}

func (dbi *W365TopLevel) makeRoomChoiceGroup(
	newdb *db.DbTopLevel,
	rooms []NodeRef,
) (NodeRef, string) {
	erlist := []string{} // Error messages
	// Collect the Ids and Tags of the component rooms.
	taglist := []string{}
	reflist := []NodeRef{}
	for _, rref := range rooms {
		r, ok := dbi.RealRooms[rref]
		if ok {
			reflist = append(reflist, rref)
			taglist = append(taglist, r.Tag)
			continue
		}
		erlist = append(erlist,
			fmt.Sprintf(
				"  ++ Invalid Room in new RoomChoiceGroup:\n  %s\n", rref))
	}
	name := strings.Join(taglist, ",")
	// Reuse existing Element when the rooms match.
	id, ok := dbi.RoomChoiceNames[name]
	if !ok {
		// Make a new Tag
		var tag string
		i := 0
		for {
			i++
			tag = "[" + strconv.Itoa(i) + "]"
			_, nok := dbi.RoomTags[tag]
			if !nok {
				break
			}
		}
		// Add new Element
		r := newdb.NewRoomChoiceGroup("")
		id = r.Id
		r.Tag = tag
		r.Name = name
		r.Rooms = reflist
		dbi.RoomTags[tag] = id
		dbi.RoomChoiceNames[name] = id
	}
	return id, strings.Join(erlist, "")
}
