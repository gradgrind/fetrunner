#include "ttbase.h"
#include "ttgrid.h"
#include <QJsonObject>

void ShowTeacher(TtGrid *grid, TtBase * ttbase, int teacher_id)
{
    auto plist = get_item_placements("TT_TEACHER_PLACEMENTS", teacher_id);
    for (const auto p : plist) {
        auto tiledata = ttbase->get_tile_data(p);
        Tile *t = new Tile(
            grid,
            QJsonObject{
                {"TEXT", tiledata->groups.join(",")},
                {"TL", tiledata->subject},
                {"BR", tiledata->rooms.join(",")},
                {"LENGTH", tiledata->length},
                {"DIV0", 0},
                {"DIVS", 1},
                {"NDIVS", 1},
            },
            p->activity);
        grid->place_tile(t, p->day, p->hour);
    }
}
