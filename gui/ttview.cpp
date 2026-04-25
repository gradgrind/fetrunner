#include "ttview.h"
#include "ui_ttview.h"

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
