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

    //ui->instance_table->horizontalHeader()->setSectionResizeMode(QHeaderView::Fixed);
    //ui->instance_table->resizeColumnsToContents();
    QTimer::singleShot(0, this, &MainWindow::resizeColumns);
    ui->instance_table->setItemDelegateForColumn( //
        4,
        new ProgressDelegate(ui->instance_table));
    //ui->specials_table->horizontalHeader()->setSectionResizeMode(QHeaderView::Fixed);
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
        &MainWindow::progress);
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

void MainWindow::resizeEvent(QResizeEvent *)
{
    resizeColumns();
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
                ui->logview->clear();
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
    ui->instance_table->setRowCount(0);
    if (backend->op1("RUN_TT_SOURCE", {}, "OK").val == "true") {
        backend->op("HARD_CONSTRAINTS");
        backend->op("SOFT_CONSTRAINTS");

        threadRunActivated(true);
        //ui->pb_stop->setEnabled(true);
        ui->elapsed_time->setText("0");
        for (int i = 0; i < 3; ++i) {
            ui->specials_table->item(i, 0)->setText("");
            ui->specials_table->item(i, 1)->setData(UserRoleInt, 0);
        }
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

    // Go through instance rows, removing "ended" ones
    // which have not been "accepted".
    struct rempair
    {
        int key;
        QTableWidgetItem *item;
    };
    QList<rempair> to_remove;
    for (const auto &[key, val] : std::as_const(instance_row_map).asKeyValueRange()) {
        if (val.state < 0 && val.item != nullptr) {
            to_remove.append({key, val.item});
        }
    }
    for (const auto &rp : to_remove) {
        //qDebug() << "?removeRow" << row << rp.key;
        auto row = rp.item->row();
        ui->instance_table->removeRow(row);
        instance_row_map.remove(rp.key);
    }
}

void MainWindow::nconstraints(const QString &data)
{
    auto slist = data.split(u'.');
    ui->c_enabled_t->setText(timeTicks);
    ui->c_enabled_h->setText(slist[0] + " / " + slist[1]);
    ui->c_enabled_s->setText(slist[2] + " / " + slist[3]);
}

const int INSTANCE0 = 3;

void MainWindow::progress(const QString &data)
{
    auto slist = data.split(u'.');
    auto key = slist[0].toInt();
    if (key < INSTANCE0) {
        ui->specials_table->item(key, 0)->setText(slist[2]);
        ui->specials_table->item(key, 1)->setData(UserRoleInt, slist[1].toInt());
    } else {
        // The entry must be in the map!
        auto irow = instance_row_map.value(key);
        int row;
        if (irow.item == nullptr) {
            auto text0 = irow.data[1];
            // FET starts all its constraints with "Constraint",
            // which doesn't really need to be displayed ...
            if (text0.startsWith("Constraint"))
                text0.remove(0, 10);
            auto item0 = new QTableWidgetItem(text0);
            auto item1 = new QTableWidgetItem(irow.data[2]);
            item1->setTextAlignment(Qt::AlignCenter);
            auto item2 = new QTableWidgetItem(irow.data[3]);
            item2->setTextAlignment(Qt::AlignCenter);
            auto item3 = new QTableWidgetItem();
            item3->setTextAlignment(Qt::AlignCenter);
            auto item4 = new QTableWidgetItem();
            row = ui->instance_table->rowCount();
            ui->instance_table->insertRow(row);
            ui->instance_table->setItem(row, 0, item0);
            ui->instance_table->setItem(row, 1, item1);
            ui->instance_table->setItem(row, 2, item2);
            ui->instance_table->setItem(row, 3, item3);
            ui->instance_table->setItem(row, 4, item4);
            irow.item = item4;
            instance_row_map[key] = irow;

            //ui->instance_table->scrollToItem(item4); // ensure new row visible

            QTimer::singleShot(0, [this, item4]() { //
                this->ui->instance_table->scrollToItem(item4);
            });

        } else {
            row = ui->instance_table->row(irow.item);
        }
        irow.item->setData(UserRoleInt, slist[1].toInt());
        ui->instance_table->item(row, 3)->setText(slist[2]);
    }
}

void MainWindow::resizeColumns()
{
    QFontMetrics fm(ui->instance_table->font());
    int table_width = ui->instance_table->width();
    int w = 0;
    for (auto col = 1; col < 5; ++col) {
        auto headerItem = ui->instance_table->horizontalHeaderItem(col);
        auto text = headerItem->text();
        int col_width = fm.horizontalAdvance(text) + 10; // add some padding
        if (col == 4 && col_width < 100)
            col_width = 120;
        w += col_width;
        ui->instance_table->setColumnWidth(col, col_width);
    }
    int wsb = qApp->style()->pixelMetric(QStyle::PM_ScrollBarExtent);
    ui->instance_table->setColumnWidth(0, table_width - w - wsb - 10);
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
}
