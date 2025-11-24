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

    if (running)
        return;

    QString opendir = backend->getConfig("gui/SourceDir");
    //QString opendir{settings->value("SourceDir").toString()};
    if (opendir.isEmpty())
        opendir = QDir::homePath();
    QString fileName = QFileDialog::getOpenFileName( //
        this,
        tr("Open Timetable Specifiation"),
        opendir,
        tr("FET / W365 Files (*.fet *_w365.json)"));

    if (fileName.isEmpty())
        return;

    //qDebug() << "Open:" << fileName;

    QDir dir(fileName);
    if (dir.cdUp())
        backend->setConfig("gui/SourceDir", dir.absolutePath());
    //qDebug() << "Dir:" << dir.absolutePath();

    backend->op("SET_FILE", {fileName});
}

void MainWindow::error_popup(QString msg)
{
    QMessageBox::critical(this, "", msg);
}
