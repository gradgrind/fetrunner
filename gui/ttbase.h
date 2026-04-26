#ifndef TTBASE_H
#define TTBASE_H

#include <QList>
#include <QString>

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
    TtPlacementList();
    ~TtPlacementList()
    {
        qDeleteAll(begin(), end());
        clear();
    }
};

const TtPlacementList get_item_placements(QString cmd, int item);

struct TtActivity
{
    int length;
    QString subject;
    QList<int> teachers;
    QList<int> atomics;
    QStringList groups;
};

class TtBase
{
private:
    QList<TtActivity *> tt_activities;
    void clear_activities()
    {
        qDeleteAll(tt_activities.begin(), tt_activities.end());
        tt_activities.clear();
    }

public:
    TtBase();
    ~TtBase() { clear_activities(); }
    void set_tt_activities();
    const QList<TtActivity *> get_tt_activities();
};

//extern TtBase tt_base;

#endif // TTBASE_H
