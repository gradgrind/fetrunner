#ifndef TTVIEW_H
#define TTVIEW_H

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

    void set_teacher(int tix);

private:
    Ui::TtView *ui;
    //Canvas *canvas;
    TtGrid *grid{nullptr};
    TtBase *ttbase{nullptr};

    void do_TEACHER_PLACEMENTS(const QString &val);
    void do_ROOM_PLACEMENTS(const QString &val);
    void do_CLASS_PLACEMENTS(const QString &val);

public slots:
    void new_tt_data();

    void select_teacher_view();
};

#endif // TTVIEW_H
