#include "mainwindow.h"
#include <QDir>
#include <QFileDialog>
#include "backend.h"
#include "fetrunner.h"
#include "globals.h"
#include "ttview.h"
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
    backend = new Backend();
    notifier = new Notifier();

    auto ttsolver = new FetRunner();
    ui->main_panel->addWidget(ttsolver);

    auto ttview = new TtView();
    ui->main_panel->addWidget(ttview);

    connect( //
        ui->open_file,
        &QPushButton::clicked,
        this,
        &MainWindow::open_file);
    connect( //
        ui->solve_timetable,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked)
                ui->main_panel->setCurrentIndex(0);
        });
    connect( //
        ui->view_timetable,
        &QRadioButton::toggled,
        this,
        [this](bool checked) {
            if (checked)
                ui->main_panel->setCurrentIndex(1);
        });
    connect( //
        notifier,
        &Notifier::setBusy,
        this,
        &MainWindow::set_busy);

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
    delete backend;
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
        notifier->emit fileChanged();
        for (const auto &kv : backend->op("SET_FILE", {filepath})) {
            if (kv.key == "SET_FILE") {
                QDir dir{kv.val};
                file_name = dir.dirName();
                dir.cdUp();
                fdir = dir.absolutePath();
                ui->file_dir->setText(fdir);
                ui->file_name->setText(file_name);
                if (fdir != file_dir) {
                    file_dir = fdir;
                    settings->setValue("gui/SourceDir", fdir);
                }
            } else if (kv.key == "DATA_TYPE") {
                file_datatype = kv.val;
            }
        }
    }
}

void MainWindow::set_busy(bool on)
{
    ui->control_panel->setDisabled(on);
}
