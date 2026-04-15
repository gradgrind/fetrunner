#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QWidget>

namespace Ui {

class MainWindow;

}

class MainWindow : public QWidget
{
    Q_OBJECT

public:
    explicit MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

private:
    Ui::MainWindow *ui;

    void open_file();
    void set_busy(bool on);

public slots:
    void error_popup(const QString &msg);
};

#endif // MAINWINDOW_H
