#ifndef TTVIEW_H
#define TTVIEW_H

#include <QList>
#include <QWidget>
#include "ttbase.h"
#include "ttgrid.h"

namespace Ui {

class TtView;

}

class TtView : public QWidget
{
    Q_OBJECT

public:
    explicit TtView(QWidget *parent = nullptr);
    ~TtView();

    void set_class(int cix);
    void set_room(int rix);
    void set_teacher(int tix);

private:
    Ui::TtView *ui;
    //Canvas *canvas;
    TtGrid *grid{nullptr};
    TtBase *ttbase{nullptr};

    // An array (days * hours) of activity index lists is used
    // for arranging the Tiles in a time slot for class views.
    QList<QList<QList<int>>> weekBuffer;
    QList<int> classAtomics; // list of atomics for viewed class

    void new_grid();
    void do_TEACHER_PLACEMENT(const QString &val);
    void do_ROOM_PLACEMENT(const QString &val);
    void do_CLASS_PLACEMENT(const QString &val);
    void do_SetupClassView(const QString &val);

public slots:
    void new_tt_data();

    void select_teacher_view();
};

#endif // TTVIEW_H
