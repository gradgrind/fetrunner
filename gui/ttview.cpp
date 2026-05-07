#include "ttview.h"
#include "backend.h"
#include "globals.h"
#include "ui_ttview.h"

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

    backend->registerResultHandler("TEACHER_PLACEMENT",
        [this](QString arg) {do_TEACHER_PLACEMENT(arg);});
    backend->registerResultHandler("ROOM_PLACEMENT",
        [this](QString arg) {do_ROOM_PLACEMENT(arg);});
    backend->registerResultHandler("CLASS_PLACEMENT",
        [this](QString arg) {do_CLASS_PLACEMENT(arg);});

    //canvas = new Canvas(ui->canvas_view);
    // Generate an example grid:
    grid = new TtGrid(ui->canvas_view,
                      {"Mo", "Tu", "We", "Th", "Fr"},
                      {"A", "B", "1", "2", "3", "4", "5", "6", "7"},
                      {2, 4, 6});
}

TtView::~TtView() {
    delete ui;
}

void TtView::new_tt_data() {
    delete ttbase;
    ttbase = nullptr;
}

//TODO
void TtView::enter_view() {
    if (ttbase == nullptr)
        emit notifier->switch_logger(">>> --TIMETABLE", 2);
        ttbase = new TtBase();
        emit notifier->switch_logger("", 0);
        new_grid(); // show a new empty grid
}

void TtView::new_grid() {
    if (ttbase == nullptr) return;
    delete grid;
    auto days = ttbase->get_days();
    auto hours = ttbase->get_hours();
    //TODO: auto breaks = tt_base->get_breaks();
    QList<int> breaks;
    QStringList dlist;
    for (const auto &d : days) {
        dlist.append(d.tag);
    }
    QStringList hlist;
    for (const auto &h : hours) {
        hlist.append(h.tag);
    }
    grid = new TtGrid(ui->canvas_view, dlist, hlist, breaks);

    //TODO: select class/room/teacher list and then resource
}
