package fetrunner

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fmt"
	"strconv"
	"strings"
)

func init() {
	OpHandlerMap["_TT_HAS_RESULT"] = has_result // no log entry
	OpHandlerMap["TT_DAYS"] = get_days
	OpHandlerMap["TT_HOURS"] = get_hours
	OpHandlerMap["TT_CLASSES"] = get_classes
	OpHandlerMap["TT_TEACHERS"] = get_teachers
	OpHandlerMap["TT_ROOMS"] = get_rooms
	OpHandlerMap["TT_ACTIVITIES"] = get_activities
	OpHandlerMap["TT_CLASS_PLACEMENTS"] = get_class_placements
	OpHandlerMap["TT_TEACHER_PLACEMENTS"] = get_teacher_placements
	OpHandlerMap["TT_ROOM_PLACEMENTS"] = get_room_placements
}

// The AutoTtData instance is available as `autotimetable.AutoTt`.

func has_result(op *DispatchOp) {
	if autotimetable.AutoTt.GetLastResult() != nil {
		op.CC = 1
	}
}

func get_days(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	for _, d := range lres.Days {
		base.LogResult(op.Op, d.Tag+":")
	}
}

func get_hours(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	for _, h := range lres.Hours {
		base.LogResult(op.Op, h.Tag+":")
	}
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
		ailist := []string{}
		for _, ai := range cls.AtomicIndexes {
			ailist = append(ailist, strconv.Itoa(ai))
		}
		base.LogResult(op.Op, cls.Tag+"::"+strings.Join(ailist, ","))

		gn := map[string]int{}
		cprefix := fmt.Sprintf("%d:", cix)
		for _, g := range cls.Groups {
			if _, ok := gmap[g.Tag]; ok {
				gailist := []string{}
				for _, ai := range g.AtomicIndexes {
					gailist = append(gailist, strconv.Itoa(ai))
				}
				base.LogResult("TT_CLASS_GROUP", cprefix+g.Tag+"::"+strings.Join(gailist, ","))
			}
			gn[g.Tag] = len(g.AtomicIndexes)
		}

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
				s := gn[g]
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
				dglist = append(dglist, fmt.Sprintf("%s:%d:%d", dg.tag, dg.offset, dg.size))
			}
			base.LogResult("TT_CLASS_DIVISION", cprefix+strings.Join(dglist, ","))
		}
	}
}

// TODO: (long) names
func get_teachers(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	for _, t := range lres.Teachers {
		base.LogResult(op.Op, t.Tag+":")
	}
}

// TODO: (long) names
func get_rooms(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	for _, r := range lres.Rooms {
		base.LogResult(op.Op, r.Tag+":")
	}
}

func get_activities(op *DispatchOp) {
	lres := autotimetable.AutoTt.GetLastResult()
	for _, a := range lres.Activities {
		tlist := []string{}
		for _, tix := range a.Teachers {
			tlist = append(tlist, strconv.Itoa(tix))
		}
		aglist := []string{}
		for _, agix := range a.AtomicGroupIndexes {
			aglist = append(aglist, strconv.Itoa(agix))
		}
		glist := []string{}
		for _, g := range a.Groups {
			glist = append(glist, g.Tag)
		}
		base.LogResult(op.Op, fmt.Sprintf("%d:%s:%s:%s:%s",
			a.Duration, a.Subject,
			strings.Join(tlist, ","),
			strings.Join(aglist, ","),
			strings.Join(glist, ",")))
	}
}

func get_class_placements(op *DispatchOp) {
	cix, err := strconv.Atoi(op.Arg)
	if err != nil {
		panic(err)
	}
	lres := autotimetable.AutoTt.GetLastResult()
	for _, p := range autotimetable.ClassPlacements(lres, cix) {
		base.LogResult("CLASS_PLACEMENT", autotimetable.SerializePlacement(p))
	}
}

func get_teacher_placements(op *DispatchOp) {
	tix, err := strconv.Atoi(op.Arg)
	if err != nil {
		panic(err)
	}
	lres := autotimetable.AutoTt.GetLastResult()
	for _, p := range autotimetable.TeacherPlacements(lres, tix) {
		base.LogResult("TEACHER_PLACEMENT", autotimetable.SerializePlacement(p))
	}
}

func get_room_placements(op *DispatchOp) {
	rix, err := strconv.Atoi(op.Arg)
	if err != nil {
		panic(err)
	}
	lres := autotimetable.AutoTt.GetLastResult()
	for _, p := range autotimetable.RoomPlacements(lres, rix) {
		base.LogResult("ROOM_PLACEMENT", autotimetable.SerializePlacement(p))
	}
}

//TODO: Consider extending BaseElement to include a "long name". Alternatively,
// these could be fetched by further "get_xxx()" calls.

//TODO: Somehow it will be necessary to recognise which classes are
// represented by the groups. If there is always a clear separator
// in the group name between class part and group part, it could be
// done as a string split with the existing data. Another possibility
// would be to provide a mapping, e.g. via a list of groups for each
// class. Another possibility would be to pass the full atomic group
// list for each activity and also (separately) the lists for each
// class, so that a map can be built.
