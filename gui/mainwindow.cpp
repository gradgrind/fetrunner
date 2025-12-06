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

    auto item = new QTableWidgetItem();
    item->setTextAlignment(Qt::AlignCenter);
    ui->specials_table->setItem(0, 0, item);
    ui->specials_table->setItem(0, 1, new QTableWidgetItem());

    item = new QTableWidgetItem();
    item->setTextAlignment(Qt::AlignCenter);
    ui->specials_table->setItem(1, 0, item);
    ui->specials_table->setItem(1, 1, new QTableWidgetItem());

    item = new QTableWidgetItem();
    item->setTextAlignment(Qt::AlignCenter);
    ui->specials_table->setItem(2, 0, item);
    ui->specials_table->setItem(2, 1, new QTableWidgetItem());

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
        &RunThreadController::ticker,
        this,
        &MainWindow::ticker);
    connect( //
        &threadrunner,
        &RunThreadController::nconstraints,
        this,
        &MainWindow::nconstraints);
    connect( //
        &threadrunner,
        &RunThreadController::progress,
        this,
        &MainWindow::progress);
    connect(&threadrunner,
            &RunThreadController::runThreadWorkerDone,
            this,
            &MainWindow::runThreadWorkerDone);

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

void MainWindow::error_popup(const QString &msg)
{
    QMessageBox::critical(this, "", msg);
}

void MainWindow::push_go()
{
    //qDebug() << "Run fetrunner";

    if (backend->op1("RUN_TT_SOURCE", {}, "OK").val == "true") {
        threadRunActivated(true);
        ui->elapsed_time->setText("0");
        for (int i = 0; i < 3; ++i) {
            ui->specials_table->item(i, 0)->setText("");
            ui->specials_table->item(i, 1)->setData(UserRoleInt, 0);
        }
        threadrunner.runTtThread();
    }
}

void MainWindow::runThreadWorkerDone()
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

void MainWindow::ticker(const QString &data)
{
    ui->elapsed_time->setText(data);
    timeTicks = data;
}

void MainWindow::nconstraints(const QString &data)
{
    auto slist = data.split(u'.');
    ui->c_enabled_t->setText(timeTicks);
    ui->c_enabled_h->setText(slist[0] + " / " + slist[1]);
    ui->c_enabled_s->setText(slist[2] + " / " + slist[3]);
}

void MainWindow::progress(const QString &data)
{
    auto slist = data.split(u'.');
    auto key = slist[0];
    if (key == "_COMPLETE") {
        ui->specials_table->item(0, 0)->setText(timeTicks);
        ui->specials_table->item(0, 1)->setData(UserRoleInt, slist[1].toInt());
    } else if (key == "_HARD_ONLY") {
        ui->specials_table->item(1, 0)->setText(timeTicks);
        ui->specials_table->item(1, 1)->setData(UserRoleInt, slist[1].toInt());
    } else if (key == "_UNCONSTRAINED") {
        ui->specials_table->item(2, 0)->setText(timeTicks);
        ui->specials_table->item(2, 1)->setData(UserRoleInt, slist[1].toInt());
    }
}
