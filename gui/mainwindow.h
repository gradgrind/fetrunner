#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QWidget>
#include "backend.h"

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
    bool running{false};

public slots:
    void error_popup(QString msg);

private slots:
    void open_file();
};
#endif // MAINWINDOW_H
