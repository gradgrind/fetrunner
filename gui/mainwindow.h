#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QWidget>
#include "ttrun.h"

QT_BEGIN_NAMESPACE
namespace Ui {

class MainWindow;

}
QT_END_NAMESPACE

class MainWindow : public QWidget
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

private:
    Ui::MainWindow *ui;

    QString filename{};
    QString filedir{};
    QString datatype{};
    bool running{false};
    TtRun *ttrun;

public slots:
    void error_popup(QString msg);
    void push_go();

private slots:
    void open_file();
};
#endif // MAINWINDOW_H
