package fet

import (
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fmt"
	"regexp"
	"strconv"

	"github.com/beevik/etree"
)

func (sourcefet *TtSourceFet) read_elements(fetroot *etree.Element) {
	{
		items := []element{}
		for _, e := range fetroot.SelectElement("Days_List").SelectElements("Day") {
			id := e.SelectElement("Name").Text()
			items = append(items, element{
				Id: base.NodeRef("Day:" + id), Tag: id})
		}
		sourcefet.days = items
	}

	{
		items := []element{}
		for _, e := range fetroot.SelectElement("Hours_List").SelectElements("Hour") {
			id := e.SelectElement("Name").Text()
			items = append(items, element{
				Id: base.NodeRef("Hour:" + id), Tag: id})
		}
		sourcefet.hours = items
	}

	{
		items := []element{}
		for _, e := range fetroot.SelectElement("Teachers_List").SelectElements("Teacher") {
			id := e.SelectElement("Name").Text()
			items = append(items, element{
				Id: base.NodeRef("Teacher:" + id), Tag: id})
		}
		sourcefet.teachers = items
	}

	{
		items := []element{}
		for _, e := range fetroot.SelectElement("Subjects_List").SelectElements("Subject") {
			id := e.SelectElement("Name").Text()
			items = append(items, element{
				Id: base.NodeRef("Subject:" + id), Tag: id})
		}
		sourcefet.subjects = items

	}

	{
		items := []element{}
		for _, e := range fetroot.SelectElement("Rooms_List").SelectElements("Room") {
			// Only include real rooms, skip virtual ones.
			if e.SelectElement("Virtual").Text() == "false" {
				id := e.SelectElement("Name").Text()
				items = append(items, element{
					Id: base.NodeRef("Room:" + id), Tag: id})
			}
		}
		sourcefet.rooms = items
	}

	{
		items := []*autotimetable.TtClass{}
		atomic_groups := []string{}
		students2atomics := map[string][]int{}
		for _, e := range fetroot.SelectElement("Students_List").SelectElements("Year") {
			id := e.SelectElement("Name").Text()

			// Read groups and subgroups, collect atomic groups
			class_ags := []int{}
			gel_list := e.SelectElements("Group")
			if len(gel_list) == 0 {
				// The class is an atomic group.
				agi := len(atomic_groups)
				students2atomics[id] = []int{agi}
				fmt.Printf("§ %s -> %v\n", id, students2atomics[id])
				atomic_groups = append(atomic_groups, id)
				class_ags = append(class_ags, agi)
			} else {
				for _, g := range gel_list {
					gtag := g.SelectElement("Name").Text()
					sgel_list := g.SelectElements("Subgroup")
					if len(sgel_list) == 0 {
						// The group is an atomic group.
						agi := len(atomic_groups)
						students2atomics[gtag] = []int{agi}
						fmt.Printf("§ %s -> %v\n", gtag, students2atomics[gtag])
						class_ags = append(class_ags, agi)
						atomic_groups = append(atomic_groups, gtag)
					} else {
						group_ags := []int{}
						for _, sg := range sgel_list {
							// These are all atomic groups, but drop repeats.
							sgtag := sg.SelectElement("Name").Text()
							agi := len(atomic_groups)
							agil, ok := students2atomics[sgtag]
							if ok {
								if len(agil) != 1 {
									panic("TODO: invalid year/group/subgroup structure")
								}
								agi = agil[0]
							} else {
								students2atomics[sgtag] = []int{agi}
								fmt.Printf("§ %s -> %v\n", sgtag, students2atomics[sgtag])
								atomic_groups = append(atomic_groups, sgtag)
								class_ags = append(class_ags, agi)
							}
							group_ags = append(group_ags, agi)
						}
						students2atomics[gtag] = group_ags
						fmt.Printf("§ %s -> %v\n", gtag, students2atomics[gtag])
					}
				}
				students2atomics[id] = class_ags
				fmt.Printf("§ %s -> %v\n", id, students2atomics[id])
			}

			items = append(items, &autotimetable.TtClass{
				Id:            base.NodeRef("Class:" + id),
				Tag:           id,
				AtomicIndexes: class_ags})
			//TODO: Groups []*autotimetable.TtGroup
		}
		sourcefet.classes = items
		sourcefet.atomic_groups = atomic_groups
		sourcefet.students2atomics = students2atomics
	}
}

func readInt(s string) int {
	i, e := strconv.Atoi(s)
	if e != nil {
		panic("Not an integer: " + s)
	}
	return i
}

