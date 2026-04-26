#include "ttbase.h"
#include "backend.h"

TtBase::TtBase() {}

void TtBase::set_tt_activities()
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
        tt_activities.append(new TtActivity{//
                                            .length = vlist.at(0).toInt(),
                                            .subject = vlist.at(1),
                                            .teachers = tlist,
                                            .atomics = aglist,
                                            .groups = glist});
    }
}

const QList<TtActivity *> TtBase::get_tt_activities()
{
    return tt_activities;
}

const TtPlacementList get_item_placements(QString cmd, int item)
{
    TtPlacementList placements;
    auto plist = backend->op(cmd, {QString::number(item)});
    for (const auto &[k, v] : std::as_const(plist)) {
        if (k != "PLACEMENT")
            continue;
        auto vlist = v.split(":");
        QList<int> rlist;
        for (const auto &r : vlist.at(3).split(",")) {
            rlist.append(r.toInt());
        }
        placements.append(new TtPlacement{//
                                          .activity = vlist.at(0).toInt(),
                                          .day = vlist.at(1).toInt(),
                                          .hour = vlist.at(2).toInt(),
                                          .rooms = rlist});
    }
    return placements;
}
