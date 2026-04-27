#ifndef TTBASE_H
#define TTBASE_H

#include <QStringList>

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

struct TtActivity
{
    int length;
    QString subject;
    QList<int> teachers;
    QList<int> atomics;
    QStringList groups;
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
private:
    QList<TtActivity *> activities;
    void clear_activities()
    {
        qDeleteAll(activities.begin(), activities.end());
        activities.clear();
    }
    QList<TtName> days;
    QList<TtName> hours;
    QList<TtClass> classes;
    QList<TtName> teachers;
    QList<TtName> rooms;
    void set_activities();
    void set_classes();
    void set_teachers();
    void set_rooms();
    void set_days();
    void set_hours();

public:
    TtBase();
    ~TtBase() { clear_activities(); }
    const QList<TtActivity *> get_activities();
    const TtClass &get_class(int cix);
    const QList<TtName> get_days() { return days; }
    const QList<TtName> get_hours() { return hours; }

    TileData *get_tile_data(TtPlacement *p);
};

//extern TtBase tt_base;

#endif // TTBASE_H
