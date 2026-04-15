#include "ttview.h"
#include "ui_ttview.h"

TtView::TtView(
    QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::TtView)
{
    ui->setupUi(this);
}

TtView::~TtView()
{
    delete ui;
}
