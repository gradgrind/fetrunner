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

public slots:
    void select_teacher_view();
};

#endif // TTVIEW_H
