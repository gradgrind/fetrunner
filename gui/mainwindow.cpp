#include "mainwindow.h"
#include <QFileDialog>
#include <QMessageBox>
#include "backend.h"
#include "ui_mainwindow.h"

Backend *backend;

MainWindow::MainWindow(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::MainWindow)
{
    ui->setupUi(this);
    ui->tableWidget->resizeColumnsToContents();

    backend = new Backend();
    connect(backend, &Backend::log, ui->logview, &QTextEdit::append);
    connect(backend, &Backend::error, this, &MainWindow::error_popup);
    connect(ui->pb_open_new, &QPushButton::clicked, this, &MainWindow::open_file);
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

void MainWindow::open_file()
{
    //qDebug() << "Open File";

    //TODO?
    if (running)
        return;

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
