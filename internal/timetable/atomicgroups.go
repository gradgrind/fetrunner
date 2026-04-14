package timetable

import (
    "strings"
)

const ATOMIC_GROUP_SEP1 = "#"
const ATOMIC_GROUP_SEP2 = "~"

// Prepare filtered versions of the class Divisions containing only
// those Divisions which have Groups used in activities.
func (tt_data *TtData) FilterDivisions() {
    db := tt_data.db
    // Collect groups used in courses
    usedgroups := map[nodeRef]bool{}

    // Gather groups from the SuperCourses.
    for _, spc := range db.SuperCourses {
        for _, sbc := range spc.SubCourses {
            for _, gref := range sbc.Groups {
                usedgroups[gref] = true
            }
        }
    }
    // Gather groups from the plain Courses.
    for _, c := range db.Courses {
        for _, gref := range c.Groups {
            usedgroups[gref] = true
        }
    }

    // Filter the class divisions, discarding the division names.
    for _, c := range db.Classes {
        divs := [][]nodeRef{}
        for _, div := range c.Divisions {
            for _, gref := range div.Groups {
                if usedgroups[gref] {
                    divs = append(divs, div.Groups)
                    break
                }
            }
        }
        tt_data.classDivisions = append(tt_data.classDivisions,
            classDivision{c, divs})
    }
}

type AtomicGroup struct {
    Class  nodeRef
    Groups []nodeRef
    Tag    string // A constructed tag to represent the atomic group
}

func (tt_data *TtData) MakeAtomicGroups() {
    db := tt_data.db

    // Set up the class index map
    tt_data.class2Index = map[nodeRef]classIndex{}

    // An atomic group is an ordered list of single groups, one from each
    // division.
    tt_data.atomicGroup2Indexes = map[nodeRef][]atomicIndex{}

    // Go through the classes inspecting their Divisions.
    // Build a list-basis for the atomic groups based on the Cartesian product.
    for i, cdivs := range tt_data.classDivisions {
        cl := cdivs.Class
        tt_data.class2Index[cl.Id] = i
        if len(cdivs.Divisions) == 0 {
            // Make an atomic group for the class
            agix := len(tt_data.atomicGroups)
            ag := &AtomicGroup{
                //Index: agix,
                Class: cl.Id,
                Tag:   cl.Tag + ATOMIC_GROUP_SEP1,
            }
            tt_data.atomicGroups = append(tt_data.atomicGroups, ag)
            tt_data.atomicGroup2Indexes[cl.ClassGroup] = []atomicIndex{
                atomicIndex(agix)}
            continue
        }

        // The atomic groups will be built as a list of lists of Refs.
        agrefs := [][]nodeRef{{}}
        for _, dglist := range cdivs.Divisions {
            // Add another division – increases underlying list lengths.
            agrefsx := [][]nodeRef{}
            for _, ag := range agrefs {
                // Extend each of the old list items by appending each
                // group of the new division in turn – multiplies the
                // total number of atomic groups.
                newlen := len(ag) + 1
                for _, g := range dglist {
                    gx := make([]nodeRef, newlen)
                    copy(gx, append(ag, g))
                    agrefsx = append(agrefsx, gx)
                }
            }
            agrefs = agrefsx
        }
        //fmt.Printf("  §§§ Divisions in %s: %+v\n", cl.Tag, divs)
        //fmt.Printf("     --> %+v\n", agrefs)

        // Make AtomicGroups
        aglist := []atomicIndex{}
        for _, ag := range agrefs {
            glist := []string{}
            for _, gref := range ag {
                gtag := db.Ref2Tag(gref)
                glist = append(glist, gtag)
            }
            agix := len(tt_data.atomicGroups)
            ag := &AtomicGroup{
                Class:  cl.Id,
                Groups: ag,
                Tag: cl.Tag + ATOMIC_GROUP_SEP1 +
                    strings.Join(glist, ATOMIC_GROUP_SEP2),
            }
            tt_data.atomicGroups = append(tt_data.atomicGroups, ag)
            aglist = append(aglist, atomicIndex(agix))
        }
        tt_data.atomicGroup2Indexes[cl.ClassGroup] = aglist
        // Map the individual groups to their atomic groups.
        count := 1
        divIndex := len(cdivs.Divisions)
        for divIndex > 0 {
            divIndex--
            divGroups := cdivs.Divisions[divIndex]
            agi := 0 // ag index
            for agi < len(aglist) {
                for _, g := range divGroups {
                    for j := 0; j < count; j++ {
                        tt_data.atomicGroup2Indexes[g] = append(
                            tt_data.atomicGroup2Indexes[g], aglist[agi])
                        agi++
                    }
                }
            }
            count *= len(divGroups)
        }
    }
}
