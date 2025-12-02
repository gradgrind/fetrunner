#ifndef THREADRUN_H
#define THREADRUN_H

#include <QObject>
#include <QThread>
#include <qdebug.h>

class TtRunWorker : public QObject
{
    Q_OBJECT

public:
    //TODO-- this is just for testing
    ~TtRunWorker() { qDebug() << "Delete TtRunWorker"; }

    bool stopFlag;

public slots:
    void doWork(const QString &parameter);

signals:
    void resultReady(const QString &result);
    void tickTime(const QString &result);
};

class RunThreadController : public QObject
{
    Q_OBJECT

    QThread runThread;
    TtRunWorker *worker;

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
    void operate(const QString &);
    void elapsedTime(const QString);
};

#endif // THREADRUN_H
