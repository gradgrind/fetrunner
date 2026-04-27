#include "ttview.h"
#include "ui_ttview.h"
#include "tt_show_resource.h"

TtView::TtView(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::TtView)
{
    ui->setupUi(this);
    //canvas = new Canvas(ui->canvas_view);
    grid = new TtGrid(ui->canvas_view,
                      {"Mo", "Tu", "We", "Th", "Fr"},
                      {"A", "B", "1", "2", "3", "4", "5", "6", "7"},
                      {2, 4, 6});
}

TtView::~TtView()
{
    delete ui;
}

void TtView::set_teacher(int tix)
{
    if (ttbase == nullptr) return;
    delete grid;
    auto days = ttbase->get_days();
    auto hours = ttbase->get_hours();
    //auto breaks = tt_base->get_breaks();
    QList<int> breaks;
    QStringList dlist;
    for (const auto &d : days) {
        dlist.append(d.tag);
    }
    QStringList hlist;
    for (const auto &h : hours) {
        dlist.append(h.tag);
    }
    grid = new TtGrid(ui->canvas_view, dlist, hlist, breaks);
    ShowTeacher(grid, ttbase, tix);
}

//TODO: There will need to be a list of teachers to select from.
void TtView::select_teacher_view()
{
    qDebug() << "Hi, testing teacher view!";
    set_teacher(15);
}
