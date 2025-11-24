#ifndef BACKEND_H
#define BACKEND_H

#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QObject>
#include <QTextEdit>

struct KeyVal
{
    QString key;
    QString val;
};

class Backend : public QObject
{
    Q_OBJECT

public:
    Backend();
    //~Backend();

    QList<KeyVal> op(QString cmd, QStringList data = {});
    KeyVal op1(QString cmd, QStringList data = {}, QString key = {});
    QString getConfig(QString key, QString fallback = {});
    void setConfig(QString key, QString val);

signals:
    void log(QString);
    void error(QString);
};

extern Backend *backend;

#endif // BACKEND_H
