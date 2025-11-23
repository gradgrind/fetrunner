#include "backend.h"
#include "../libfetrunner/libfetrunner.h"
#include "support.h"

QJsonArray backend(
    QString op, QStringList data)
{
    QJsonObject cmd{{"Op", op}, {"Data", QJsonArray::fromStringList(data)}};
    QJsonDocument doc(cmd);
    auto cs = doc.toJson(QJsonDocument::Compact);
    auto result = FetRunner(cs.data());
    qDebug() << "§" << cmd << "->" << result;
    auto jsondoc = QJsonDocument::fromJson(result);
    if (!jsondoc.isArray()) {
        showError(QString{"BackendReturnError: "} + result);
        return QJsonArray{};
    }
    return jsondoc.array();
}
