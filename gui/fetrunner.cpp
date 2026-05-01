#include "fetrunner.h"
#include <QFileDialog>
#include <QMessageBox>
#include <QTimer>
#include "backend.h"
#include "globals.h"
#include "ui_fetrunner.h"

FetRunner::FetRunner(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::FetRunner)
{
    ui->setupUi(this);
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
        notifier,
        &Notifier::closeRequest,
        this,
        &FetRunner::close_request);
    connect( //
        notifier,
        &Notifier::fileChanged,
        this,
        &FetRunner::reset_display);
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
        ui->tt_processes,
        QOverload<int>::of(&QSpinBox::valueChanged),
        this,
        &FetRunner::nprocesses);
    connect( //
        ui->pb_go,
        &QPushButton::clicked,
        this,
        &FetRunner::push_go);
    connect( //
        ui->pb_stop,
        &QPushButton::clicked,
        this,
        &FetRunner::push_stop);
    connect( //
        ui->select_tmp_dir,
        &QPushButton::clicked,
        this,
        &FetRunner::select_tmp_dir);
    connect( //
        ui->default_tmp_dir,
        &QPushButton::clicked,
        this,
        &FetRunner::select_default_tmp_dir);
    connect( //
        ui->select_fet_path,
        &QPushButton::clicked,
        this,
        &FetRunner::select_fet_path);
    connect( //
        ui->default_fet_path,
        &QPushButton::clicked,
        this,
        &FetRunner::select_default_fet_path);
    connect( //
        &threadrunner,
        &RunThreadController::ticker,
        this,
        &FetRunner::ticker);
    connect( //
        &threadrunner,
        &RunThreadController::nconstraints,
        this,
        &FetRunner::nconstraints);
    connect( //
        &threadrunner,
        &RunThreadController::iprogress,
        this,
        &FetRunner::iprogress);
    connect( //
        &threadrunner,
        &RunThreadController::istart,
        this,
        &FetRunner::istart);
    connect( //
        &threadrunner,
        &RunThreadController::iend,
        this,
        &FetRunner::iend);
    connect( //
        &threadrunner,
        &RunThreadController::iaccept,
        this,
        &FetRunner::iaccept);
    connect( //
        &threadrunner,
        &RunThreadController::ieliminate,
        this,
        &FetRunner::ieliminate);
    connect(&threadrunner,
            &RunThreadController::runThreadWorkerDone,
            this,
            &FetRunner::runThreadWorkerDone);

    QValidator *validator1 = new QIntValidator(0, 99999, this);
    ui->tt_timeout->setValidator(validator1);

    QTimer::singleShot(0, this, &FetRunner::init2);
}

void FetRunner::init2()
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

FetRunner::~FetRunner()
{
    delete ui;
}

void FetRunner::nprocesses(int n)
{
    auto nn = QString::number(n);
    auto mp = backend->op1("TT_PARAMETER", {"MAXPROCESSES", nn}, "MAXPROCESSES");
    if (mp.val != nn)
        notifier->emit errorPopup("BUG: invalid number of processes: " + nn);
    ui->tt_processes->setValue(mp.val.toInt());
}

void FetRunner::reset_display()
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
    ui->completed_instance_table->setRowCount(0);
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

void FetRunner::push_go()
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

    if (backend->op("RUN_TT_SOURCE")) {
        setup_progress_table();
        threadRunActivated(true);
        threadrunner.runTtThread();
    }
}

void FetRunner::set_tmp_dir(QString tdir)
{
    QDir qtdir{tdir};
    QString d{qtdir.dirName()};
    d.prepend(QDir::separator());
    qtdir.cdUp();
    QString val{QDir::toNativeSeparators(qtdir.absolutePath())};
    ui->tmp_dir->setText(val);
    ui->tmp_dir_name->setText(d);
}

bool FetRunner::set_fet_path(QString fetpath0)
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
        } else if (fetpath == "") {
            fetpath = FET_CL;
        }
        //TODO: as continuation ...
        for (const auto &kv : backend->op("GET_FET", fetpath)) {
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

        // Show log tab in case the warnings are useful.
        ui->tabWidget->setCurrentWidget(ui->tab_log);

        QMessageBox::warning( //
            this,
            tr("FET not found"),
            tr("Seek 'FET' command-line executable in file system"));
        fetpath = "?";
    }
    return true;
}

