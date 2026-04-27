#include "backend.h"
#include <QMap>
//#include <qdebug.h>
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
    //qDebug() << "?" << cmd;
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
        //qDebug() << "=" << logline;
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
