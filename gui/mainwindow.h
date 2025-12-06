#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QCloseEvent>
#include <QTableWidgetItem>
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

struct instance_row
{
    QStringList data;
    QTableWidgetItem *item;
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
    QString timeTicks{};
    QHash<int, instance_row> instance_row_map;

public slots:
    void error_popup(const QString &msg);

private slots:
    void open_file();
    void push_go();
    void ticker(const QString &data);
    void nconstraints(const QString &data);
    void progress(const QString &data);
    void istart(const QString &data);
    void iend(const QString &data);
    void iaccept(const QString &data);
    void runThreadWorkerDone();
};
#endif // MAINWINDOW_H
