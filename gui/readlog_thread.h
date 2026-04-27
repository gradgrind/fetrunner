#ifndef READLOG_THREAD_H
#define READLOG_THREAD_H

#include "backend.h"
#include <QObject>
#include <QThread>

// Use a thread to read the output of a command with many logged lines,
// to avoid filling, and thus blocking, the back-end buffer.

class ReadLogWorker : public QObject
{
    Q_OBJECT

public slots:
    void readLogRun();

signals:
    void readLogWorkerDone();
    void resultLine(KeyVal);
};

class ReadLogThreadController : public QObject
{
    Q_OBJECT

    QThread readLogThread;
    ReadLogWorker *readLogWorker{nullptr};

public:
    //ReadLogThreadController();
    ~ReadLogThreadController()
    {
        readLogThread.quit();
        readLogThread.wait();
        delete readLogWorker;
    }

    void runReadLogThread(QString op);

    QList<KeyVal> results;

public slots:
    void addResult(KeyVal);

signals:
    void startReadLogThread();
    void readLogWorkerDone();
};

#endif // READLOG_THREAD_H
