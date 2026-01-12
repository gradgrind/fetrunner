#ifndef MAINWINDOW_H
#define MAINWINDOW_H
#include <QCloseEvent>
#include <QMessageBox>
#include <QSettings>
#include <QTableWidgetItem>
#include <QWidget>
#include "threadrun.h"

#ifdef Q_OS_WIN
const QString FET_CL = "fet-clw.exe";
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
    int state; // 0: running, -1: stopped, 1: stopped & accepted
};

struct progress_line
{
    int index;
    int progress;
    int total;
};

struct progress_changed
{
    QString constraint;
    QString number;
};

class MainWindow : public QWidget
{
    Q_OBJECT

public:
    MainWindow(QWidget *parent = nullptr);
    ~MainWindow();

    bool dump_log(QString fname);

private:
    QSettings *settings;
    Ui::MainWindow *ui;
    void closeEvent(QCloseEvent *e) override;
    void fail(QString msg);
    void init2();
    void reset_display();
    void init_ttgen_tables();    // at start of program
    void setup_progress_table(); // when starting ttgen run
    void setup_instance_table(); // when starting ttgen run
    void tableProgress(progress_changed update);
    void tableProgressSet(bool hard_only);
    void tableProgressAll();
    void tableProgressHard();
    void tableProgressGroup(QHash<QString, progress_line>);
    void instanceRowProgress(int key, QStringList parms);
    //QString constraint_name(QString name);
    void set_tmp_dir(QString tdir);
    bool set_fet_path(QString fetpath);

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
    QHash<QString, progress_line> constraint_map;
    //TODO--QHash<QString, progress_line> hard_constraint_map;
    //TODO--QHash<QString, progress_line> soft_constraint_map;
    //QList<int> instance_rows_changed;
    QList<progress_changed> progress_rows_changed;

public slots:
    void error_popup(const QString &msg);

private slots:
    void nprocesses(int n);
    void open_file();
    void push_go();
    void push_stop();
    void select_tmp_dir();
    void select_default_tmp_dir();
    void select_fet_path();
    void select_default_fet_path();
    void ticker(const QString &data);
    void nconstraints(const QString &data);
    void iprogress(const QString &data);
    void istart(const QString &data);
    void iend(const QString &data);
    void iaccept(const QString &data);
    void runThreadWorkerDone();
};
#endif // MAINWINDOW_H
