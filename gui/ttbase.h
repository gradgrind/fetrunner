#ifndef TTBASE_H
#define TTBASE_H

#include <QStringList>

/*
struct TtPlacement
{
    int activity;
    int day;
    int hour;
    QList<int> rooms;
};

class TtPlacementList : public QList<TtPlacement *>
{
public:
    TtPlacementList(QString cmd, int item);
    ~TtPlacementList()
    {
        qDeleteAll(begin(), end());
        clear();
    }
};
*/

struct TtActivity
{
    int length;
    int day;
    int hour;
    QString subject;
    QList<int> teachers;
    QList<int> atomics;
    QStringList groups;
    QList<int> rooms;
};

struct TtName
{
    QString tag;
    QString name;
};

struct TtClass : TtName
{
    QList<int> atomics;
    QStringList groups;
};

struct TileData
{
    int length;
    QString subject;
    QStringList teachers;
    QStringList rooms;
    QList<int> atomics;
    QStringList groups;
};

class TtBase
{
    //TODO: I might want to have the class-group separator
    // (currently only ".") available here.
private:
    void clear_activities()
    {
        qDeleteAll(activities.begin(), activities.end());
        activities.clear();
    }
    void set_activity(const QString &val);
    void set_class(const QString &val);
    void set_teacher(const QString &val);
    void set_room(const QString &val);
    void set_day(const QString &val);
    void set_hour(const QString &val);

public:
    TtBase();
    ~TtBase() { clear_activities(); }
    const QList<TtActivity *> get_activities();
    const TtClass &get_class(int cix);
    const QList<TtName> get_days() { return days; }
    const QList<TtName> get_hours() { return hours; }

    int place_activity(const QString &val);

    //TileData *get_tile_data(TtPlacement *p);

    QList<TtActivity *> activities;
    QList<TtName> days;
    QList<TtName> hours;
    QList<TtClass> classes;
    QList<TtName> teachers;
    QList<TtName> rooms;
};

//extern TtBase tt_base;

#endif // TTBASE_H
