#include "backend.h"
#include <QMap>
#include "../libfetrunner/libfetrunner.h"
//#include <iostream>

// display colours for the log
QMap<QString, QColor> colours{{"*INFO*", "#009000"},
                              {"*WARNING*", "#eb8900"},
                              {"*ERROR*", "#d00000"},
                              {"+++", "#000000"},
                              {"***", "#000000"},
                              {"---", "#000000"},
                              {"$", "#53a0ff"}};

Backend::Backend()
    : QObject()
{}
Backend *backend;

QList<KeyVal> Backend::op(QString cmd, QStringList data)
{
    if (!data.empty()) {
        cmd += "|" + data.join("|");
    }
    FetRunnerCommand(cmd.toUtf8().data());

    // Collect log up to "---" or "***"
    QList<KeyVal> results;
    QStringList errors;
    while (true) {
        auto key_val = readlogline();
        auto key = key_val.key;
        if (key == "+++")
            continue;
        if (key == "---" || key == "***")
            break;
        auto val = key_val.val;
        if (key == "*ERROR*") {
            errors.append(val);
            continue;
        }
        if (key == "$") {
            // a result
            auto rkv = readresult(val);
            if (rkv.key.isEmpty()) {
                errors.append(rkv.val);
            } else {
                results.append(rkv);
            }
            //continue;
        }
    }
    if (!errors.empty()) {
        if (errors.length() > 5) {
            auto elist = errors;
            errors = QStringList();
            errors << elist[0] << elist[1] << elist[2] << elist[3] << elist[4];
            errors << "...";
        }
        emit error(errors.join("\n"));
    }
    return results;
}

KeyVal Backend::readresult(QString r)
{
    auto n = r.indexOf('=');
    if (n < 0)
        return KeyVal{"", QString{"BUG in backend result: "} + logline};
    auto rkey = r.left(n);
    auto rval = r.right(r.length() - n - 1);
    return KeyVal{rkey, rval};
}

KeyVal Backend::readlogline()
{
    while (true) {
        logline = QString(FetRunnerReadLog());
        if (logline.length() != 0 && logline.at(0) != " ")
            break;
        emit log(logline); // write to log without change of colour
    }
    auto i = logline.indexOf(" ");
    QString msgtype, msgrest;
    if (i < 0) {
        // there is only the type
        msgtype = logline;
        msgrest = "";
    } else {
        // split into message-type and rest
        msgtype = logline.left(i);
        msgrest = logline.right(logline.length() - i - 1);
    }
    // The type determines the display colour.
    emit logcolour(colours.value(msgtype, QColor{0x76, 0x5e, 0xff}));
    emit log(logline.replace("||", "\n + ")); // write to log
    return KeyVal{msgtype, msgrest};
}

// Run an op, expect a single result whose key may be specified.
KeyVal Backend::op1(
    QString cmd, QStringList data, QString key)
{
    auto results = op(cmd, data);
    if (results.length() == 1) {
        auto kv = results[0];
        if (key.isEmpty() || key == kv.key)
            return kv;
    }
    return {};
}

void get_tt_activities()
{
    QList<Activity *> activity_list;
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
        activity_list.append(new Activity{//
                                          .length = vlist.at(0).toInt(),
                                          .subject = vlist.at(1),
                                          .teachers = tlist,
                                          .atomics = aglist,
                                          .groups = glist});
    }
    tt_activities = activity_list;
}

QList<Placement *> get_item_placements(QString cmd, int item)
{
    QList<Placement *> placements;
    auto plist = backend->op(cmd, {QString::number(item)});
    for (const auto &[k, v] : std::as_const(plist)) {
        if (k != "PLACEMENT")
            continue;
        auto vlist = v.split(":");
        QList<int> rlist;
        for (const auto &r : vlist.at(3).split(",")) {
            rlist.append(r.toInt());
        }
        placements.append(new Placement{//
                                        .activity = vlist.at(0).toInt(),
                                        .day = vlist.at(1).toInt(),
                                        .hour = vlist.at(2).toInt(),
                                        .rooms = rlist});
    }
    return placements;
}
