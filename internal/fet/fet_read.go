package fet

import (
	"encoding/json"
	"fetrunner/internal/base"
	"strconv"

	"github.com/beevik/etree"
)

func (s *TtSourceFet) SourceType() string {
	return "FET"
}

// In FET there are "time" constraints and "space" constraints. They are
// all lumped together in the `ConstraintElements` list.

func FetRead(
	bdata *base.BaseData,
	fetpath string,
) *TtSourceFet {
	doc := etree.NewDocument()
	if err := doc.ReadFromFile(fetpath); err != nil {
		base.LogError("%s", err)
		return nil
	}
	weightTable := MakeFetWeights()
	sourcefet := &TtSourceFet{
		doc:         doc,
		weightTable: weightTable,
	}
	fetroot := doc.Root()
	sourcefet.read_elements(fetroot)
	// Collect the activities, inactive ones will be ignored.
	if inactive := sourcefet.read_activities(fetroot); inactive != 0 {
		base.LogResult("INACTIVE_ACTIVITIES", strconv.Itoa(inactive))
	}
	{
		// Collect the constraints, dividing into soft and hard groups.
		// Inactive constraints will be ignored.
		t_inactive, s_inactive := sourcefet.read_constraints(fetroot)
		if t_inactive != 0 {
			base.LogResult("INACTIVE_TIME_CONSTRAINTS", strconv.Itoa(t_inactive))
		}
		if s_inactive != 0 {
			base.LogResult("INACTIVE_SPACE_CONSTRAINTS", strconv.Itoa(s_inactive))
		}
	}
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
