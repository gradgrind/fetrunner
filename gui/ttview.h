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

    void set_teacher(TtBase *ttbase, int tix);

private:
    Ui::TtView *ui;
    //Canvas *canvas;
    TtGrid *grid;

public slots:
    void select_teacher_view();
};

#endif // TTVIEW_H
