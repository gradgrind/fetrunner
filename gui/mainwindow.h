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
    void closeEvent(QCloseEvent *e) override;
    void quit_register_wait(QString module);

    void open_file();
    void set_busy(bool on);

    bool quit_requested{false};
    bool quit_confirmed{false};
    QStringList waiting_on;

public slots:
    void error_popup(const QString msg);
    void handle_finished(QString module);
};

#endif // MAINWINDOW_H
