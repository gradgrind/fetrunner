package fetrunner

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fmt"
	"strconv"
	"strings"
)

func init() {
	OpHandlerMap["TT_CLASSES_AND_GROUPS"] = get_classes_and_groups
	OpHandlerMap["TT_CLASSES"] = get_classes               //TODO: perhaps deprecated?
	OpHandlerMap["TT_CLASSVIEW_DATA"] = get_classview_data //TODO
}

// TODO: The idea is to replace the old TT_CLASSES by a version returning classes,
// atomics and the "used" groups. Include class/group -> atomics info?
// The timetable slot structure would be handled separately, primarily in the
// fetrunner library (Go) and for one class at a time. Also for the atomics, or
// would these have their own operation?
func get_classes_and_groups(op *DispatchOp) {
}

func get_classes(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	// Collect group tags which are actually used by activities.
	gmap := map[string]int{}
	for _, a := range lres.Activities {
		for _, g := range a.Groups {
			gmap[g.Tag] = -1
		}
	}

	type divgroup struct {
		tag    string
		offset int
		size   int
	}
	for cix, cls := range lres.Classes {
		// Get the atomic groups for the whole class.
		//TODO? If I could ensure that the atomic groups of a class are consecutive,
		// an alternative would be to include just start and end index here.
		//TODO: ailist := []string{}
		//TODO: for _, ai := range cls.AtomicIndexes {
		//TODO: 	ailist = append(ailist, strconv.Itoa(ai))
		//TODO: }
		//TODO: I might rather want the intersection of groups for each atomic group:
		//TODO: base.LogResult(op.Op, cls.Tag+"::"+strings.Join(ailist, ","))

		g2n := map[string]int{}    // map groups to their number of atomic indexes
		a2gs := map[int][]string{} // map atomic indexes to the groups containing them
		//TODO: cprefix := fmt.Sprintf("%d:", cix)
		for _, g := range cls.Groups {
			if _, ok := gmap[g.Tag]; ok {
				//TODO: gailist := []string{}
				for _, ai := range g.AtomicIndexes {
					//TODO: gailist = append(gailist, strconv.Itoa(ai))
					a2gs[ai] = append(a2gs[ai], g.Tag)
				}
				//TODO: base.LogResult("TT_CLASS_GROUP", cprefix+g.Tag+"::"+strings.Join(gailist, ","))
			}
			g2n[g.Tag] = len(g.AtomicIndexes)
		}
		ags := []string{}
		for _, ai := range cls.AtomicIndexes {
			ags = append(ags, fmt.Sprintf("%d=%s", ai, strings.Join(a2gs[ai], "&")))
		}
		//TODO: save ags for the class somewhere in the Result?
		base.LogResult(op.Op, cls.Tag+"::"+strings.Join(ags, ","))

		// Discover the possible class divisions and reduce these to include only
		// groups which are actually used, and eliminate subsets.
		dglists := autotimetable.ClassDivisions(lres, cix)
		udglists := [][]divgroup{}
	gnext:
		for _, glist := range dglists {
			// Filter out groups not used by activities, collecting offsets and sizes.
			uglist := []divgroup{}
			i := 0
			for _, g := range glist {
				s := g2n[g]
				if c, ok := gmap[g]; ok {
					if c != -1 && c != cix {
						panic("Group defined in two classes: " + g)
					}
					gmap[g] = cix
					// Group is used by an activity
					uglist = append(uglist, divgroup{g, i, s})
				}
				i += s
			}

			// Filter out subsets.
		knext:
			for k, kgl := range udglists {
				if len(uglist) > len(kgl) {
					// If the groups in `kgl` are a subset of those in `uglist`,
					// replace the entry in `udglists`.
				kgnext:
					for _, kg := range kgl {
						for _, ug := range uglist {
							if kg.tag == ug.tag {
								// found a match
								continue kgnext
							}
						}
						// not a subset
						continue knext
					}
					// a subset => replace
					udglists[k] = uglist
					continue gnext
				} else {
					// If the groups in `uglist` are a subset of those in `kgl`,
					// don't add to `udglists`.
				ugnext:
					for _, ug := range uglist {
						for _, kg := range kgl {
							if kg.tag == ug.tag {
								// found a match
								continue ugnext
							}
						}
						// not a subset
						continue knext
					}
					// It is a subset.
					continue gnext
				}
			}
			// Otherwise add non-empty `uglist` to `udglists`.
			if len(uglist) != 0 {
				udglists = append(udglists, uglist)
			}
		}
		for _, gl := range udglists {
			dglist := []string{}
			for _, dg := range gl {
				dglist = append(dglist, fmt.Sprintf("%s@%d+%d", dg.tag, dg.offset, dg.size))
			}
			//TODO: base.LogResult("TT_CLASS_DIVISION", cprefix+strings.Join(dglist, ","))
		}
	}
}

func get_classview_data(op *DispatchOp) {
	cix, err := strconv.Atoi(op.Arg)
	if err != nil {
		panic(err)
	}
	lres := autotimetable.AutoTt.GetLastResult()

	//TODO ...
	// Place the activities in slots, building a list for each slot. If multiple groups
	// are listed within the current class, add multiple entries, one for each group.
	// Multiple-slot activities (duration > 1) should get entries in each of the covered
	// slots, each entry indicating its index within the activity.

	/* TODO--
	type TtActivityPlacement struct {
		Activity ActivityIndex
		Day      int
		Hour     int
		Rooms    []RoomIndex
	}
	type TtActivity struct {
		Id string // FET: "Activity:" + ActivityId
		// DB: NodeRef of the source activity from which this is derived.
		Tag string // optionally usable by the back-end

		Duration           int
		Subject            string
		Groups             []base.ElementBase // a `Class` is represented by its ClassGroup
		AtomicGroupIndexes []AtomicIndex
		Teachers           []TeacherIndex
	}
	*/
	for _, p := range autotimetable.ClassPlacements(lres, cix) {
		//? base.LogResult("CLASS_PLACEMENT", autotimetable.SerializePlacement(p))
		a := lres.Activities[p.Activity]

		//TODO

		_ = a

	}

}

// If the GUI is allowed to move activities, there must be enough information available
// to test the validity of a move. It can be a complicated affair, depending on which
// constraints are to be observed. The simplest would be to check just teachers and
// student groups (presumably using the atomic groups), perhaps also rooms. At present
// the room requirements are not directly available, but they could be made accessible.
// On the other hand, it might be better to allow any available rooms to be allocated,
// on the basis that the operator should know what is acceptable and should have the
// possibility of overriding the specification. This argument could also be extended
// to other constraints, leaving only the teacher and student-group clashes as illegal.
// Any general inclusion of automatic constraint testing would require that the
// constraints are available in a usable form to the GUI. At least in the case of a FET
// file that is not at all trivial, it would require a complete implementation of all
// FET constraints.
