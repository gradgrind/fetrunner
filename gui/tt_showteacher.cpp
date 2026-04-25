#include "tt_showteacher.h"
#include <QJsonObject>
#include "backend.h"

//TODO: Consider having just activity index in the placements, and
// fetching all activity data, so that also unplaced activities can
// be handled.

ShowTeacher::ShowTeacher(TtGrid *grid, int teacher_id)
{
    auto plist = get_item_placements("TT_TEACHER_PLACEMENTS", teacher_id);
    //TODO
    for (const auto p : plist) {
        // Needs room, teacher, etc. vectors because just the indexes are supplied.
        auto rooms = p->rooms;

        Tile *t = new Tile( //
            grid,
            QJsonObject{
            //
            {"TEXT", p->groups.join(",")},
            {"TL", p->subject},
            {"BR", p->rooms.join(",")},
            {"LENGTH", p->length},
            {"DIV0", 0},
            {"DIVS", 1},
            {"NDIVS", 1},
            },
            lid);
        grid->place_tile(t, p->day, p->hour);
    }
}
