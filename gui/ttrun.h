#ifndef TTRUN_H
#define TTRUN_H

#include <QObject>
#include <QThread>

class TtRunWorker : public QObject
{
    Q_OBJECT

    int stopstate{0};

public slots:
    void doWork(const QString &parameter);
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
    TtRun();
    ~TtRun()
    {
        workerThread.quit();
        workerThread.wait();
    }

public slots:
    //void handleResults(const QString &);

signals:
    void operate(const QString &);
};

#endif // TTRUN_H
