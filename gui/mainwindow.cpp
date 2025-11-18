#include "mainwindow.h"
#include <QFileDialog>
#include "backend.h"
#include "ui_mainwindow.h"

MainWindow::MainWindow(QWidget *parent)
    : QWidget(parent)
    , ui(new Ui::MainWindow)
{
    ui->setupUi(this);
    ui->tableWidget->resizeColumnsToContents();

    settings = new QSettings("gradgrind", "fetrunner");
    //const auto geometry = settings->value("MainWindow", QByteArray()).toByteArray();
    const auto geometry = settings->value("MainWindowSize").value<QSize>();
    if (!geometry.isEmpty())
        //restoreGeometry(geometry);
        resize(geometry);

    set_connections();

    QString s{tr("Hello")};
    ui->lineEdit_3->setText(s);
    QString s2{tr("Something")};
    ui->lineEdit_4->setText(s2);
}

MainWindow::~MainWindow()
{
    //settings->setValue("MainWindow", saveGeometry());
    settings->setValue("MainWindowSize", size());
    delete settings;

    delete ui;
}

void MainWindow::set_connections()
{
    QObject::connect(ui->pb_open_new, &QPushButton::clicked, this, &MainWindow::open_file);
}

void MainWindow::open_file()
{
    qDebug() << "Open File";

    if (running)
        return;

    QString opendir{settings->value("SourceDir").toString()};
    if (opendir.isEmpty())
        opendir = QDir::homePath();
    QString fileName = QFileDialog::getOpenFileName( //
        this,
        tr("Open Timetable Specifiation"),
        opendir,
        tr("FET / W365 Files (*.fet *_w365.json)"));

    if (fileName.isEmpty())
        return;

    qDebug() << "Open:" << fileName;

    QDir dir(fileName);
    if (dir.cdUp())
        settings->setValue("SourceDir", dir.absolutePath());
    qDebug() << "Dir:" << dir.absolutePath();

    qDebug() << "???" << test_backend(fileName);
}
