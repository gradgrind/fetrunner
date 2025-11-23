#include "backend.h"
#include "../libfetrunner/libfetrunner.h"
#include "support.h"

struct KeyVal
{
    QString key;
    QString val;
};

QString jresult(QJsonArray jarr)
{
    QList<KeyVal> results;
    //QStringList messages;
    for (auto &&c : jarr) {
        auto e = c.toObject();
        auto key = e["Type"].toString();
        auto val = e["Text"].toString();
        if (key == "$") {
            // a result
            auto n = val.indexOf('=');
            if (n < 0) {
                //TODO: error
                qDebug() << "BUG:" << key << val;
            } else {
                key = val.first(n);
                val = val.sliced(n + 1);
                results.append({key, val});
                qDebug() << "$$$" << key << "=" << val;
            }
        } else if (key == "+++") {
        } else {
            //messages.append(key + " " + val);
            qDebug() << key << val;
        }
    }
    return "";
}

QString backend(QString op, QStringList data)
{
    auto darray = QJsonArray::fromStringList(data);
    qDebug() << QString{"+++ "} + op << darray;
    QJsonObject cmd{{"Op", op}, {"Data", darray}};
    QJsonDocument doc(cmd);
    auto cs = doc.toJson(QJsonDocument::Compact);
    auto result = FetRunner(cs.data());
    //qDebug() << "§" << cmd << "->" << result;
    auto jsondoc = QJsonDocument::fromJson(result);
    if (!jsondoc.isArray()) {
        showError(QString{"BackendReturnError: "} + result);
    } else
        jresult(jsondoc.array());
    return QString{"--- "} + op;
}
