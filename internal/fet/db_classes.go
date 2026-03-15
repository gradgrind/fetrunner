package fet

import (
	"fetrunner/internal/base"
)

func (fetbuild *fet_build) set_classes() {
	source := fetbuild.ttsource
	aglist := source.GetAtomicGroups() // needed for Subgroups
	fetyears := fetbuild.fetroot.CreateElement("Students_List")
	for _, cl := range source.GetClasses() {
		cname := cl.Tag
		// Skip "special" classes.
		if cname == "" {
			continue
		}
		fetyear := fetyears.CreateElement("Year")
		fetyear.CreateElement("Name").SetText(cname)

		// Construct the Groups and Subgroups
		for _, g := range cl.Groups {
			// Need to construct group name with class, CLASS_GROUP_SEP
			// and group
			fetgroup := fetyear.CreateElement("Group")
			fetgroup.CreateElement("Name").SetText(cl.Tag + CLASS_GROUP_SEP + g.Tag)
			for _, agix := range g.AtomicIndexes {
				fetsubgroup := fetgroup.CreateElement("Subgroup")
				fetsubgroup.CreateElement("Name").SetText(aglist[agix])
			}
		}
	}
}

// TODO???
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
