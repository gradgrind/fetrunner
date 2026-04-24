#include "ttview.h"
#include "canvas.h"
#include "ui_ttview.h"

TtView::TtView(
    QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::TtView)
{
    ui->setupUi(this);
    auto canvas = new Canvas(ui->canvas_view);
}

TtView::~TtView()
{
    delete ui;
}
