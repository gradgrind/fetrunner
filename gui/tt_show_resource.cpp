#include "backend.h"
#include "globals.h"
#include "ttview.h"

int TtBase::place_activity(const QString &val) {
    auto vlist = val.split(":");
    QList<int> rlist;
    auto rt = vlist.at(3);
    if (!rt.isEmpty()) {
        for (const auto &r : rt.split(",")) {
            rlist.append(r.toInt());
        }
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
    emit notifier->switch_logger(">>> --TIMETABLE_TEACHER", 3);
    emit notifier->clear_log(3);
    backend->op("TT_TEACHER_PLACEMENTS", QString::number(tix));
    emit notifier->switch_logger("", 0);
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
    emit notifier->switch_logger(">>> --TIMETABLE_ROOM", 3);
    emit notifier->clear_log(3);
    backend->op("TT_ROOM_PLACEMENTS", QString::number(rix));
    emit notifier->switch_logger("", 0);
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
    emit notifier->switch_logger(">>> --TIMETABLE_CLASS", 3);
    emit notifier->clear_log(3);

    // Build an array for the week (days x hours), each slot
    // containing a list of tile_data items.
    auto ndays = grid->daylist.length();
    auto nhours = grid->hourlist.length();
    weekBuffer.clear();
    weekBuffer.resize(ndays);
    for (int i = 0; i < ndays; ++i) {
        weekBuffer[i].resize(nhours);
    }
    auto cdata = ttbase->get_class(cix);
    classAtomics = cdata.atomics;
    classGroups = cdata.groups.keys();
    backend->op("TT_CLASS_PLACEMENTS", QString::number(cix));
    setupClassView();
    emit notifier->switch_logger("", 0);
}

void TtView::do_CLASS_PLACEMENT(const QString &val) {
    auto aix = ttbase->place_activity(val);
    auto a = ttbase->activities.at(aix);
    // Extract activity atomics and groups for this class
    a->selected_atomics.clear();
    a->selected_groups.clear();
    for (int agix : std::as_const(a->atomics)) {
        if (classAtomics.contains(agix))
            a->selected_atomics.append(agix);
    }
    for (const QString &g : std::as_const(a->groups)) {
        if (classGroups.contains(g)) {
            a->selected_groups.append(g);
        }
    }
    // Add to `weekBuffer` for each covered time slot
    for (int i = 0; i < a->length; ++i) {
        weekBuffer[a->day][a->hour + i].append({aix, i});
    }
}

//TODO: A better Tile placement scheme ...
// Can class divisions be determined from the atomic groups?
// If so, perhaps the offsets can be derived from these?
// Can long activities for non-class groups be rejoined?
// It might be better to do at least some of this processing
// in the (Go) back-end.
//
// Assume the atomics are sorted (increasing).
//void TtView::setupClassView() {
void TtView::setupClassView2() {
    auto ndays = grid->daylist.length();
    auto nhours = grid->hourlist.length();
    for (int d = 0; d < ndays; ++d) {
        for (int h = 0; h < nhours; ++h) {
            auto aixilist = weekBuffer.at(d).at(h);

            //???
            if (aixilist.empty()) continue;

            //TODO: Get tile list, splitting activities if >1 group
            QList<TtActivityTile> tile_list;
            for (const auto &aixi : aixilist) {
                auto a1 = ttbase->activities.at(aixi.activity);
                for (const auto &g : std::as_const(a1->selected_groups)) {
                    TtActivityTile a2{g, aixi.activity, aixi.index, -1, -1};
                    //TODO? The last elements are not set yet. The `offset` depends
                    // on the division. The `n_atomics` could perhaps be set here?
                    tile_list.append(a2);
                }
            }

            //TODO:
            // - determine the possible divisions and thus the offsets for each tile
            // - for length > 1 check subsequent slots; if there is a mismatch, eliminate
            //   division choice and move to next; if no next, fail?
            // - that process is iterative, including activities with length > 1 from
            //   following slots
            //???
            if (tile_list.length() == 1) {
                //TODO: Single tile at specified offset with fractionfrom size and total


            }

            //???
            if (aixilist.length() > 1) {
                // Sort on activity atomics from this class.
                std::sort(aixilist.begin(), aixilist.end(),
                    [this](const auto & aixi1, const auto & aixi2) {
                    auto a1 = ttbase->activities.at(aixi1.activity);
                    auto a2 = ttbase->activities.at(aixi2.activity);
                    return (a1->selected_atomics[0] < a2->selected_atomics[0]);
                });
            }
            int offset = 0;
            for (const auto &aixi : aixilist) {
                auto a =  ttbase->activities.at(aixi.activity);
                int divs = a->selected_atomics.length();
                int ndivs = classAtomics.length();
                int length = a->length;
                // "Rejoin" split activities if they are for the whole class
                if (length > 1) {
                    if (divs == ndivs) {
                        if (aixi.index != 0) continue;
                    } else {
                        length = 1;
                    }
                }
                int agix0 = a->selected_atomics.at(0);
                int o0 = classAtomics.indexOf(agix0);
                if (o0 > offset)
                    offset = o0;
                Tile *t = new Tile(grid, aixi.activity, aixi.index);
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
                t->length = length;
                t->div0 = offset;
                t->divs = divs;
                t->ndivs = ndivs;
                // Place in grid
                grid->place_tile(t, a->day, a->hour + aixi.index);
                offset += t->divs;
            }
        }
    }
}

// Assume the atomics are sorted (increasing).
void TtView::setupClassView() {
    auto ndays = grid->daylist.length();
    auto nhours = grid->hourlist.length();
    for (int d = 0; d < ndays; ++d) {
        for (int h = 0; h < nhours; ++h) {
            auto aixilist = weekBuffer.at(d).at(h);
            if (aixilist.length() > 1) {
                // Sort on activity atomics from this class.
                std::sort(aixilist.begin(), aixilist.end(),
                    [this](const auto & aixi1, const auto & aixi2) {
                    auto a1 = ttbase->activities.at(aixi1.activity);
                    auto a2 = ttbase->activities.at(aixi2.activity);
                    return (a1->selected_atomics[0] < a2->selected_atomics[0]);
                });
            }
            int offset = 0;
            for (const auto &aixi : aixilist) {
                auto a =  ttbase->activities.at(aixi.activity);
                int divs = a->selected_atomics.length();
                int ndivs = classAtomics.length();
                int length = a->length;
                // "Rejoin" split activities if they are for the whole class
                if (length > 1) {
                    if (divs == ndivs) {
                        if (aixi.index != 0) continue;
                    } else {
                        length = 1;
                    }
                }
                int agix0 = a->selected_atomics.at(0);
                int o0 = classAtomics.indexOf(agix0);
                if (o0 > offset)
                    offset = o0;
                Tile *t = new Tile(grid, aixi.activity, aixi.index);
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
                t->length = length;
                t->div0 = offset;
                t->divs = divs;
                t->ndivs = ndivs;
                // Place in grid
                grid->place_tile(t, a->day, a->hour + aixi.index);
                offset += t->divs;
            }
        }
    }
}
