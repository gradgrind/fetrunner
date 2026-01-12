#include "mainwindow.h"
#include <QFileDialog>
#include <QMessageBox>
#include <QTimer>
#include "backend.h"
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
        QOverload<int>::of(&QSpinBox::valueChanged),
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
        ui->select_tmp_dir,
        &QPushButton::clicked,
        this,
        &MainWindow::select_tmp_dir);
    connect( //
        ui->default_tmp_dir,
        &QPushButton::clicked,
        this,
        &MainWindow::select_default_tmp_dir);
    connect( //
        ui->select_fet_path,
        &QPushButton::clicked,
        this,
        &MainWindow::select_fet_path);
    connect( //
        ui->default_fet_path,
        &QPushButton::clicked,
        this,
        &MainWindow::select_default_fet_path);
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
        &RunThreadController::iprogress,
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
    if (!set_fet_path(settings->value("fet/FetPath").toString())) {
        QApplication::exit(1);
        return;
    }

    // Set/show default temporary directory
    select_default_tmp_dir();
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
    ui->progress_complete->setEnabled(true);
    ui->progress_hard_only->clear();
    ui->progress_hard_only->setEnabled(true);
    ui->progress_hard->setValue(0);
    ui->progress_hard->setEnabled(false);
    ui->label_hard->setEnabled(false);
    ui->progress_soft->setValue(0);
    ui->progress_soft->setEnabled(false);
    ui->label_soft->setEnabled(false);
    ui->progress_unconstrained->clear();
    ui->progress_unconstrained->setEnabled(true);
    hard_count.clear();
    soft_count.clear();
    timeTicks.clear();
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
    instance_row_map.clear();
    progress_rows_changed.clear();
    reset_display();

    // Set parameters
    auto t = ui->tt_timeout->text();
    backend->op("TT_PARAMETER", {"TIMEOUT", t});
    auto sh = ui->tt_skip_hard->isChecked();
    backend->op("TT_PARAMETER", {"SKIP_HARD", sh ? "true" : "false"});
    auto rs = ui->tt_real_soft->isChecked();
    backend->op("TT_PARAMETER", {"REAL_SOFT", rs ? "true" : "false"});
    auto wff = ui->write_fet_file->isChecked();
    backend->op("TT_PARAMETER", {"WRITE_FET_FILE", wff ? "true" : "false"});

    for (const auto &kv : backend->op("RUN_TT_SOURCE")) {
        if (kv.key == "TMP_DIR") {
            set_tmp_dir(kv.val);
        } else if (kv.key == "OK" && kv.val == "true") {
            setup_progress_table();
            threadRunActivated(true);
            threadrunner.runTtThread();
        }
    }
}

void MainWindow::set_tmp_dir(QString tdir)
{
    QDir qtdir{tdir};
    QString d{qtdir.dirName()};
    d.prepend(QDir::separator());
    qtdir.cdUp();
    QString val{QDir::toNativeSeparators(qtdir.absolutePath())};
    ui->tmp_dir->setText(val);
    ui->tmp_dir_name->setText(d);
}

bool MainWindow::set_fet_path(QString fetpath0)
{
    auto fetpath = fetpath0;
    QString fetv;
    QString fetp;
    while (true) {
        if (fetpath == "?") {
            fetpath = QFileDialog::getOpenFileName( //
                this,
                tr("Seek FET executable"),
                QDir::homePath(),
                tr("FET executable") + " (" + FET_CL + ")");
            if (fetpath.isEmpty()) {
                return false;
            }
        }
        for (const auto &kv : backend->op("GET_FET", {fetpath, "W"})) {
            if (kv.key == "FET_PATH")
                fetp = kv.val;
            else if (kv.key == "FET_VERSION")
                fetv = kv.val;
        }
        if (!fetp.isEmpty()) {
            settings->setValue("fet/FetPath", fetpath);
            // Set GUI
            ui->fet_path->setText(fetp);
            ui->fet_version->setText(fetv);
            break;
        }

        // Handle FET executable not found.

        if (!fetpath.isEmpty()) {
            // Try the default.
            fetpath.clear();
            continue;
        }

        QMessageBox::warning( //
            this,
            tr("FET not found"),
            tr("Seek 'FET' command-line executable in file system"));
        fetpath = "?";
    }
    return true;
}

