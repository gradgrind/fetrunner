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
};

extern Notifier *notifier;

#endif // GLOBALS_H
