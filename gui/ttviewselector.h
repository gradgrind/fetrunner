#ifndef TTVIEWSELECTOR_H
#define TTVIEWSELECTOR_H

#include <QWidget>
#include "ttview.h"

namespace Ui {
class TtViewSelector;
}

class TtViewSelector : public QWidget
{
    Q_OBJECT

public:
    explicit TtViewSelector(TtView *ttview, QWidget *parent = nullptr);
    ~TtViewSelector();

private:
    Ui::TtViewSelector *ui;
    TtView *ttview; // convenience copy, not owned here

private slots:
    void select_teacher_view();
    void chosen(int i);
};

#endif // TTVIEWSELECTOR_H
