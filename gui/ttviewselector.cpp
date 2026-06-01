#include "ttviewselector.h"
#include "ui_ttviewselector.h"

TtViewSelector::TtViewSelector(TtView *ttview_in, QWidget *parent) :
    QWidget(parent),
    ui(new Ui::TtViewSelector),
    ttview(ttview_in)
{
    ui->setupUi(this);
    buttonGroup.addButton(ui->rb_view_class);
    buttonGroup.addButton(ui->rb_view_teacher);
    buttonGroup.addButton(ui->rb_view_room);

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
        ui->rb_view_room,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                this->select_room_view();
            }
        }
    );
    connect(
        ui->rb_view_class,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                this->select_class_view();
            }
        }
    );

    connect(
        ui->view_choice_list,
        &QTreeWidget::currentItemChanged,
        this,
        &TtViewSelector::chosen
    );
}

TtViewSelector::~TtViewSelector()
{
    delete ui;
}

void TtViewSelector::do_new_tt_data() {
    ui->view_choice_list->clear();
    // Uncheck all radio buttons
    auto b = buttonGroup.checkedButton();
    if(b != nullptr) {
       // Disable the exclusive property of the Button Group
       buttonGroup.setExclusive(false);
       // Get the checked button and uncheck it
       b->setChecked(false);
       // Enable the exclusive property of the Button Group
       buttonGroup.setExclusive(true);
    }
    // Deal with TtView
    ttview->do_new_tt_data();
}

void TtViewSelector::select_teacher_view()
{
    ui->view_choice_list->clear();
    int i = 0;
    for (const auto &t : std::as_const(ttview->ttbase->teachers)) {
        QStringList choice{t.tag, t.name};
        if (t.name == t.tag)
            choice[1] = "";
        auto twitem = new QTreeWidgetItem(ui->view_choice_list, choice);
        twitem->setData(0, Qt::UserRole, i++);
    }
    set_view = [this](int i) {
        this->ttview->set_teacher(i);
    };
}

void TtViewSelector::select_room_view()
{
    ui->view_choice_list->clear();
    int i = 0;
    for (const auto &r : std::as_const(ttview->ttbase->rooms)) {
        QStringList choice{r.tag, r.name};
        if (r.name == r.tag)
            choice[1] = "";
        auto twitem = new QTreeWidgetItem(ui->view_choice_list, choice);
        twitem->setData(0, Qt::UserRole, i++);
    }
    set_view = [this](int i) {
        this->ttview->set_room(i);
    };
}

void TtViewSelector::select_class_view()
{
    ui->view_choice_list->clear();
    int i = 0;
    for (const auto &c : std::as_const(ttview->ttbase->classes)) {
        QStringList choice{c.tag, c.name};
        if (c.name == c.tag)
            choice[1] = "";
        auto twitem = new QTreeWidgetItem(ui->view_choice_list, choice);
        twitem->setData(0, Qt::UserRole, i++);
    }
    set_view = [this](int i) {
        this->ttview->set_class(i);
    };
}

//TODO: set_view will probably need adapting to cope with atomic groups ...

void TtViewSelector::chosen(QTreeWidgetItem *current, QTreeWidgetItem *previous) {
    if (current == nullptr) {
        qDebug() << "No QTreeWidgetItem";
        //TODO?
        ttview->new_grid();
    } else {
        auto i = current->data(0, Qt::UserRole).toInt();
        qDebug() << "chosen" << i << ttview->ttbase->teachers.at(i).tag;
        set_view(i); // for class, room or teacher
    }
}