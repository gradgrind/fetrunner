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

struct progress_line
{
    int index;
    int progress;
    int total;
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
    void init2();
    void reset_display();
    void init_ttgen_tables();    // at start of program
    void setup_progress_table(); // when starting ttgen run
    void setup_instance_table(); // when starting ttgen run
    void tableProgress(QString constraint, QString number, bool hard);
    void tableProgressAll();
    void tableProgressHard();
    QString constraint_name(QString name);

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
    int instance_table_fixed_width;

    QString hard_count;
    QString soft_count;
    QHash<QString, progress_line> hard_constraint_map;
    QHash<QString, progress_line> soft_constraint_map;

public slots:
    void error_popup(const QString &msg);

private slots:
    void nprocesses(int n);
    void open_file();
    void push_go();
    void push_stop();
    void ticker(const QString &data);
    void nconstraints(const QString &data);
    void iprogress(const QString &data);
    void istart(const QString &data);
    void iend(const QString &data);
    void iaccept(const QString &data);
    void runThreadWorkerDone();
};
#endif // MAINWINDOW_H
