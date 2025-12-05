#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QCloseEvent>
#include <QWidget>
#include "threadrun.h"

QT_BEGIN_NAMESPACE
namespace Ui {

class MainWindow;

}
QT_END_NAMESPACE

enum UserRoleN { // Used for table data types
    UserRoleInt = Qt::UserRole
};

class MainWindow : public QWidget
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

private:
    Ui::MainWindow *ui;
    void closeEvent(QCloseEvent *e);

    QString filename{};
    QString filedir{};
    QString datatype{};
    void threadRunActivated(bool active);

    RunThreadController threadrunner;

public slots:
    void error_popup(QString msg);
    void threadRunFinished();

private slots:
    void open_file();
    void push_go();
};
#endif // MAINWINDOW_H
