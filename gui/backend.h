#ifndef BACKEND_H
#define BACKEND_H

#include <QColor>
#include <QObject>
#include <QString>
#include <QThread>
#include <QHash>

/*
The dispatcher (Backend::op) runs in the main thread, so it blocks
until the call has completed. This should be fine in most cases, but
"RUN_TT" takes a long time. So an additional command, "!RUN_TT", is
available, which runs the function in a separate goroutine.
While this is running, no further commands (except "_STOP_TT") should
be run. The GUI should adapt to this state.
*/

struct KeyVal
{
    QString key;
    QString val;
};

typedef std::function<void(const QString&)> resultHandler;

class Backend : public QObject
{
    Q_OBJECT
    QThread loggerThread;

public:
    Backend();
    ~Backend() {
        loggerThread.quit();
        loggerThread.wait();
    }

    int op(QString cmd, QString arg = "");
    void registerResultHandler(QString key, resultHandler handler) {
        resultHandlerMap[key] = handler;
    }

//TODO???
    QString getConfig(QString key, QString fallback = {});
    void setConfig(QString key, QString val);

private:
    QHash<QString, resultHandler> resultHandlerMap;

signals:
    //TODO: void error(QString);

private slots:
    //void handleDone();
    void handleResult(KeyVal kv) {
        resultHandlerMap[kv.key](kv.val);
    }
signals:
    void readLog(QPrivateSignal);
    void logcolour(QColor);
    void log(QString);
    void error(QString);
};

extern Backend *backend;

class ReadLogWorker : public QObject
{
    Q_OBJECT

    QString logline;
    KeyVal readlogline();
    KeyVal readresult(QString r);

//public:
//    explicit ReadLogWorker(QObject *parent = nullptr);

public slots:
    void readLog();
signals:
    void result(KeyVal logresult);
    void opDone();
    void logcolour(QColor);
    void log(QString);
    //void error(QString);
};

#endif // BACKEND_H
