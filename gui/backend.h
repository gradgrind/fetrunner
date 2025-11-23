#ifndef BACKEND_H
#define BACKEND_H

#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QObject>

struct KeyVal
{
    QString key;
    QString val;
};

QList<KeyVal> backend(QString op, QStringList data = {});
QString getConfig(QString key);
void setConfig(QString key, QString val);

#endif // BACKEND_H