void FetRunner::push_stop()
{
    ui->pb_stop->setEnabled(false);
    threadrunner.stopThread();
    closingMessageBox.setText(tr("Finishing ..."));
    closingMessageBox.setIcon(QMessageBox::Information);
    closingMessageBox.setStandardButtons(QMessageBox::NoButton);
    closingMessageBox.exec();
}

void FetRunner::select_tmp_dir()
{
    QString dirpath = QFileDialog::getExistingDirectory( //
        this,
        tr("Select base folder for temporary files"),
        "/",
        QFileDialog::ShowDirsOnly);
    if (!dirpath.isEmpty()) {
        if (!backend->op("TMP_PATH " + dirpath)) {
            ui->tmp_dir->clear();
            ui->tmp_dir_name->setText("-");
        } else {
            //TODO: The path will be in the log as "$ TMP_DIR=path"
            //set_tmp_dir(path);
        }
    }
}

void FetRunner::select_default_tmp_dir()
{
    if (!backend->op("TMP_PATH")) {
        ui->tmp_dir->clear();
        ui->tmp_dir_name->setText("-");
    } else {
        //TODO: The path will be in the log as "$ TMP_DIR=path"
        //set_tmp_dir(path);
    }
}

void FetRunner::select_fet_path()
{
    set_fet_path("?");
}

void FetRunner::select_default_fet_path()
{
    set_fet_path("");
}

void FetRunner::runThreadWorkerDone()
{
    //qDebug() << "threadRunFinished" << "FetRunner";
    threadRunActivated(false);
    closingMessageBox.hide();
    notifier->emit finished("FetRunner");
}

void FetRunner::close_request()
{
    //qDebug() << "close_request()" << thread_running;
    if (thread_running) {
        notifier->emit quit_register_wait("FetRunner");
        push_stop();
    } else {
        notifier->emit finished("FetRunner");
    }
}

void FetRunner::threadRunActivated(bool active)
{
    thread_running = active;
    ui->pb_go->setDisabled(active);
    ui->pb_stop->setEnabled(active);

    notifier->emit setBusy(active);

    ui->frame_parameters->setDisabled(active);

    //TODO: If !active I could test whether there is a result
    // and set up the timetable viewer.
    //TODO: If active, the timetable viewer should be disabled.
    if (active) {

    } else {
        notifier->emit new_tt_data();
    }
}

void FetRunner::ticker(const QString &data)
{
    // The last call here has an empty string, so that things
    // can be tidied up a bit.
    if (!data.isEmpty()) {
        ui->elapsed_time->setText(data);
        timeTicks = data;
    }

    // Go through instance rows, removing "ended" ones.
    // If accepted (state = 1), add it to the "completed" table.
    struct rmdata
    {
        int key;
        instance_row irow;
        //TODO--int state;
        //TODO--QTableWidgetItem *item;
    };
    QList<rmdata> to_remove;
    for (auto it = instance_row_map.cbegin(); it != instance_row_map.cend(); ++it) {
        auto val = it.value();
        if (val.state != 0 && val.item != nullptr)
            to_remove.append({it.key(), val});
    }
    for (const auto &rp : to_remove) {
        //qDebug() << "?removeRow" << row << rp.key;
        auto irow = rp.irow;
        if (irow.state == 1) {
            auto ctype = irow.data[1]; // constraint type
            add_completed_instance(    //
                irow.data[2],          // number of constraints
                QString{"/ %1"}.arg(constraint_map[ctype].total),
                ctype);
        }
        auto row = irow.item->row();
        ui->instance_table->removeRow(row);
        instance_row_map.remove(rp.key);
    }

    //TODO--ui->instance_table->scrollToBottom();

    //TODO: if (new rows)
    ui->completed_instance_table->scrollToBottom();

    // Changes to progress table
    for (const auto &update : std::as_const(progress_rows_changed)) {
        tableProgress(update);
    }
    progress_rows_changed.clear();
}

