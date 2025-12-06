#include "mainwindow.h"
#include <QFileDialog>
#include <QMessageBox>
#include "backend.h"
#include "progress_delegate.h"
#include "ui_mainwindow.h"

Backend *backend;

MainWindow::MainWindow(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::MainWindow)
{
    ui->setupUi(this);
    ui->instance_table->resizeColumnsToContents();
    ui->instance_table->setItemDelegateForColumn( //
        3,
        new ProgressDelegate(ui->instance_table));
    ui->specials_table->setItemDelegateForColumn( //
        1,
        new ProgressDelegate(ui->specials_table));

    auto it_progress0 = new QTableWidgetItem();
    //it_progress0->setData(UserRoleInt, 30);
    ui->specials_table->setItem(0, 1, it_progress0);
    auto it_progress1 = new QTableWidgetItem();
    //it_progress1->setData(UserRoleInt, 50);
    ui->specials_table->setItem(1, 1, it_progress1);
    auto it_progress2 = new QTableWidgetItem();
    //it_progress2->setData(UserRoleInt, 80);
    ui->specials_table->setItem(2, 1, it_progress2);

    backend = new Backend();
    connect( //
        backend,
        &Backend::logcolour,
        ui->logview,
        &QTextEdit::setTextColor);
    connect( //
        backend,
        &Backend::log,
        ui->logview,
        &QTextEdit::append);
    connect( //
        backend,
        &Backend::error,
        this,
        &MainWindow::error_popup);
    connect( //
        ui->pb_open_new,
        &QPushButton::clicked,
        this,
        &MainWindow::open_file);
    connect( //
        ui->pb_go,
        &QPushButton::clicked,
        this,
        &MainWindow::push_go);
    connect( //
        ui->pb_stop,
        &QPushButton::clicked,
        &threadrunner,
        &RunThreadController::stopThread);
    connect( //
        &threadrunner,
        &RunThreadController::elapsedTime,
        ui->elapsed_time,
        &QLineEdit::setText);
    connect( //
        &threadrunner,
        &RunThreadController::nconstraints,
        this,
        &MainWindow::nconstraints);
    connect(&threadrunner,
            &RunThreadController::handleRunFinished,
            this,
            &MainWindow::threadRunFinished);

    backend->op("CONFIG_INIT");

    /*
    settings = new QSettings("gradgrind", "fetrunner");
    //const auto geometry = settings->value("MainWindow", QByteArray()).toByteArray();
    const auto geometry = settings->value("MainWindowSize").value<QSize>();
    if (!geometry.isEmpty())
        //restoreGeometry(geometry);
        resize(geometry);
    */

    auto s = backend->getConfig("gui/MainWindowSize");
    if (!s.isEmpty()) {
        auto wh = s.split("x");
        resize(wh[0].toInt(), wh[1].toInt());
    }
}

MainWindow::~MainWindow()
{
    //settings->setValue("MainWindow", saveGeometry());
    //settings->setValue("MainWindowSize", size());
    //delete settings;
    auto s = QString("%1x%2").arg( //
        QString::number(width()),
        QString::number(height()));
    backend->setConfig("gui/MainWindowSize", s);
    delete backend;
    delete ui;
}

void MainWindow::closeEvent(
    QCloseEvent *e)
{
    qDebug() << "Quitting ...";

    QWidget::closeEvent(e);
}

void MainWindow::open_file()
{
    //qDebug() << "Open File";

    QString fdir = filedir;
    if (fdir.isEmpty()) {
        fdir = backend->getConfig("gui/SourceDir");
        if (fdir.isEmpty()) {
            fdir = QDir::homePath();
        }
    }
    QString filepath = QFileDialog::getOpenFileName( //
        this,
        tr("Open Timetable Specifiation"),
        fdir,
        tr("FET / W365 Files (*.fet *_w365.json)"));

    if (!filepath.isEmpty()) {
        for (const auto &kv : backend->op("SET_FILE", {filepath})) {
            if (kv.key == "SET_FILE") {
                QDir dir{kv.val};
                filename = dir.dirName();
                dir.cdUp();
                fdir = dir.absolutePath();
                ui->currentDir->setText(fdir);
                ui->currentFile->setText(filename);
                if (fdir != filedir) {
                    filedir = fdir;
                    backend->setConfig("gui/SourceDir", fdir);
                }
            } else if (kv.key == "DATA_TYPE") {
                datatype = kv.val;
            }
        }
    }
}

void MainWindow::error_popup(QString msg)
{
    QMessageBox::critical(this, "", msg);
}

void MainWindow::push_go()
{
    //qDebug() << "Run fetrunner";

    if (backend->op1("RUN_TT_SOURCE", {}, "OK").val == "true") {
        threadRunActivated(true);
        ui->elapsed_time->setText("0");
        threadrunner.runTtThread();
    }
}

void MainWindow::threadRunFinished()
{
    qDebug() << "threadRunFinished";
    threadRunActivated(false);
}

void MainWindow::threadRunActivated(
    bool active)
{
    ui->pb_go->setDisabled(active);
    ui->pb_stop->setEnabled(active);
    ui->pb_open_new->setDisabled(active);
    ui->frame_parameters->setDisabled(active);
}

ProgressDelegate::ProgressDelegate(
    QObject *parent)
    : QStyledItemDelegate(parent)
{}

void ProgressDelegate::paint( //
    QPainter *painter,
    const QStyleOptionViewItem &option,
    const QModelIndex &index) const
{
    auto progress = index.data(UserRoleInt).toInt();
    QStyleOptionProgressBar progbar;
    progbar.rect = option.rect;
    progbar.minimum = 0;
    progbar.maximum = 100;
    progbar.progress = progress;
    progbar.text = QString::number(progress).append('%');
    progbar.textVisible = true;
    QApplication::style()->drawControl( //
        QStyle::CE_ProgressBar,
        &progbar,
        painter);
}

void MainWindow::nconstraints(QString data)
{
    auto slist = data.split(u'.');
    ui->c_enabled_t->setText(slist[0]);
    ui->c_enabled_h->setText(slist[1] + " / " + slist[2]);
    ui->c_enabled_s->setText(slist[3] + " / " + slist[4]);
}
