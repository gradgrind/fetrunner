#include "mainwindow.h"
#include <QDir>
#include <QFileDialog>
#include <QMessageBox>
#include "backend.h"
#include "globals.h"
#include "ui_mainwindow.h"

QSettings *settings;
QString file_dir;
QString file_name;
QString file_datatype;
Notifier *notifier;

MainWindow::MainWindow(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::MainWindow)
{
    ui->setupUi(this);
    notifier = new Notifier();

    ui->help_view->setSource(QUrl("qrc:/help/using_fetrunner.md"));

    backend->registerResultHandler("FETRUNNER_VERSION",
        [this](QString arg) {do_FETRUNNER_VERSION(arg);});
    backend->registerResultHandler("SET_FILE",
        [this](QString arg) {do_SET_FILE(arg);});
    backend->registerResultHandler("DATA_TYPE",
        [this](QString arg) {do_DATA_TYPE(arg);});

    log_view = ui->base_log_view;
    connect( //
        backend,
        &Backend::logcolour,
        this,
        &MainWindow::setLogColour);
    connect( //
        backend,
        &Backend::log,
        this,
        &MainWindow::logLine);

    ttsolver = new FetRunner();
    ui->main_panel->addWidget(ttsolver);

    ttview = new TtView();
    ui->main_panel->addWidget(ttview);
    connect( //
        ui->rb_view_teacher,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                ttview->select_teacher_view();
            }
        }
    );

    connect( //
        ui->open_file,
        &QPushButton::clicked,
        this,
        &MainWindow::open_file);
    connect( //
        ui->help,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                ui->main_panel->setCurrentIndex(0);
                ui->side_panel_sub->setCurrentIndex(0);
            }
        });
    connect( //
        ui->general_log,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                ui->main_panel->setCurrentIndex(1);
                ui->side_panel_sub->setCurrentIndex(0);
            }
        });
    connect( //
        ui->solve_timetable,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                ui->main_panel->setCurrentIndex(2);
                ui->side_panel_sub->setCurrentIndex(0);
            }
        });
    connect( //
        ui->view_timetable,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked) {
                ui->main_panel->setCurrentIndex(3);
                ui->side_panel_sub->setCurrentIndex(1);
                ttview->enter_view();
            }
        });
    connect( //
        notifier,
        &Notifier::switch_logger,
        this,
        &MainWindow::switch_logger);
    connect( //
        notifier,
        &Notifier::setBusy,
        this,
        &MainWindow::set_busy);
    connect( //
        notifier,
        &Notifier::errorPopup,
        this,
        &MainWindow::error_popup);
    connect( //
        notifier,
        &Notifier::quit_register_wait,
        this,
        &MainWindow::quit_register_wait);
    connect( //
        notifier,
        &Notifier::finished,
        this,
        &MainWindow::handle_finished);
    connect( //
        notifier,
        &Notifier::new_tt_data,
        this,
        &MainWindow::new_tt_data);
    connect( //
        notifier,
        &Notifier::new_tt_data,
        ttview,
        &TtView::new_tt_data);
    connect( //
        notifier,
        &Notifier::no_tt_data,
        this,
        &MainWindow::no_tt_data);
    connect( //
        notifier,
        &Notifier::fileChanged,
        this,
        &MainWindow::new_file);

    connect( //
        backend,
        &Backend::error,
        this,
        &MainWindow::error_popup);

    settings = new QSettings("gradgrind", "fetrunner");
    const auto geometry = settings->value("gui/FetRunnerSize").value<QSize>();
    if (!geometry.isEmpty())
        resize(geometry);
}

MainWindow::~MainWindow()
{
    delete ui;
    settings->setValue("gui/FetRunnerSize", size());
    delete settings;
}

void MainWindow::closeEvent(QCloseEvent *e)
{
    //qDebug() << "closeEvent()" << quit_confirmed << quit_requested << waiting_on.length();
    if (quit_confirmed) {
        QWidget::closeEvent(e);
        return;
    }
    if (!quit_requested) {
        quit_requested = true;
        notifier->emit closeRequest();
        // Assume the signal used Qt::DirectConnection, so that all
        // entries in `waiting_on` are now set.
        if (waiting_on.length() == 0) {
            QWidget::closeEvent(e);
            return;
        }
    }
    e->ignore();
}

void MainWindow::logLine(QString line) {
    log_view->append(line);
}

void MainWindow::setLogColour(QColor colour) {
    log_view->setTextColor(colour);
}


void MainWindow::quit_register_wait(QString module)
{
    //qDebug() << "quit_register_wait()" << module;
    if (!waiting_on.contains(module)) {
        waiting_on.append(module);
    }
}

void MainWindow::handle_finished(QString module)
{
    //qDebug() << "handle_finished()" << quit_requested;
    if (quit_requested) {
        waiting_on.removeOne(module);
        if (waiting_on.length() == 0) {
            quit_confirmed = true;
            close();
        }
    } else {
        log_view = ui->base_log_view; // revert to base log view
    }
}

void MainWindow::error_popup(const QString msg)
{
    QMessageBox::critical(this, "", msg);
}

void MainWindow::open_file()
{
    //qDebug() << "Open File";

    QString fdir = file_dir;
    if (fdir.isEmpty()) {
        fdir = settings->value("gui/SourceDir", QDir::homePath()).toString();
    }
    QString filepath = QFileDialog::getOpenFileName( //
        this,
        tr("Open Timetable Specification File"),
        fdir,
        tr("FET / W365 Files (*.fet *_w365.json)"));

    if (!filepath.isEmpty()) {
        if (backend->op("SET_FILE", {filepath}))
            notifier->emit fileChanged();
    }
}

void MainWindow::new_file() {
    // Select fetrunner view, disable timetable view.
    no_tt_data();
    ui->solve_timetable->setEnabled(true);
    ui->solve_timetable->click();
}

void MainWindow::set_busy(bool on) {
    ui->control_panel->setDisabled(on);
}

void MainWindow::switch_logger(QString msg, QTextEdit *log_view_widget) {
    if (log_view == nullptr)
        log_view = ui->base_log_view;
    else {
        log_view->append(msg);
        log_view = log_view_widget;
    }
}

void MainWindow::do_FETRUNNER_VERSION(const QString &val) {
    ui->fetrunner_version->setText(val);
}

void MainWindow::do_SET_FILE(const QString &val) {
    QDir dir{val};
    file_name = dir.dirName();
    dir.cdUp();
    ui->file_path->setText(dir.absoluteFilePath(file_name));
    auto fdir = dir.absolutePath();
    if (fdir != file_dir) {
        file_dir = fdir;
        settings->setValue("gui/SourceDir", fdir);
    }
}

void MainWindow::do_DATA_TYPE(const QString &val) {
    file_datatype = val;
}

void MainWindow::new_tt_data() {
    if (quit_requested) return;
    ui->view_timetable->setEnabled(true);
}

void MainWindow::no_tt_data() {
    ui->view_timetable->setEnabled(false);
}