void FetRunner::add_completed_instance(
    //
    QString number,
    QString total,
    QString ctype)
//QString message)
{
    auto nrow = ui->completed_instance_table->rowCount();
    ui->completed_instance_table->insertRow(nrow);
    ui->completed_instance_table->setVerticalHeaderItem( //
        nrow,
        new QTableWidgetItem(QString("%1").arg(nrow + 1, 2)));
    auto item0 = new QTableWidgetItem(ctype);
    auto item1 = new QTableWidgetItem(number); // number of constraints
    auto item2 = new QTableWidgetItem(total);
    //if (!message.isEmpty()) {
    //    item1->setToolTip(message);
    //}
    item1->setTextAlignment(Qt::AlignCenter);
    item2->setTextAlignment(Qt::AlignCenter);
    ui->completed_instance_table->setItem(nrow, 0, item1);
    ui->completed_instance_table->setItem(nrow, 1, item2);
    ui->completed_instance_table->setItem(nrow, 2, item0);
}

void FetRunner::iprogress(const QString &data)
{
    //qDebug() << "iprogress:" << data;
    QStringList slist = data.split(u'.');
    // slist: instance index, percent complete, instance run time
    auto key = slist[0].toInt();
    switch (key) {
    case INSTANCE_COMPLETE:
        ui->progress_complete->setText(slist[1] + "% @ " + slist[2]);
        break;
    case INSTANCE_HARD_ONLY:
        ui->progress_hard_only->setText(slist[1] + "% @ " + slist[2]);
        break;
    case INSTANCE_PRIORITY:
        break;
    case INSTANCE_UNCONSTRAINED:
        ui->progress_unconstrained->setText(slist[1] + "% @ " + slist[2]);
        break;
    default: // constaint-type tests
        instanceRowProgress(key, slist);
    }
}

void FetRunner::istart(const QString &data)
{
    //qDebug() << "istart:" << data;
    auto slist = data.split(u'.');
    // slist: instance index, constraint type,
    // number of individual constraints, time-out
    auto key = slist[0].toInt();
    if (key < 0)
        return;
    instance_row_map[key] = {slist, nullptr, 0};
}

void FetRunner::iend(const QString &data)
{
    auto slist = data.split(u'.');
    auto key = slist[0].toInt();
    switch (key) {
    case INSTANCE_COMPLETE:
        ui->progress_complete->setEnabled(false);
        break;
    case INSTANCE_HARD_ONLY:
        ui->progress_hard_only->setEnabled(false);
        break;
    case INSTANCE_UNCONSTRAINED:
        ui->progress_unconstrained->setEnabled(false);
    case INSTANCE_PRIORITY:
        break;
    default:
        auto irow = instance_row_map[key];
        if (irow.state == 0) {
            irow.state = -1;
            instance_row_map[key] = irow;
        }
    }
}

void FetRunner::iaccept(const QString &data)
{
    //qDebug() << "iaccept:" << data;
    auto slist = data.split(u'.');
    auto key = slist[0].toInt();
    switch (key) {
    case INSTANCE_COMPLETE: // "full" completed
        tableProgressGroupDone(-1);
        break;
    case INSTANCE_HARD_ONLY: // "all hard" completed
        tableProgressGroupDone(0);
        break;
    case INSTANCE_PRIORITY: // "priority" completed
        tableProgressGroupDone(1);
        break;
    case INSTANCE_UNCONSTRAINED: // "unconstrained" completed
        break;
    default:
        instance_row &irow = instance_row_map[key];
        irow.state = 1;
        //if (!instance_rows_changed.contains(key))
        //    instance_rows_changed.append(key);
        progress_rows_changed.append({irow.data[1], irow.data[2]});
    }
}

void FetRunner::ieliminate(const QString &data)
{
    auto slist = data.split(u'.');
    auto ctype = slist[0];
    add_completed_instance( //
        QString{"--- [%1]"}.arg(slist[1]),
        QString{"/ %1"}.arg(constraint_map[ctype].total),
        ctype);
}
