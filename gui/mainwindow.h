#ifndef MAINWINDOW_H
#define MAINWINDOW_H
#include <QCloseEvent>
#include <QMessageBox>
#include <QSettings>
#include <QTableWidgetItem>
#include <QWidget>
#include "threadrun.h"

#ifdef Q_OS_WIN
const QString FET_CL = "fet-cl.exe";
#else
const QString FET_CL = "fet-cl";
#endif

QT_BEGIN_NAMESPACE
namespace Ui {

class MainWindow;

}
QT_END_NAMESPACE

struct instance_row
{
    QStringList data;
    QTableWidgetItem *item;
    int state;
};

class MainWindow : public QWidget
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

private:
    QSettings *settings;
    Ui::MainWindow *ui;
    void closeEvent(QCloseEvent *e) override;
    void resizeEvent(QResizeEvent *) override;
    void init2();
    void resizeColumns();

    bool quit_requested{false};
    bool thread_running{false};
    QString filename{};
    QString filedir{};
    QString datatype{};
    QMessageBox closingMessageBox;
    void threadRunActivated(bool active);

    RunThreadController threadrunner;
    QString timeTicks{};
    QHash<int, instance_row> instance_row_map;

    QString hard_count;
    QString soft_count;

public slots:
    void error_popup(const QString &msg);

private slots:
    void nprocesses(int n);
    void open_file();
    void push_go();
    void push_stop();
    void ticker(const QString &data);
    void nconstraints(const QString &data);
    void progress(const QString &data);
    void istart(const QString &data);
    void iend(const QString &data);
    void iaccept(const QString &data);
    void runThreadWorkerDone();
};
#endif // MAINWINDOW_H
