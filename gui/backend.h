#ifndef BACKEND_H
#define BACKEND_H

#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QObject>

QString backend(QString op, QStringList data = {});

#endif // BACKEND_H
