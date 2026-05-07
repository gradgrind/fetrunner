#include "ttviewselector.h"
#include "ui_ttviewselector.h"

TtViewSelector::TtViewSelector(TtView *ttview_in, QWidget *parent) :
    QWidget(parent),
    ui(new Ui::TtViewSelector),
    ttview(ttview_in)
{
    ui->setupUi(this);

    connect(
        ui->rb_view_teacher,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                this->select_teacher_view();
            }
        }
    );

    connect(
        ui->view_choice_list,
        &QListWidget::currentRowChanged,
        this,
        &TtViewSelector::chosen
    );
}

TtViewSelector::~TtViewSelector()
{
    delete ui;
}

void TtViewSelector::select_teacher_view()
{
    ui->view_choice_list->clear();
    for (const auto &t : std::as_const(ttview->ttbase->teachers)) {
        QString choice{t.tag};
        if (!t.name.isEmpty() && t.name != t.tag)
            choice.append(" " + t.name);
        ui->view_choice_list->addItem(choice);
    }
}

void TtViewSelector::chosen(int i) {
    if (i < 0) {
        //TODO?
        ttview->new_grid();
    } else {
        //qDebug() << "chosen" << i << ttview->ttbase->teachers.at(i).tag;

        //TODO: teacher, room or class?
        ttview->set_teacher(i);
    }
}