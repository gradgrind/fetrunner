#include "ttbase.h"
#include "backend.h"
#include <qdebug.h>

TtBase::TtBase() {
    backend->registerResultHandler("TT_DAYS",
        [this](QString arg) {set_day(arg);});
    backend->registerResultHandler("TT_HOURS",
        [this](QString arg) {set_hour(arg);});
    backend->registerResultHandler("TT_CLASSES",
        [this](QString arg) {set_class(arg);});
    backend->registerResultHandler("TT_TEACHERS",
        [this](QString arg) {set_teacher(arg);});
    backend->registerResultHandler("TT_ROOMS",
        [this](QString arg) {set_room(arg);});
    backend->registerResultHandler("TT_ACTIVITIES",
        [this](QString arg) {set_activity(arg);});

    days.clear();
    backend->op("TT_DAYS");
    hours.clear();
    backend->op("TT_HOURS");
    //qDebug() << "set_classes()";
    classes.clear();
    backend->op("TT_CLASSES");
    //qDebug() << "set_teachers()";
    teachers.clear();
    backend->op("TT_TEACHERS");
    //qDebug() << "set_rooms()";
    rooms.clear();
    backend->op("TT_ROOMS");
    //qDebug() << "set_activities()";
    clear_activities();
    backend->op("TT_ACTIVITIES");
    //qDebug() << "end TtBase()";
}

void TtBase::set_day(const QString &val) {
    auto vlist = val.split(':');
    auto name = vlist.at(1);
    if (name.isEmpty())
        name = vlist.at(0);
    days.append(TtName{vlist.at(0), name});
}

void TtBase::set_hour(const QString &val) {
    auto vlist = val.split(":");
    auto name = vlist.at(1);
    if (name.isEmpty())
        name = vlist.at(0);
    hours.append(TtName{vlist.at(0), name});
}

void TtBase::set_activity(const QString &val)
{
    auto vlist = val.split(":");
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
        .day = -1,
        .hour = -1,
        .subject = vlist.at(1),
        .teachers = tlist,
        .atomics = aglist,
        .groups = glist});
}

void TtBase::set_class(const QString &val)
{
    auto vlist = val.split(":");
    auto name = vlist.at(1);
    if (name.isEmpty())
        name = vlist.at(0);
    QList<int> aglist;
    for (const auto &i : vlist.at(2).split(",")) {
        aglist.append(i.toInt());
    }
    classes.append(TtClass{vlist.at(0), name, aglist, vlist.at(3).split(",")});
}

const TtClass & TtBase::get_class(int cix)
{
    return classes[cix];
}

void TtBase::set_teacher(const QString &val)
{
    auto vlist = val.split(":");
    auto name = vlist.at(1);
    if (name.isEmpty())
        name = vlist.at(0);
    teachers.append(TtName{vlist.at(0), name});
}

void TtBase::set_room(const QString &val)
{
    auto vlist = val.split(":");
    auto name = vlist.at(1);
    if (name.isEmpty())
        name = vlist.at(0);
    rooms.append(TtName{vlist.at(0), name});
}

const QList<TtActivity *> TtBase::get_activities()
{
    return activities;
}

/*
TtPlacementList::TtPlacementList(QString cmd, int item) : QList<TtPlacement *>()
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
*/
