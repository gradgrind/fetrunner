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
    ui->specials_table->setItemDelegateForColumn( //
        1,
        new ProgressDelegate(ui->specials_table));

    auto it_progress0 = new QTableWidgetItem();
    it_progress0->setData(Qt::UserRole + 1000, 30);
    ui->specials_table->setItem(0, 1, it_progress0);
    auto it_progress1 = new QTableWidgetItem();
    it_progress1->setData(Qt::UserRole + 1000, 50);
    ui->specials_table->setItem(1, 1, it_progress1);
    auto it_progress2 = new QTableWidgetItem();
    it_progress2->setData(Qt::UserRole + 1000, 80);
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

    //TODO--
    QString s1{tr("Hello")};
    ui->lineEdit_3->setText(s1);
    QString s2{tr("Something")};
    ui->lineEdit_4->setText(s2);
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

//TODO: adapt for progress bar
void ProgressDelegate::paint(
    //
    QPainter *painter,
    const QStyleOptionViewItem &option,
    const QModelIndex &index) const
{
    auto progress = index.data(Qt::UserRole + 1000).toInt();

    qDebug() << "P" << progress;

    QStyleOptionProgressBar progbar;
    progbar.rect = option.rect;
    progbar.minimum = 0;
    progbar.maximum = 100;
    progbar.progress = progress;
    progbar.text = QString::number(progress);
    progbar.textVisible = true;
    QApplication::style()->drawControl( //
        QStyle::CE_ProgressBar,
        &progbar,
        painter);
}

/*
data = [("1", "Baharak", 10), ("2", "Darwaz", 60),
        ("3", "Fays abad", 20), ("4", "Ishkashim", 80), 
        ("5", "Jurm", 100)]

class ProgressDelegate(QtWidgets.QStyledItemDelegate):
    def paint(self, painter, option, index):
        progress = index.data(QtCore.Qt.UserRole+1000)
        opt = QtWidgets.QStyleOptionProgressBar()
        opt.rect = option.rect
        opt.minimum = 0
        opt.maximum = 100
        opt.progress = progress
        opt.text = "{}%".format(progress)
        opt.textVisible = True
        QtWidgets.QApplication.style().drawControl(QtWidgets.QStyle.CE_ProgressBar, opt, painter)

if __name__ == '__main__':
    import sys
    app = QtWidgets.QApplication(sys.argv)
    w = QtWidgets.QTableWidget(0, 3)
    delegate = ProgressDelegate(w)
    w.setItemDelegateForColumn(2, delegate)

    w.setHorizontalHeaderLabels(["ID", "Name", "Progress"])
    for r, (_id, _name, _progress) in enumerate(data):
        it_id = QtWidgets.QTableWidgetItem(_id)
        it_name = QtWidgets.QTableWidgetItem(_name)
        it_progress = QtWidgets.QTableWidgetItem()
        it_progress.setData(QtCore.Qt.UserRole+1000, _progress)
        w.insertRow(w.rowCount())
        for c, item in enumerate((it_id, it_name, it_progress)):
            w.setItem(r, c, item)
    w.show()
    sys.exit(app.exec_())
*/
