#ifndef BACKEND_H
#define BACKEND_H

#include <QColor>
#include <QJsonArray>
#include <QJsonDocument>
#include <QJsonObject>
#include <QObject>
#include <QStringList>

struct KeyVal
{
    QString key;
    QString val;
};

// From autotimetable/placements.go, keep synchronized
const int PF_DAY = 0;
const int PF_HOUR = 1;
const int PF_LENGTH = 2;
const int PF_SUBJECT = 3;
const int PF_GROUPS = 4;
const int PF_ATOMICS = 5;
const int PF_TEACHERS = 6;
const int PF_ROOMS = 7;

struct Placement
{
    int activity;
    int day;
    int hour;
    QList<int> rooms;
};

QList<Placement *> get_placements(QString cmd, int item);

//TODO: When releasing a list of pointers, the objects themselves
// must be deleted:
//    qDeleteAll(list.begin(), list.end());
//    list.clear();

class Backend : public QObject
{
    Q_OBJECT

public:
    Backend();
    //~Backend();

    QList<KeyVal> op(QString cmd, QStringList data = {});
    KeyVal readlogline();
    KeyVal readresult(QString r);

    //TODO-- QList<KeyVal> op(QString cmd, QStringList data = {});
    KeyVal op1(QString cmd, QStringList data = {}, QString key = {});
    QString getConfig(QString key, QString fallback = {});
    void setConfig(QString key, QString val);

private:
    QString logline;

signals:
    void logcolour(QColor);
    void log(QString);
    void error(QString);
};

extern Backend *backend;

#endif // BACKEND_H
