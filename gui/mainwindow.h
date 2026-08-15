#ifndef MAINWINDOW_H
#define MAINWINDOW_H

#include <QWidget>
#include <QTextEdit>
#include "fetrunner.h"
#include "ttview.h"
#include "ttviewselector.h"

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
    void switch_logger(QString msg, int log_viewer);
    void do_FETRUNNER_VERSION(const QString &val);
    void do_SET_FILE(const QString &val);
    void do_DATA_TYPE(const QString &val);
    void logLine(QString line);
    void setLogColour(QColor colour);
    void clearLog(int logger);
    void dumpLog(int logger);
    void showLogger(int logger);
    QTextEdit *selectLogger(int logger);

    bool quit_requested{false};
    bool quit_confirmed{false};
    QStringList waiting_on;
    FetRunner *ttsolver{nullptr};
    TtView *ttview{nullptr};
    TtViewSelector *ttviewselector{nullptr};
    QTextEdit *log_view{nullptr};

public slots:
    void error_popup(const QString msg);
    void handle_finished(QString module);
    void new_file();
    void do_new_tt_data();
    void do_no_tt_data();
};

#endif // MAINWINDOW_H
