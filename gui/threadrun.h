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
    void ticker(const QString &result);
    void nconstraints(const QString &result);
    void istart(const QString &result);
    void progress(const QString &result);
    void iend(const QString &result);
    void iaccept(const QString &result);
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
    void stopThread();

signals:
    void startTtRun(const QString &);
    void ticker(const QString &);
    void runThreadWorkerDone();
    void nconstraints(const QString &result);
    void istart(const QString &result);
    void progress(const QString &result);
    void iend(const QString &result);
    void iaccept(const QString &result);
};

#endif // THREADRUN_H
