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

    // For fractional tile sizes, number of atomic groups, so the
    // number of atomic groups for the whole class is needed, too.
    a->offset = vlist.at(4).toInt();
    a->size = vlist.at(5).toInt(); // 0 => full cell

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
    auto cdata = ttbase->get_class(cix);
    classAtomics = cdata.atomics;
    backend->op("TT_CLASS_PLACEMENTS", QString::number(cix));
    emit notifier->switch_logger("", 0);
}

void TtView::do_CLASS_PLACEMENT(const QString &val) {
    int ndivs = classAtomics.length();
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
    t->div0 = a->offset;
    t->divs = a->size;
    t->ndivs = ndivs;
    // Place in grid
    qDebug() << "PLACE" << a->subject << a->day << a->hour << ndivs << a->offset << a->size;
    grid->place_tile(t, a->day, a->hour);
}
