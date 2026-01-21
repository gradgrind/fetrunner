# Timetable View

The primary timetable view consists of a table with days and time-slots as the axes (and headers), with a choice of orientation. A time-slot will be divided when showing parallel activities, especially when a class is divided into groups having different activities. Each activity is represented by a **tile**, which can cover all or part of one or more time-slots.

**For vertical days orientation**

The total width is: time_slot_width * number_of_time_slots per_day + vertical_header_width (+ borders).

The total height is: time_slot_height * number_of_days + horizontal_header_height (+ borders).

**For horizontal days orientation**

The total width is: time_slot_width * number_of_days + vertical_header_width (+ borders).

The total height is: time_slot_height * number_of_time_slots per_day + horizontal_header_height (+ borders).

Through the **View Scale Factor** the view can adapt to the size of the display window. There should be a standard size, corresponding to the text size in the rest of the interface. This should provide easily legible text with no need for enlargement. However, it is quite possible that the whole timetable then doesn't fit into the display window, causing scroll bars to appear. It should then be possible to reduce the timetable size, perhaps with a special option to automatically fill one dimension, leaving a single scroll-bar.

How much text can be shown in a time-slot will depend on a number of factors, especially the number of divisions. The main data is subject/activity, class, group, teacher and room, some of which can comprise more than one item. Exactly which of these should be shown in a time-slot depends to some extent on the type of view (e.g. in a teacher-view, the teacher need not be shown, and divisions are only likely in a class view). As it is such a complicated matter, it should perhaps be handled by calling functions chosen according to certain criteria.

The complete data for an activity could be provided when the mouse hovers over a tile, or by a selection mechanism, being shown either as a tool-tip or in a dedicated display area..

This primary view would show the timetable for a single class, a single teacher or a single room. It would be nice to also have **overview** displays, showing the timetable for all classes, all teachers or all rooms. The possibility of showing (legible) text in such a view is of course very limited, but colour coding might help somewhat.

**TODO** As the question of which information to show, and how, is such a tricky one, it would probably be sensible to start simple (e.g. show a minimum of data per cell), but provide for later extension. For vertical days, the divisions would be vertical, so the data could be shown in a line; for horizontal days, the data could be shown in a column. From this a minimum size requirement could be determined (based on text height and width). A later step might involve a certain degree of text shrinkage. What constitutes "a minimum" needs to be determined. One possibility is not to show all items, another is to abbreviate lists to the first item + comma.

## Placement of tiles representing divisions

Basically, a tile should occupy a space proportionate to the number of atomic groups it covers. It should start at a position corresponding to the first of its (ordered) atomic groups. This will, however, not always be possible, for example when a tile's constituent atomic groups are not contiguous – it will then occupy the space of an atomic group belonging to a different tile, which must then be placed somewhere else.

Within a single time-slot it should not be too difficult to adapt to the situation, but when an activity occupies more than one time-slot it's tile should be contiguous. This means that it might affect the ordering of activities in subsequent time-slots. Perhaps an iterative procedure, re-ordering the previous time-slot if one proves impossible, might work. That would probably call for a determinate ordering of the possibilities for a time-slot.

To avoid deviating too far from the optimal division ordering, it might be best to limit the possible positions of a tile. Perhaps the ordering should be limited, so that a tile can be ordered and positioned (somehow!) either by its first or by its last atomic group. That would be 2^n possibilities for a time-slot with n activities.
