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

void TtView::set_teacher(int tix)
{
    new_grid();
    backend->op("TT_TEACHER_PLACEMENTS", QString::number(tix));
}

void TtView::do_TEACHER_PLACEMENT(const QString &val) {
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

void TtView::set_room(int rix) {
    new_grid();
    backend->op("TT_ROOM_PLACEMENTS", QString::number(rix));
}

void TtView::do_ROOM_PLACEMENT(const QString &val) {
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

void TtView::set_class(int cix) {
    new_grid();

    // Build an array for the week (days x hours), each slot
    // containing a list of tile_data items.
    auto ndays = grid->daylist.length();
    auto nhours = grid->hourlist.length();
    weekBuffer.clear();
    weekBuffer.resize(ndays);
    for (int i = 0; i < ndays; ++i) {
        weekBuffer[i].resize(nhours);
    }
    classAtomics = ttbase->get_class(cix).atomics;
    backend->op("TT_CLASS_PLACEMENTS", QString::number(cix));
}

void TtView::do_CLASS_PLACEMENT(const QString &val) {
    auto aix = ttbase->place_activity(val);
    auto a = ttbase->activities.at(aix);
    // Extract activity atomics for this class
    a->selected_atomics.clear();
    for (int agix : a->atomics) {
        if (classAtomics.contains(agix))
            a->selected_atomics.append(agix);
    }
    // Add to `weekBuffer` for each covered time slot
    for (int i = 0; i < a->length; ++i) {
        weekBuffer[a->day][a->hour + i].append(aix);
    }
}

//TODO: A better Tile placement scheme ...
// Assume the atomics are sorted (increasing).
void TtView::do_SetupClassView(const QString &val) {
    auto cix = val.toInt();
    auto ndays = grid->daylist.length();
    auto nhours = grid->hourlist.length();
    for (int d = 0; d < ndays; ++d) {
        for (int h = 0; h < nhours; ++h) {
            auto aixlist = weekBuffer.at(d).at(h);
            if (aixlist.length() > 1) {
                // Sort on activity atomics from this class.
                std::sort(aixlist.begin(), aixlist.end(),
                    [this](const int& aix1, const int& aix2) {
                    auto a1 = ttbase->activities.at(aix1);
                    auto a2 = ttbase->activities.at(aix2);
                    return (a1->selected_atomics[0] < a2->selected_atomics[0]);
                });
            }
            int offset = 0;
            for (int aix : aixlist) {
                auto a =  ttbase->activities.at(aix);
                int agix0 = a->selected_atomics.at(0);
                int o0 = classAtomics.indexOf(agix0);
                if (o0 > offset)
                    offset = o0;
                Tile *t = new Tile(grid, aix);
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
                t->length = 1;
                t->div0 = offset;
                t->divs = a->selected_atomics.length();
                t->ndivs = classAtomics.length();
                // Place in grid
                grid->place_tile(t, a->day, a->hour);
                offset += t->divs;
            }
        }
    }
}
