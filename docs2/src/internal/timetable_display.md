# Displaying a timetable

There are various ways in which timetables may be displayed, various "views" onto the data. Typically timetables will be displayed for a class, a teacher or a room, sometimes as an overview for all classes, teachers or rooms.

A teacher or room can typically only be associated with a single activity at a time, which allows a relatively straightforward display of the corresponding timetable. There may be exceptions to this rule, but I would rather ignore that possibility for now (at some point, though, it should receive some consideration).

A class, however, may be subdivided into groups having at least partly differing timetables. The question then arises of how to show the potentially multiple activities in each time slot. There may be multiple ways to divide a class, which can lead to further complications.

Perhaps the most obvious difficulty with displaying multiple activities in a single time slot is that the individual activity "tiles" will need to be smaller, limiting – perhaps severely – the space for showing the details of the activities. Also the positioning of the group tiles within a slot can be difficult, especially when a class can be divided in different wasy.

For some division-based groupings where there are not too many groups in a division, the multiple tiles per time slot approach might well be the neatest, but further difficulties can arise with multi-slot activities in obscure cases where different divisions are used in the covered slots. In short, it is quite hard to develop a display algorithm which can cover all possible group configurations, and `FET` is very flexible as far as defining groups and subgroups is concerned.

An approach which avoids the multiple-tiles-per-slot problem completely would be to base the display on an atomic group rather than a class. This would be a view for the set of students belonging to the displayed atomic group. Such a group view could be designated by the intersecting set of student-groups containing this atomic group. Although this would mean rather a lot of views for classes with many atomic groups, it seems to be the most straightforward way of handling the problem. It may be desirable to have a two-tier menu (selecting first class and then atomic group) rather than presenting a long list of combined class-group entries. Selecting the class would present a multiple-tiles-per-slot view, selecting an atomic group its single-tile per slot view.

Note that if there are redundant atomic groups, there will be some identical timetables. Rather than trying to detect and suppress duplicates, it might be better to display them all and leave it up to the user to remove the superfluous groups/divisions.

Depending on the definitions of groups and subgroups in `FET`, it may be difficult to determine which set of intersecting groups should be used to designate an atomic group. By restricting the groups used to those which are mentioned in the activities can help, as can simple one-to-one correspondences where a used group corresponds to a single atomic group.

Perhaps hovering could cause fuller information to appear?
