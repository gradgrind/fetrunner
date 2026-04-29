#ifndef READLOG_THREAD_H
#define READLOG_THREAD_H

#include "backend.h"
#include <QObject>
#include <QThread>
#include <qdebug.h>

// Use a thread to read logged lines.

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
    // Destruction of `readLogWorker` should be done by the
    // Q_OBJECT mechanisms.

public:
    //ReadLogThreadController();
    ~ReadLogThreadController()
    {
        readLogThread.quit();
        readLogThread.wait();
    }

    void runReadLogThread(QString op);

    QList<KeyVal> results;

public slots:
    void addResult(KeyVal);
    void readLogWorkerDone() { qDebug() << "SLOT readLogWorkerDone"; }

signals:
    void startReadLogThread();
    //void readLogWorkerDone();
};

#endif // READLOG_THREAD_H
