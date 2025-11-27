#ifndef TTRUN_H
#define TTRUN_H

#include <QObject>
#include <QThread>
#include <qdebug.h>

class TtRunWorker : public QObject
{
    Q_OBJECT

    int stopstate{0};

public:
    //TODO-- this is just for testing
    ~TtRunWorker() { qDebug() << "Delete TtRunWorker"; }

public slots:
    void doWork(const QString &parameter);
    void tick();
    void stop();

signals:
    void resultReady(const QString &result);
};

class TtRun : public QObject
{
    Q_OBJECT

    QThread workerThread;
    TtRunWorker *worker;

public:
    //TtRun();
    ~TtRun()
    {
        workerThread.quit();
        workerThread.wait();
    }

    void run();

public slots:
    void handleResults(const QString &);

signals:
    void operate(const QString &);
};

#endif // TTRUN_H
