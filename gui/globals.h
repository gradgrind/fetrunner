#ifndef GLOBALS_H
#define GLOBALS_H

#include <QSettings>
#include <QString>

extern QString file_dir;
extern QString file_name;
extern QString file_datatype;

extern QSettings *settings;

class Notifier : public QObject
{
    Q_OBJECT

signals:
    void fileChanged();
    void setBusy(bool on);
    void errorPopup(QString msg);
    void closeRequest();
    void quit_register_wait(QString module);
    void finished(QString module);
    void new_tt_data();
};

extern Notifier *notifier;

#endif // GLOBALS_H
