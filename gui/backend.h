#ifndef BACKEND_H
#define BACKEND_H

#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QObject>

QJsonArray backend(QString op, QStringList data = {});

#endif // BACKEND_H
