#ifndef TTLAYOUT_H
#define TTLAYOUT_H

#include <QStringList>

class TtLayout
{
public:
    TtLayout();

    void set_headers( //
        QStringList days,
        QStringList timeslots,
        bool days_vertical);
};

#endif // TTLAYOUT_H
