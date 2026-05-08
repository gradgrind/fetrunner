#ifndef TTVIEWSELECTOR_H
#define TTVIEWSELECTOR_H

#include <QWidget>
#include <QButtonGroup>
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
    std::function<void(int)> set_view;
    QButtonGroup buttonGroup;

public slots:
    void do_new_tt_data();

private slots:
    void select_teacher_view();
    void select_room_view();
    void select_class_view();
    void chosen(int i);
};

#endif // TTVIEWSELECTOR_H
