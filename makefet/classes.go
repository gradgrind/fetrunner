package makefet

import (
	"fetrunner/base"
	"strconv"
)

func (fetbuild *FetBuild) set_classes() {
	ids := []IdPair{}
	tt_data := fetbuild.ttdata
	db0 := tt_data.Db
	fetyears := fetbuild.fetroot.CreateElement("Students_List")
	for _, cdiv := range tt_data.ClassDivisions {
		cl := cdiv.Class
		cname := cl.GetTag()
		// Skip "special" classes.
		if cname == "" {
			continue
		}

		fetyear := fetyears.CreateElement("Year")
		fetyear.CreateElement("Name").SetText(cname)
		fetyear.CreateElement("Long_Name").SetText(cl.Name)

		ids = append(ids, IdPair{Source: string(cl.GetRef()), Backend: cname})

		// Construct the "Categories" (divisions)
		divs := cdiv.Divisions
		fetyear.CreateElement("Number_of_Categories").SetText(strconv.Itoa(len(divs)))
		fetyear.CreateElement("Separator").SetText(CLASS_GROUP_SEP)
		for _, divl := range divs {
			fetdiv := fetyear.CreateElement("Category")
			fetdiv.CreateElement("Number_of_Divisions").SetText(strconv.Itoa(len(divl)))
			for _, i := range divl {
				fetdiv.CreateElement("Division").SetText(db0.Elements[i].GetTag())
			}
		}

		// Construct the Groups and Subgroups
		for _, div := range divs {
			for _, gref := range div {
				// Need to construct group name with class, group
				// and CLASS_GROUP_SEP
				g := fetGroupTag(db0.Elements[gref].(*base.Group))
				fetgroup := fetyear.CreateElement("Group")
				fetgroup.CreateElement("Name").SetText(g)

				for _, agix := range tt_data.AtomicGroupIndex[gref] {
					ag := tt_data.AtomicGroups[agix].GetResourceTag()
					fetsubgroup := fetgroup.CreateElement("Subgroup")
					fetsubgroup.CreateElement("Name").SetText(ag)
				}
			}
		}
	}
	fetbuild.rundata.ClassIds = ids
}

// In FET the group identifier is constructed from the class tag,
// CLASS_GROUP_SEP and the group tag. However, if the group is the
// whole class, just the class tag is used.
func fetGroupTag(g *base.Group) string {
	gt := g.Class.Tag
	if g.Tag != "" {
		gt += CLASS_GROUP_SEP + g.Tag
	}
	return gt
}
