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
        QString citem{c.tag};
        if (c.name != c.tag)
            citem += " : " + c.name;
        auto twitem = new QTreeWidgetItem(ui->view_choice_list);
        twitem->setText(0, citem);
        twitem->setData(0, Qt::UserRole, i);
        // Add sub-items, sorted atomic groups, shown as list of groups
        auto aglist = c.atom_groups.keys();
        if (aglist.length() > 1) {
            std::sort(aglist.begin(), aglist.end());
            for (auto ag : std::as_const(aglist)) {
                QStringList glist = c.atom_groups.value(ag);
                citem = QString::number(ag) + " : " + glist.join(",");
                auto twitem2 = new QTreeWidgetItem(twitem);
                twitem2->setText(0, citem);
                twitem2->setData(0, Qt::UserRole, i);
                twitem2->setData(0, Qt::UserRole+1, ag);
                //TODO: Use the agto select the atomic group
            }
        }
        i++;
    }
    set_view = [this](int i) {
        this->ttview->set_class(i);
    };
}

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