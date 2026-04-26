#include "tt_show_resource.h"
#include <QJsonObject>

void ShowRoom(TtGrid *grid, TtBase * ttbase, int rix)
{
    const TtPlacementList plist("TT_ROOM_PLACEMENTS", rix);
    for (const auto p : plist) {
        const auto tiledata = ttbase->get_tile_data(p);
        Tile *t = new Tile(
            grid,
            QJsonObject{
                {"TEXT", tiledata->groups.join(",")},
                {"TL", tiledata->subject},
                {"BR", tiledata->teachers.join(",")},
                {"LENGTH", tiledata->length},
                {"DIV0", 0},
                {"DIVS", 1},
                {"NDIVS", 1},
            },
            p->activity);
        grid->place_tile(t, p->day, p->hour);
    }
}
