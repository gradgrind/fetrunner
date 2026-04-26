#ifndef BACKEND_H
#define BACKEND_H

#include <QColor>
#include <QObject>
#include <QStringList>

struct KeyVal
{
    QString key;
    QString val;
};

class Backend : public QObject
{
    Q_OBJECT

public:
    Backend();
    //~Backend();

    QList<KeyVal> op(QString cmd, QStringList data = {});
    KeyVal readlogline();
    KeyVal readresult(QString r);

    KeyVal op1(QString cmd, QStringList data = {}, QString key = {});
    QString getConfig(QString key, QString fallback = {});
    void setConfig(QString key, QString val);

private:
    QString logline;

signals:
    void logcolour(QColor);
    void log(QString);
    void error(QString);
};

extern Backend *backend;

#endif // BACKEND_H
