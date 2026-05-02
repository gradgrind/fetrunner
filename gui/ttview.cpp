#include "ttview.h"
#include "backend.h"
#include "ui_ttview.h"
#include "tt_show_resource.h"

//TODO: It might be useful to be able to reload old timetables
// for viewing. The current approach can't really support that,
// as the placements are not stored with the input data.
// Perhaps by loading a result file this could be done? Of course,
// without tight coupling between the two files, they could get
// "out of sync". Perhaps saving the placements within the source
// file would be a possibility?

TtView::TtView(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::TtView)
{
    ui->setupUi(this);

    backend.registerResultHandler("TT_TEACHER_PLACEMENTS",
        [this](QString arg) {do_TEACHER_PLACEMENTS(arg);});
    backend.registerResultHandler("TT_ROOM_PLACEMENTS",
        [this](QString arg) {do_ROOM_PLACEMENTS(arg);});
    backend.registerResultHandler("TT_CLASS_PLACEMENTS",
        [this](QString arg) {do_CLASS_PLACEMENTS(arg);});

    //canvas = new Canvas(ui->canvas_view);
    // Generate an example grid:
    grid = new TtGrid(ui->canvas_view,
                      {"Mo", "Tu", "We", "Th", "Fr"},
                      {"A", "B", "1", "2", "3", "4", "5", "6", "7"},
                      {2, 4, 6});
}

TtView::~TtView()
{
    delete ui;
}

void TtView::new_tt_data()
{
    qDebug() << "new_tt_data";
    //TODO: Skip this if a quit has been requested.
    //return;
    delete ttbase;
    ttbase = new TtBase();
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
