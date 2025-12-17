#include "mainwindow.h"
#include <QFileDialog>
#include <QMessageBox>
#include <QTimer>
#include "backend.h"
#include "progress_delegate.h"
#include "ui_mainwindow.h"

Backend *backend;

MainWindow::MainWindow(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::MainWindow)
{
    ui->setupUi(this);
    backend = new Backend();
    init_ttgen_tables();

    // Get range for number of processes.
    // Do this before connecting the "valueChanged" signal, to
    // avoid triggering this before any actual change.
    auto nps = backend->op1("N_PROCESSES", {}, "N_PROCESSES").val;
    auto nn = nps.split(".");
    auto n0 = nn[0].toInt();
    auto n1 = nn[1].toInt();
    if (n1 < n0)
        n1 = n0;
    auto n = nn[2].toInt();
    ui->tt_processes->setMinimum(n0);
    ui->tt_processes->setMaximum(n1);
    ui->tt_processes->setValue(n);

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
        ui->tt_processes,
        &QSpinBox::valueChanged,
        this,
        &MainWindow::nprocesses);
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
        this,
        &MainWindow::push_stop);
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
        &MainWindow::iprogress);
    connect( //
        &threadrunner,
        &RunThreadController::istart,
        this,
        &MainWindow::istart);
    connect( //
        &threadrunner,
        &RunThreadController::iend,
        this,
        &MainWindow::iend);
    connect( //
        &threadrunner,
        &RunThreadController::iaccept,
        this,
        &MainWindow::iaccept);
    connect(&threadrunner,
            &RunThreadController::runThreadWorkerDone,
            this,
            &MainWindow::runThreadWorkerDone);

    QValidator *validator1 = new QIntValidator(0, 99999, this);
    ui->tt_timeout->setValidator(validator1);

    settings = new QSettings("gradgrind", "fetrunner");
    const auto geometry = settings->value("gui/MainWindowSize").value<QSize>();
    if (!geometry.isEmpty())
        resize(geometry);

    QTimer::singleShot(0, this, &MainWindow::init2);
}

void MainWindow::init2()
{
    // This is run immediately after starting the event loop.
    ui->progress_table->resizeColumnsToContents();
    ui->instance_table->resizeColumnsToContents();
    reset_display();

    // Check FET
    auto fetpath0 = settings->value("fet/FetPath").toString();
    auto fetpath = fetpath0;
    while (true) {
        if (!fetpath.isEmpty()) {
            backend->op("TT_PARAMETER", {"FETPATH", fetpath});
        }
        auto fetv = backend->op1("GET_FET", {}, "FET_VERSION").val;
        if (!fetv.isEmpty()) {
            if (fetpath != fetpath0) {
                settings->setValue("fet/FetPath", fetpath);
            }
            break;
        }
        // Handle FET executable not found
        fetpath = QFileDialog::getOpenFileName( //
            this,
            tr("Seek FET executable"),
            QDir::homePath(),
            tr("FET executable") + " (" + FET_CL + ")");
        if (fetpath.isEmpty()) {
            QApplication::exit(1);
            break;
        }
    }
}

MainWindow::~MainWindow()
{
    settings->setValue("gui/MainWindowSize", size());
    delete settings;
    delete backend;
    delete ui;
}

void MainWindow::closeEvent(
    QCloseEvent *e)
{
    quit_requested = true;
    if (thread_running) {
        push_stop();
        e->ignore();
    } else
        QWidget::closeEvent(e);
}

void MainWindow::nprocesses(
    int n)
{
    auto nn = QString::number(n);
    auto mp = backend->op1("TT_PARAMETER", {"MAXPROCESSES", nn}, "MAXPROCESSES");
    if (mp.val != nn)
        error_popup("BUG: invalid number of processes: " + nn);
    ui->tt_processes->setValue(mp.val.toInt());
}

void MainWindow::reset_display()
{
    ui->logview->clear();

    ui->progress_table->setRowCount(0);
    ui->hard_naccepted->clear();
    ui->hard_nconstraints->clear();
    ui->hard_tlastchange->clear();
    ui->soft_naccepted->clear();
    ui->soft_nconstraints->clear();
    ui->soft_tlastchange->clear();

    ui->instance_table->setRowCount(0);
    ui->elapsed_time->setText("0");
    ui->progress_complete->clear();
    ui->progress_hard->setValue(0);
    ui->progress_hard_only->clear();
    ui->progress_soft->setValue(0);
    ui->progress_soft->setEnabled(0);
    ui->progress_unconstrained->clear();
    hard_count.clear();
    soft_count.clear();
}

