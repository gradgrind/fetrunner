#ifndef TTVIEW_H
#define TTVIEW_H

#include <QWidget>

namespace Ui { class TtView; }

class TtView : public QWidget
{
    Q_OBJECT

public:
    explicit TtView(QWidget *parent = nullptr);
    ~TtView();

private:
    Ui::TtView *ui;
};

#endif // TTVIEW_H
