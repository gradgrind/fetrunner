#include "ttbase.h"
#include "backend.h"

TtBase::TtBase() {}

void TtBase::set_activities()
{
    clear_activities();
    auto alist = backend->op("TT_ACTIVITIES");
    for (const auto &[k, v] : std::as_const(alist)) {
        if (k != "TT_ACTIVITIES")
            continue;
        auto vlist = v.split(":");
        QList<int> tlist;
        for (const auto &t : vlist.at(2).split(",")) {
            tlist.append(t.toInt());
        }
        QList<int> aglist;
        for (const auto &ag : vlist.at(3).split(",")) {
            aglist.append(ag.toInt());
        }
        QStringList glist;
        for (const auto &g : vlist.at(4).split(",")) {
            glist.append(g);
        }
        activities.append(new TtActivity{
            .length = vlist.at(0).toInt(),
            .subject = vlist.at(1),
            .teachers = tlist,
            .atomics = aglist,
            .groups = glist});
    }
}

void TtBase::set_classes()
{
    classes.clear();
    auto alist = backend->op("TT_CLASSES");
    for (const auto &[k, v] : std::as_const(alist)) {
        if (k != "TT_CLASSES")
            continue;
        auto vlist = v.split(":");
        auto name = vlist.at(1);
        if (name.isEmpty())
            name = vlist.at(0);
        QList<int> aglist;
        for (const auto &i : vlist.at(2).split(",")) {
            aglist.append(i.toInt());
        }
        classes.append(TtClass{vlist.at(0), name, aglist, vlist.at(3).split(",")});
    }
}

const TtClass & TtBase::get_class(int cix)
{
    return classes[cix];
}

void TtBase::set_teachers()
{
    teachers.clear();
    auto alist = backend->op("TT_TEACHERS");
    for (const auto &[k, v] : std::as_const(alist)) {
        if (k != "TT_TEACHERS")
            continue;
        auto vlist = v.split(":");
        auto name = vlist.at(1);
        if (name.isEmpty())
            name = vlist.at(0);
        teachers.append(TtName{vlist.at(0), name});
    }
}

void TtBase::set_rooms()
{
    rooms.clear();
    auto alist = backend->op("TT_ROOMS");
    for (const auto &[k, v] : std::as_const(alist)) {
        if (k != "TT_ROOMS")
            continue;
        auto vlist = v.split(":");
        auto name = vlist.at(1);
        if (name.isEmpty())
            name = vlist.at(0);
        rooms.append(TtName{vlist.at(0), name});
    }
}

const QList<TtActivity *> TtBase::get_activities()
{
    return activities;
}

TtPlacementList::TtPlacementList (QString cmd, int item) : QList<TtPlacement *>()
{
    auto plist = backend->op(cmd, {QString::number(item)});
    for (const auto &[k, v] : std::as_const(plist)) {
        if (k != "PLACEMENT")
            continue;
        auto vlist = v.split(":");
        QList<int> rlist;
        for (const auto &r : vlist.at(3).split(",")) {
            rlist.append(r.toInt());
        }
        append(new TtPlacement{
            .activity = vlist.at(0).toInt(),
            .day = vlist.at(1).toInt(),
            .hour = vlist.at(2).toInt(),
            .rooms = rlist});
    }
}

TileData *TtBase::get_tile_data(TtPlacement *p)
{
    auto a = activities.at(p->activity);
    QStringList tlist;
    for (const auto &tix : std::as_const(a->teachers)) {
        tlist.append(teachers.at(tix).tag);
    }
    QStringList rlist;
    for (const auto &rix : std::as_const(p->rooms)) {
        rlist.append(rooms.at(rix).tag);
    }
    return new TileData{
        .length = a->length,
        .subject = a->subject,
        .teachers = tlist,
        .rooms = rlist,
        .atomics = a->atomics,
        .groups = a->groups};
}