void MainWindow::push_stop()
{
    ui->pb_stop->setEnabled(false);
    threadrunner.stopThread();
    closingMessageBox.setText(tr("Finishing ..."));
    closingMessageBox.setIcon(QMessageBox::Information);
    closingMessageBox.setStandardButtons(QMessageBox::NoButton);
    closingMessageBox.exec();

    //TODO--
    //dump_log("dump.log");
}

void MainWindow::select_tmp_dir()
{
    QString dirpath = QFileDialog::getExistingDirectory( //
        this,
        tr("Select base folder for temporary files"),
        "/",
        QFileDialog::ShowDirsOnly);
    if (!dirpath.isEmpty()) {
        auto kv = backend->op1("TMP_PATH", {dirpath}, "TMP_DIR");
        if (kv.key == "") {
            ui->tmp_dir->clear();
            ui->tmp_dir_name->setText("-");
        } else {
            set_tmp_dir(kv.val);
        }
    }
}

void MainWindow::select_default_tmp_dir()
{
    auto kv = backend->op1("TMP_PATH", {""}, "TMP_DIR");
    if (kv.key == "") {
        ui->tmp_dir->clear();
        ui->tmp_dir_name->setText("-");
    } else {
        set_tmp_dir(kv.val);
    }
}

void MainWindow::select_fet_path()
{
    set_fet_path("?");
}

void MainWindow::select_default_fet_path()
{
    set_fet_path("");
}

void MainWindow::runThreadWorkerDone()
{
    //qDebug() << "threadRunFinished";
    threadRunActivated(false);
    closingMessageBox.hide();
    if (quit_requested)
        close();
}

void MainWindow::threadRunActivated(bool active)
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
    struct rmdata
    {
        int key;
        int state;
        QTableWidgetItem *item;
    };
    QList<rmdata> to_remove;
    for (auto it = instance_row_map.cbegin(); it != instance_row_map.cend(); ++it) {
        auto val = it.value();
        if (val.state != 0 && val.item != nullptr)
            to_remove.append({it.key(), val.state, val.item});
    }
    for (const auto &rp : to_remove) {
        //qDebug() << "?removeRow" << row << rp.key;
        if (rp.state == 1) {
            rp.item->setText("+++");
        } else {
            auto row = rp.item->row();
            ui->instance_table->removeRow(row);
        }
        instance_row_map.remove(rp.key);
    }

    //TODO--
    ui->instance_table->scrollToBottom();

    ui->completed_instance_table->scrollToBottom();

    // Changes to progress table
    for (const auto &update : std::as_const(progress_rows_changed)) {
        tableProgress(update);
    }
    progress_rows_changed.clear();
}

const int INSTANCE0 = 3;

void MainWindow::iprogress(const QString &data)
{
    QStringList slist = data.split(u'.');
    // slist: instance index, percent complete, instance run time
    // Instance 0: fully constrained
    // Instance 1: all hard constraints
    // Instance 2: no constraints
    // Other instances: constraint-type tests
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
        instanceRowProgress(key, slist);
    }
}

void MainWindow::istart(const QString &data)
{
    auto slist = data.split(u'.');
    // slist: instance index, constraint type,
    // number of individual constraints, time-out
    auto key = slist[0].toInt();
    if (key < INSTANCE0)
        return;
    instance_row_map[key] = {slist, nullptr, 0};
}

void MainWindow::iend(const QString &data)
{
    auto slist = data.split(u'.');
    auto key = slist[0].toInt();
    switch (key) {
    case 0:
        ui->progress_complete->setEnabled(false);
        break;
    case 1:
        ui->progress_hard_only->setEnabled(false);
        break;
    case 2:
        ui->progress_unconstrained->setEnabled(false);
        break;
    default:
        auto irow = instance_row_map[key];
        if (irow.state == 0) {
            irow.state = -1;
            instance_row_map[key] = irow;
        }
    }
}

void MainWindow::iaccept(const QString &data)
{
    auto slist = data.split(u'.');
    auto key = slist[0].toInt();
    switch (key) {
    case 0: // "full" completed
        //TODO--tableProgressAll();
        tableProgressSet(false);
        break;
    case 1: // "all hard" completed
        //TODO--tableProgressHard();
        tableProgressSet(true);
        break;
    case 2: // "unconstrained" completed
        return;
    default:
        instance_row &irow = instance_row_map[key];
        irow.state = 1;
        //if (!instance_rows_changed.contains(key))
        //    instance_rows_changed.append(key);
        progress_rows_changed.append({irow.data[1], irow.data[2]});
    }
}
