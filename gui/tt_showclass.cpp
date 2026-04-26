#include "tt_show_resource.h"
#include <QJsonObject>

struct tile_data {
    QString subject;
    QString teachers;
    QString groups;
    QString rooms;
    int natomics;
    int index;
    int activity;
};

//TODO: A better Tile placement scheme!
void ShowClass(TtGrid *grid, TtBase * ttbase, int cix)
{
    auto total = ttbase->get_class(cix).atomics.length();
    // Build an array for the week (days x hours), each slot
    // containing a list of tile_data items.
    auto ndays = grid->daylist.length();
    auto nhours = grid->hourlist.length();
    QList<QList<QList<tile_data>>> week(ndays);
    for (int i = 0; i < ndays; ++i) {
        week[i].resize(nhours);
    }
    // Place single-cell activity Tiles in this week
    const TtPlacementList plist("TT_CLASS_PLACEMENTS", cix);
    for (const auto p : plist) {
        const auto tiledata = ttbase->get_tile_data(p);
        int fraction = tiledata->atomics.length();
        auto groups = tiledata->groups.join(",");
        auto teachers = tiledata->teachers.join(",");
        auto rooms = tiledata->rooms.join(",");
        //TODO: if fractions too small, simplify the data.
        for (int i = 0; i < tiledata->length; ++i) {
            auto data = tile_data{
                .subject = tiledata->subject,
                .teachers = teachers,
                .groups = groups,
                .rooms = rooms,
                .natomics = fraction,
                .index = i,
                .activity = p->activity
            };
            week[p->day][p->hour + i].append(data);
        }
    }
    for (int d = 0; d < ndays; ++d) {
        for (int h = 0; h < nhours; ++h) {
            auto olist = week.at(d).at(h);
            if (olist.length() > 1) {
                // Sort on groups.
                std::sort(olist.begin(), olist.end(), [](const tile_data& o1, const tile_data& o2) {
                    return (o1.groups < o2.groups);
                });
            }
            int offset = 0;
            for (const auto &o : olist) {
                auto n = o.natomics;
                Tile *t = new Tile(grid,
                    QJsonObject{
                        {"TEXT", o.subject},
                        {"TL", o.teachers},
                        {"TR", o.groups},
                        {"BR", o.rooms},
                        //{"LENGTH", }, // default = 1
                        {"DIV0", offset},
                        {"DIVS", n},
                        {"NDIVS", total},
                    },
                    o.activity);
                offset += n;
                grid->place_tile(t, d, h);
            }
        }
    }
}
