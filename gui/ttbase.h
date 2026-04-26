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
    QList<TtClass> classes;
    QList<TtName> teachers;
    QList<TtName> rooms;

public:
    TtBase();
    ~TtBase() { clear_activities(); }
    void set_activities();
    const QList<TtActivity *> get_activities();
    void set_classes();
    const TtClass &get_class(int cix);
    void set_teachers();
    void set_rooms();

    TileData *get_tile_data(TtPlacement *p);
};

//extern TtBase tt_base;

#endif // TTBASE_H
