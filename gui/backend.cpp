#include "backend.h"
#include <QMap>
#include "../libfetrunner/libfetrunner.h"
#include <iostream>

QMap<QString, QColor> colours{
    // display colours for the log
    {"*INFO*", "#009000"},
    {"*WARNING*", "#eb8900"},
    {"*ERROR*", "#d00000"},
    {"+++", "#000000"},
    {"$", "#53a0ff"},
};

Backend::Backend()
    : QObject()
{}

QList<KeyVal> Backend::op(QString cmd, QStringList data)
{
    QList<KeyVal> results;
    auto darray = QJsonArray::fromStringList(data);
    emit log("+++ " + cmd + " [" + data.join(", ") + "]");
    //qDebug() << QString{"+++ "} + cmd << darray;
    QJsonObject cmdobj{{"Op", cmd}, {"Data", darray}};
    QJsonDocument doc(cmdobj);
    auto cs = doc.toJson(QJsonDocument::Compact);
    auto result = FetRunner(cs.data());
    std::cout << "<<<" << cs.data() << std::endl;
    std::cout << ">>>" << result << std::endl;
    auto jsondoc = QJsonDocument::fromJson(result);
    if (!jsondoc.isArray()) {
        // This needs to be thread-safe, so use a signal.
        emit error(QString{"BackendReturnError: "} + result);
    } else {
        QStringList errors;
        for (auto &&c : jsondoc.array()) {
            auto e = c.toObject();
            auto key = e["Type"].toString();
            auto val = e["Text"].toString();
            auto t0 = key + " " + val;
            emit logcolour(colours.value(key, QColor{0x76, 0x5e, 0xff}));
            emit log(t0);
            if (key == "*ERROR*") {
                errors.append(val);
                continue;
            }
            if (key == "$") {
                // a result
                auto n = val.indexOf('=');
                if (n < 0) {
                    errors.append(QString{"BUG in backend result: "} + t0);
                } else {
                    key = val.first(n);
                    val = val.sliced(n + 1);
                    results.append({key, val});
                    //qDebug() << "$$$" << key << "=" << val;
                }
                continue;
            }
            if (key == "+++") {
                //std::cout << "+++ " << qUtf8Printable(val) << std::endl;
                continue;
            }
            // else:
            //messages.append(key + " " + val);
            //std::cout << ">>> " << qUtf8Printable(key) << " " << qUtf8Printable(val) << std::endl;
        }
        if (!errors.empty()) {
            if (errors.length() > 5) {
                errors = errors.first(5);
                errors << "...";
            }
            emit error(errors.join("\n"));
        }
    }
    return results;
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

QString Backend::getConfig(
    QString key, QString fallback)
{
    auto kv = op1("GET_CONFIG", {key}, key);
    if (kv.key.isEmpty())
        return fallback;
    return kv.val;
}

void Backend::setConfig(QString key, QString val)
{
    op("SET_CONFIG", {key, val});
}
