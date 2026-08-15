package fetrunner

import (
	"cmp"
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fmt"
	"slices"
	"strconv"
	"strings"
)

func init() {
	OpHandlerMap["TT_CLASSES"] = get_classes
	OpHandlerMap["TT_CLASS_PLACEMENTS"] = get_class_placements
	OpHandlerMap["TT_ATOMIC_GROUP_PLACEMENTS"] = get_atomic_group_placements
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
		a2gsx := map[int]string{}  // special map for groups with just one atomic index
		//TODO: cprefix := fmt.Sprintf("%d:", cix)
		for _, g := range cls.Groups {
			if _, ok := gmap[g.Tag]; ok {
				//TODO: gailist := []string{}
				for _, ai := range g.AtomicIndexes {
					//TODO: gailist = append(gailist, strconv.Itoa(ai))
					a2gs[ai] = append(a2gs[ai], g.Tag)
				}
				//TODO: base.LogResult("TT_CLASS_GROUP", cprefix+g.Tag+"::"+strings.Join(gailist, ","))
				if len(g.AtomicIndexes) == 1 {
					a2gsx[g.AtomicIndexes[0]] = g.Tag
				}
			}
			g2n[g.Tag] = len(g.AtomicIndexes)
		}
		ags := []string{}
		for _, ai := range cls.AtomicIndexes {
			ag, ok := a2gsx[ai]
			if ok {
				ags = append(ags, fmt.Sprintf("%d=%s", ai, ag))
			} else {
				ags = append(ags, fmt.Sprintf("%d=%s", ai, strings.Join(a2gs[ai], "&")))
			}
		}
		//TODO: save ags for the class somewhere in the Result?
		base.LogResult(op.Op, cls.Tag+"::"+strings.Join(ags, ",")+":"+cls.Separator)

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

func get_atomic_group_placements(op *DispatchOp) {
	agix, err := strconv.Atoi(op.Arg)
	if err != nil {
		panic(err)
	}
	lres := autotimetable.AutoTt.GetLastResult()
	for _, p := range autotimetable.AtomicGroupPlacements(lres, agix) {
		base.LogResult("ATOMIC_GROUP_PLACEMENT", autotimetable.SerializePlacement(p))
	}
}

func get_class_placements(op *DispatchOp) {
	cix, err := strconv.Atoi(op.Arg)
	if err != nil {
		panic(err)
	}
	lres := autotimetable.AutoTt.GetLastResult()

	// Place the activities in slots, building a list for each slot.
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

	week_slots = make([][][]ClassActivityPart, len(lres.Days))
	for d := range len(lres.Days) {
		week_slots[d] = make([][]ClassActivityPart, len(lres.Hours))
	}

	//TODO: Check that the atomic indexes are pre-sorted
	agall := lres.Classes[cix].AtomicIndexes
	natural_offsets := map[autotimetable.AtomicIndex]int{}
	for i, ag := range agall {
		natural_offsets[ag] = i
	}
	nagall := len(agall)
	for _, p := range autotimetable.ClassPlacements(lres, cix) {
		//TODO-- base.LogResult("CLASS_PLACEMENT", autotimetable.SerializePlacement(p))
		a := lres.Activities[p.Activity]
		// Get the atomics just from this class
		ags := []autotimetable.AtomicIndex{}
		for _, ag := range a.AtomicGroupIndexes {
			if slices.Contains(agall, ag) {
				ags = append(ags, ag)
			}
		}
		ag0 := ags[0]
		nag := len(ags)
		for i := range a.Duration {
			week_slots[p.Day][p.Hour+i] = append(week_slots[p.Day][p.Hour+i],
				ClassActivityPart{
					Index:         i,
					ActivityPtr:   a,
					Placement:     p,
					NaturalOffset: natural_offsets[ag0],
					TileFraction:  nag})
		}
	}

	// Determine tile offsets and sizes for each activity (part) in each slot.
	//TODO--
	//fmt.Printf("Class Index %d (%d)\n", cix, nagall)
	for d := range len(lres.Days) {
		for h := range len(lres.Hours) {
			//TODO--
			//fmt.Printf("::: %d.%d :::\n", d, h)

			// Sort in-place
			slot := week_slots[d][h]
			slices.SortStableFunc(slot, func(a1, a2 ClassActivityPart) int {
				return cmp.Compare(a1.NaturalOffset, a2.NaturalOffset)
			})
			// Calculate offsets, aggregating unused atomics as one empty space
			offset := 0
			empty := 0
			{
				sum := 0
				for _, ap := range slot {
					sum += ap.TileFraction
				}
				empty = nagall - sum
				if empty < 0 {
					panic("Too many groups in slot")
				}
			}
			for api, ap := range slot {
				if ap.NaturalOffset > offset {
					// Place empty space here.
					if empty == 0 {
						panic("No empty space to fill slot")
						//fmt.Println("No empty space to fill slot")
					}
					offset += empty
					empty = 0
				}
				slot[api] = ap

				//TODO: If an activity part has Index > 0, check that it has the same
				// offset as in the previous slot.

				if ap.Index == 0 {
					ap.Placement.Offset = offset
					ap.Placement.Size = ap.TileFraction

					base.LogResult("CLASS_PLACEMENT", autotimetable.SerializePlacement(ap.Placement))
				}

				/*TODO--
				fmt.Printf("  + i: %d, a: %s, nat: %d, o: %d, nag: %d\n",
					ap.Index,
					ap.ActivityPtr.Id,
					ap.NaturalOffset,
					offset,
					ap.TileFraction)
				*/

				offset += ap.TileFraction
			}
		}
	}
}

type ClassActivityPart struct {
	Index         int
	ActivityPtr   *autotimetable.TtActivity
	Placement     *autotimetable.TtActivityPlacement
	NaturalOffset int // Based on first atomic from the current class
	TileFraction  int // Number of atomics from the current class
}

var week_slots [][][]ClassActivityPart

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