// Get active activities, count inactive ones. Return the number of inactive
// activities in the source – these are ignored.
func (sourcefet *TtSourceFet) read_activities(fetroot *etree.Element) int {
	//a_elements := []*etree.Element{}
	activities := []*ttActivity{}
	ael := fetroot.SelectElement("Activities_List")
	inactive := 0
	// A teacher-map is required, tag -> element
	tmap := map[string]int{}
	for i, t := range sourcefet.teachers {
		tmap[t.Tag] = i
	}
	for _, a := range ael.ChildElements() {
		if a.SelectElement("Active").Text() == "true" {
			//a_elements = append(a_elements, a)
			id := a.SelectElement("Id").Text()
			glist := []element{}
			aglist := []int{}
			for _, g := range a.SelectElements("Students") {
				gt := g.Text()
				glist = append(glist, element{Id: "Group:" + NodeRef(gt), Tag: gt})
				aglist = append(aglist, sourcefet.students2atomics[gt]...)
			}
			tlist := []autotimetable.TeacherIndex{}
			for _, t := range a.SelectElements("Teacher") {
				tt := t.Text()
				tix, ok := tmap[tt]
				if !ok {
					panic("Unknown teacher: " + tt)
				}
				tlist = append(tlist, tix)
			}
			activities = append(activities, &ttActivity{
				Id:  "Activity:" + id,
				Tag: id, // In this case, this field is set here, for the "FET" back-end.
				// Although they are not strictly required, the remaining fields can
				// make reading the JSON result file easier.
				Duration:           readInt(a.SelectElement("Duration").Text()),
				Subject:            a.SelectElement("Subject").Text(),
				Groups:             glist,
				AtomicGroupIndexes: aglist,
				Teachers:           tlist,
			})
		} else {
			inactive++
		}
	}
	//sourcefet.activityElements = a_elements
	sourcefet.activities = activities
	return inactive
}

// Collect the constraints, dividing into soft and hard groups and counting
// inactive ones, which are then ignored. Return the number of inactive ones
// (time and space separately)
func (sourcefet *TtSourceFet) read_constraints(fetroot *etree.Element) (int, int) {

	// Regexp to match constraint comment which has a number tag already:
	// Soft constraints also have a weight.
	r_constraint_number := regexp.MustCompile(`^\[[0-9]+.*\](.*)$`)

	hard_constraint_map := map[constraintType][]constraintIndex{}
	soft_constraint_map := map[constraintType][]constraintIndex{}
	constraint_types := []constraintType{}

	var t_inactive int
	var s_inactive int
	for timespace := range 2 {
		// First (timespace == 0) collect active time constraints,
		// then (timespace == 1) collect active space constraints.

		var et *etree.Element
		var bc string
		if timespace == 0 {
			et = fetroot.SelectElement("Time_Constraints_List")
			bc = "ConstraintBasicCompulsoryTime"
		} else {
			et = fetroot.SelectElement("Space_Constraints_List")
			bc = "ConstraintBasicCompulsorySpace"
		}
		tsindexes := []int{}
		inactive := 0
		for ic, e := range et.ChildElements() {
			// Count and skip if inactive
			if e.SelectElement("Active").Text() == "false" {
				inactive++ // count inactive constraints
				continue
			}
			ctype := constraintType(e.Tag)
			if ctype == bc {
				// Basic, non-negotiable constraint
				continue
			}
			tsindexes = append(tsindexes, ic)

			i := len(sourcefet.constraints)
			sourcefet.constraintElements = append(sourcefet.constraintElements, e)
			//if timespace == 0 {
			//  sourcefet.timeConstraints = append(sourcefet.timeConstraints, i)
			//} else {
			//  sourcefet.spaceConstraints = append(sourcefet.spaceConstraints, i)
			//}

			w := e.SelectElement("Weight_Percentage").Text()
			wdb := FetWeight2Db(w, sourcefet.weightTable)
			//fmt.Printf(" ++ %02d: %s (%s -> %02d)\n", i, ctype, w, wdb)
			if w == "100" {
				// Hard constraint
				hard_constraint_map[ctype] = append(hard_constraint_map[ctype],
					constraintIndex(i))
			} else {
				// Soft constraint
				wctype := fmt.Sprintf("%02d:%s", wdb, ctype)
				soft_constraint_map[wctype] = append(soft_constraint_map[wctype],
					constraintIndex(i))
				sourcefet.softWeights = append(sourcefet.softWeights, softWeight{i, w})
			}
			constraint_types = append(constraint_types, ctype)
			// ... duplicates wil be removed in `sort_constraint_types`

			// Ensure that the constraints are numbered in their Comments.
			// This is to ease referencing in the results object.
			comments := e.SelectElement("Comments")
			comment := ""
			if comments == nil {
				comments = e.CreateElement("Comments")
			} else {
				// Remove any existing comment id
				comment = comments.Text()
				parts := r_constraint_number.FindStringSubmatch(comment)
				if parts != nil {
					comment = parts[1]
				}
			}
			wtag := ""
			if w != "100" {
				wtag = ":" + w
			}
			// In FET, the constraints have no identifiers/tags, so one is
			// added in the "Comments"  field.
			cid := fmt.Sprintf("[%d%s]", i, wtag)
			comments.SetText(cid + comment)
			sourcefet.constraints = append(sourcefet.constraints, &ttConstraint{
				Id:     cid,
				CType:  ctype,
				Weight: wdb,
			})
		}

		if timespace == 0 {
			sourcefet.t_constraints = tsindexes
			t_inactive = inactive
		} else {
			sourcefet.s_constraints = tsindexes
			s_inactive = inactive
		}
	}

	//sourcefet.nConstraints = constraintIndex(len(sourcefet.constraintElements))
	sourcefet.constraintTypes = autotimetable.SortConstraintTypes(
		constraint_types, ConstraintPriority)
	sourcefet.hardConstraintMap = hard_constraint_map
	sourcefet.softConstraintMap = soft_constraint_map
	return t_inactive, s_inactive
}
