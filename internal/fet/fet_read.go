package fet

import (
	"encoding/json"
	"fetrunner/internal/autotimetable"
	"fetrunner/internal/base"
	"fmt"
	"regexp"
	"strconv"

	"github.com/beevik/etree"
)

type SourceFET struct {
	doc *etree.Document
}

func (s *SourceFET) SourceType() string {
	return "FET"
}

// In FET there are "time" constraints and "space" constraints. They are
// all lumped together in the `ConstraintElements` list.

func FetRead(
	bdata *base.BaseData,
	fetpath string,
) *SourceFET {
	logger := bdata.Logger
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fetpath); err != nil {
		logger.Error("%s", err)
		return nil
	}
	return &SourceFET{
		doc: doc,
	}
}

func MakeTimetableData(bd *base.BaseData) autotimetable.TtSource {
	logger := bd.Logger
	weightTable := MakeFetWeights()
	//newdoc := sfet.doc.Copy()
	//sourcefet := &TtSourceFet{doc: newdoc}
	sourcefet := &TtSourceFet{}
	//fmt.Printf("sourcefet.WeightTable = %+v\n\n", sourcefet.WeightTable)
	fetroot := bd.Source.(*SourceFET).doc.Root()

	/*
	   fmt.Printf("ROOT element: %s (%+v)\n", root.Tag, root.Attr)
	   for i, e := range root.ChildElements() {
	       fmt.Printf(" -- %02d: %s\n", i, e.Tag)
	   }
	   fmt.Println("\n  --------------------------")
	*/

	// Get active activities, count inactive ones
	{
		//a_elements := []*etree.Element{}
		activities := []*ttActivity{}
		ael := fetroot.SelectElement("Activities_List")
		inactive := 0
		for _, a := range ael.ChildElements() {
			if a.SelectElement("Active").Text() == "true" {
				//a_elements = append(a_elements, a)
				activities = append(activities, &ttActivity{
					Id: a.SelectElement("Id").Text(),
					//TODO?
					// These are probably not needed if the back-end just uses a copy
					// of the FET source:
					//Tag:                string // optionally usable by the back-end,
					//Duration:           int,
					//Groups:             []*base.Group,
					//AtomicGroupIndexes: []AtomicIndex,
				})
			} else {
				inactive++
			}
		}
		if inactive != 0 {
			logger.Result("INACTIVE_ACTIVITIES", strconv.Itoa(inactive))
		}
		//sourcefet.activityElements = a_elements
		sourcefet.activities = activities
	}

	// Collect the constraints, dividing into soft and hard groups.
	// Inactive constraints will be removed.

	// Regexp to match constraint comment which has a number tag already:
	// Soft constraints also have a weight.
	r_constraint_number := regexp.MustCompile(`^\[[0-9]+.*\](.*)$`)

	hard_constraint_map := map[constraintType][]constraintIndex{}
	soft_constraint_map := map[constraintType][]constraintIndex{}
	constraint_types := []constraintType{}

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
			wdb := FetWeight2Db(w, weightTable)
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
			if inactive != 0 {
				logger.Result("INACTIVE_TIME_CONSTRAINTS", strconv.Itoa(inactive))
			}
		} else {
			sourcefet.s_constraints = tsindexes
			if inactive != 0 {
				logger.Result("INACTIVE_SPACE_CONSTRAINTS", strconv.Itoa(inactive))
			}
		}
	}

	//sourcefet.nConstraints = constraintIndex(len(sourcefet.constraintElements))
	sourcefet.constraintTypes = autotimetable.SortConstraintTypes(
		constraint_types, ConstraintPriority)
	sourcefet.hardConstraintMap = hard_constraint_map
	sourcefet.softConstraintMap = soft_constraint_map

	return sourcefet
}

// Generate a JSON version of the given element. Only a simple subset of
// XML is covered, but it should be enough for a FET file.
func WriteElement(e *etree.Element) string {
	k, v := jsonElement(e)
	if v == nil {
		return ""
	}
	m := map[string]any{}
	m[k] = v
	//jsonBytes, err := json.MarshalIndent(m, "", "   ")
	jsonBytes, err := json.Marshal(m)
	if err != nil {
		panic(err)
	}
	return string(jsonBytes)
}

func jsonElement(e *etree.Element) (string, any) {
	children := e.ChildElements()
	if len(children) == 0 {
		v := e.Text()
		if len(v) == 0 {
			return "", nil
		}
		return e.FullTag(), v
	}
	m0 := map[string][]any{}
	for _, c := range children {
		k, v := jsonElement(c)
		if v == nil {
			continue
		}
		m0[k] = append(m0[k], v)
	}
	if len(m0) == 0 {
		return "", nil
	}
	m := map[string]any{}
	for k, v := range m0 {
		if len(v) == 1 {
			m[k] = v[0]
		} else {
			m[k] = v
		}
	}
	return e.FullTag(), m
}
