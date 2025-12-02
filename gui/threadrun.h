#ifndef THREADRUN_H
#define THREADRUN_H

#include <QObject>
#include <QThread>
#include <qdebug.h>

class RunThreadWorker : public QObject
{
    Q_OBJECT

public:
    //TODO-- this is just for testing
    ~RunThreadWorker() { qDebug() << "Delete RunThreadWorker"; }

    bool stopFlag;

public slots:
    void ttrun(const QString &parameter);

signals:
    void runThreadWorkerDone(const QString &result);
    void tickTime(const QString &result);
};

class RunThreadController : public QObject
{
    Q_OBJECT

    QThread runThread;
    RunThreadWorker *runThreadWorker{nullptr};

public:
    //RunThreadController();
    ~RunThreadController()
    {
        runThread.quit();
        runThread.wait();
    }

    void runTtThread();

public slots:
    void handleResults(const QString &);
    void stopThread();

signals:
    void startTtRun(const QString &);
    void elapsedTime(const QString);
};

#endif // THREADRUN_H