void MainWindow::open_file()
{
    //qDebug() << "Open File";

    QString fdir = filedir;
    if (fdir.isEmpty()) {
        fdir = settings->value("gui/SourceDir", QDir::homePath()).toString();
    }
    QString filepath = QFileDialog::getOpenFileName( //
        this,
        tr("Open Timetable Specification File"),
        fdir,
        tr("FET / W365 Files (*.fet *_w365.json)"));

    if (!filepath.isEmpty()) {
        reset_display();
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
                    settings->setValue("gui/SourceDir", fdir);
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
    // Clear log
    ui->logview->clear();

    // Set parameters
    auto t = ui->tt_timeout->text();
    backend->op("TT_PARAMETER", {"TIMEOUT", t});
    auto sh = ui->tt_skip_hard->isChecked();
    backend->op("TT_PARAMETER", {"SKIP_HARD", sh ? "true" : "false"});

    instance_row_map.clear();
    reset_display();
    if (backend->op1("RUN_TT_SOURCE", {}, "OK").val == "true") {
        setup_progress_table();
        threadRunActivated(true);
        threadrunner.runTtThread();
    }
}

void MainWindow::push_stop()
{
    ui->pb_stop->setEnabled(false);
    threadrunner.stopThread();
    closingMessageBox.setText(tr("Finishing ..."));
    closingMessageBox.setIcon(QMessageBox::Information);
    closingMessageBox.setStandardButtons(QMessageBox::NoButton);
    closingMessageBox.exec();
}

void MainWindow::runThreadWorkerDone()
{
    //qDebug() << "threadRunFinished";
    threadRunActivated(false);
    closingMessageBox.hide();
    if (quit_requested)
        close();
}

void MainWindow::threadRunActivated(
    bool active)
{
    thread_running = active;
    ui->pb_go->setDisabled(active);
    ui->pb_stop->setEnabled(active);
    ui->pb_open_new->setDisabled(active);
    ui->frame_parameters->setDisabled(active);
}

void MainWindow::ticker(const QString &data)
{
    ui->elapsed_time->setText(data);
    timeTicks = data;

    // Go through instance rows, removing "ended" ones
    // which have not been "accepted".
    struct rempair
    {
        int key;
        QTableWidgetItem *item;
    };
    QList<rempair> to_remove;    
    for (auto it = instance_row_map.cbegin(); it != instance_row_map.cend(); ++it) {
        auto val = it.value();
        if (val.state < 0 && val.item != nullptr)
            to_remove.append({it.key(), val.item});
    }
    for (const auto &rp : to_remove) {
        //qDebug() << "?removeRow" << row << rp.key;
        auto row = rp.item->row();
        ui->instance_table->removeRow(row);
        instance_row_map.remove(rp.key);
    }
}

const int INSTANCE0 = 3;

void MainWindow::iprogress(const QString &data)
{
    auto slist = data.split(u'.');
    auto key = slist[0].toInt();
    switch (key) {
    case 0:
        ui->progress_complete->setText(slist[1] + "% @ " + slist[2]);
        break;
    case 1:
        ui->progress_hard_only->setText(slist[1] + "% @ " + slist[2]);
        break;
    case 2:
        ui->progress_unconstrained->setText(slist[1] + "% @ " + slist[2]);
        break;
    default:
        // The entry must be in the map!
        auto irow = instance_row_map.value(key);
        int row;
        if (irow.item == nullptr) {
            auto text0 = irow.data[1]; // constraint type
            // FET starts all its constraints with "Constraint",
            // which doesn't really need to be displayed ...
            if (text0.startsWith("Constraint"))
                text0.remove(0, 10);
            auto item0 = new QTableWidgetItem(text0);
            auto item1 = new QTableWidgetItem(irow.data[2]); // number of constraints
            item1->setTextAlignment(Qt::AlignCenter);
            auto item2 = new QTableWidgetItem(irow.data[3]); // timeout
            item2->setTextAlignment(Qt::AlignCenter);
            auto item3 = new QTableWidgetItem(); // @ time
            item3->setTextAlignment(Qt::AlignCenter);
            auto item4 = new QTableWidgetItem(); // progress (%)
            row = ui->instance_table->rowCount();
            ui->instance_table->insertRow(row);
            ui->instance_table->setItem(row, 0, item1);
            ui->instance_table->setItem(row, 1, item2);
            ui->instance_table->setItem(row, 2, item3);
            ui->instance_table->setItem(row, 3, item4);
            ui->instance_table->setItem(row, 4, item0);
            irow.item = item4;
            instance_row_map[key] = irow;

            QTimer::singleShot(0, [this, item4]() { //
                this->ui->instance_table->scrollToItem(item4);
            });

        } else {
            row = ui->instance_table->row(irow.item);
        }
        irow.item->setData(UserRoleInt, slist[1].toInt()); // progress (%)
        ui->instance_table->item(row, 2)->setText(slist[2]);
    }
}

void MainWindow::istart(const QString &data)
{
    auto slist = data.split(u'.');
    auto key = slist[0].toInt();
    if (key < INSTANCE0)
        return;
    instance_row_map[key] = {slist, nullptr, 0};
}

void MainWindow::iend(const QString &data)
{
    auto slist = data.split(u'.');
    auto key = slist[0].toInt();
    if (key < INSTANCE0)
        return;
    auto irow = instance_row_map[key];
    if (irow.state == 0) {
        irow.state = -1;
        instance_row_map[key] = irow;
    }
}

void MainWindow::iaccept(const QString &data)
{
    auto slist = data.split(u'.');
    auto key = slist[0].toInt();
    if (key < INSTANCE0)
        return;
    auto irow = instance_row_map[key];
    irow.state = 1;
    instance_row_map[key] = irow;
    tableProgress(irow.data[1], irow.data[2], !irow.data[4].isEmpty());
}
