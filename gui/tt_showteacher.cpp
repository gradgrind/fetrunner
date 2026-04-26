#include "tt_show_resource.h"
#include <QJsonObject>

void ShowTeacher(TtGrid *grid, TtBase * ttbase, int tix)
{
    const TtPlacementList plist("TT_TEACHER_PLACEMENTS", tix);
    for (const auto p : plist) {
        const auto tiledata = ttbase->get_tile_data(p);
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
