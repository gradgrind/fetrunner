#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QSettings>
#include <QWidget>

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
    QSettings *settings;
    bool running{false};

    void set_connections();

private slots:
    void open_file();
};
#endif // MAINWINDOW_H
