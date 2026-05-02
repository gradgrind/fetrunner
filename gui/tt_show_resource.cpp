#include "tt_show_resource.h"
#include "backend.h"
#include "ttview.h"

int TtBase::place_activity(const QString &val) {
    auto vlist = val.split(":");
    QList<int> rlist;
    for (const auto &r : vlist.at(3).split(",")) {
        rlist.append(r.toInt());
    }
    auto aix = vlist.at(0).toInt();
    auto a = activities[aix];
    a->day = vlist.at(1).toInt();
    a->hour = vlist.at(2).toInt();
    a->rooms = rlist;
    return aix;
}

void showTeacher(TtGrid *grid, TtBase * ttbase, int tix) {
    backend.op("TT_TEACHER_PLACEMENTS", QString::number(tix));
}

void TtView::do_TEACHER_PLACEMENTS(const QString &val) {
    auto aix = ttbase->place_activity(val);
    Tile *t = new Tile(grid, aix);
    // Set fields
    auto a = ttbase->activities.at(aix);
    //QStringList tlist;
    //for (const auto &tix : std::as_const(a->teachers)) {
    //    tlist.append(ttbase->teachers.at(tix).tag);
    //}
    QStringList rlist;
    for (const auto &rix : std::as_const(a->rooms)) {
        rlist.append(ttbase->rooms.at(rix).tag);
    }
    t->middle = a->groups.join(",");
    t->tl = a->subject;
    t->br = rlist.join(",");
    t->length = a->length;
    t->div0 = 0;
    t->divs = 1;
    t->ndivs = 1;
    // Place in grid
    grid->place_tile(t, a->day, a->hour);
}

void showRoom(TtGrid *grid, TtBase * ttbase, int rix) {
    backend.op("TT_ROOM_PLACEMENTS", QString::number(rix));
}

void TtView::do_ROOM_PLACEMENTS(const QString &val) {
    auto aix = ttbase->place_activity(val);
    Tile *t = new Tile(grid, aix);
    // Set fields
    auto a = ttbase->activities.at(aix);
    QStringList tlist;
    for (const auto &tix : std::as_const(a->teachers)) {
        tlist.append(ttbase->teachers.at(tix).tag);
    }
    //QStringList rlist;
    //for (const auto &rix : std::as_const(a->rooms)) {
    //    rlist.append(ttbase->rooms.at(rix).tag);
    //}
    t->middle = a->groups.join(",");
    t->tl = a->subject;
    t->br = tlist.join(",");
    t->length = a->length;
    t->div0 = 0;
    t->divs = 1;
    t->ndivs = 1;
    // Place in grid
    grid->place_tile(t, a->day, a->hour);
}

void showClass(TtGrid *grid, TtBase * ttbase, int cix) {
    backend.op("TT_ROOM_PLACEMENTS", QString::number(cix));
}

//TODO: A better Tile placement scheme!
void TtView::do_CLASS_PLACEMENTS(const QString &val) {
    auto aix = ttbase->place_activity(val);
    Tile *t = new Tile(grid, aix);
    // Set fields
    auto a = ttbase->activities.at(aix);
    QStringList tlist;
    for (const auto &tix : std::as_const(a->teachers)) {
        tlist.append(ttbase->teachers.at(tix).tag);
    }
    QStringList rlist;
    for (const auto &rix : std::as_const(a->rooms)) {
        rlist.append(ttbase->rooms.at(rix).tag);
    }
    t->middle = a->subject;
    t->tl = tlist.join(",");
    t->tr = a->groups.join(",");
    t->br = rlist.join(",");
    t->length = a->length;

    //TODO ...
    auto fraction = a->atomics.length();
    //TODO: Surely those are atomics from all classes in the groups ...
    auto total = ttbase->get_class(cix).atomics.length();

    t->div0 = 0;
    t->divs = fraction;
    t->ndivs = 1;
    // Place in grid
    grid->place_tile(t, a->day, a->hour);
}
/* old idea:
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
*/
