#include "backend.h"
#include <QThread>

class ReadLogWorker : public QObject
{
    Q_OBJECT
    bool stopFlag;
    void readLogs()
    {
        while (true) {
           auto kv = backend->readlogline();
           //qDebug() << "+" << kv.key << kv.val;
            if (kv.key == "$") {
                emit result(backend->readresult(kv.val));
            } else if (kv.key == "---") {
                break;
            }
        }
    }
    void readTtRunLogs()
    {
        bool stopped = false;
        stopFlag = false;
        while (true) {
            if (stopFlag && !stopped) {
                backend->op("_STOP_TT");
                stopped = true;
            }
            auto kv = backend->readlogline();
            //qDebug() << kv.key << kv.val;
            if (kv.key == "$") {
                auto kvr = backend->readresult(kv.val);
                if (kvr.key == ".TICK") {
                    //qDebug() << "???" << kvr.val;
                    emit ticker(kvr.val);
                } else if (kvr.key == ".NCONSTRAINTS") {
                    emit nconstraints(kvr.val);
                } else if (kvr.key == ".PROGRESS") {
                    emit iprogress(kvr.val);
                } else if (kvr.key == ".START") {
                    emit istart(kvr.val);
                } else if (kvr.key == ".END") {
                    emit iend(kvr.val);
                } else if (kvr.key == ".ACCEPT") {
                    emit iaccept(kvr.val);
                } else if (kvr.key == ".ELIMINATE") {
                    emit ieliminate(kvr.val);
                }
            } else if (kv.key == "---") {
                emit ticker("");
                break;
            }
        }

    }
public slots:
    void readLog() {
        readLogs();
        emit opDone();
    }
    void cancelRun() {
        stopFlag = true;
    }
signals:
    void result(KeyVal logresult);
    void opDone();
};

class ReadLogController : public QObject
{
    Q_OBJECT
    QThread workerThread;
public:
    ReadLogController() {
        ReadLogWorker *worker = new ReadLogWorker;
        worker->moveToThread(&workerThread);
        connect(&workerThread, &QThread::finished, worker, &QObject::deleteLater);
        connect(this, &ReadLogController::readLog, worker, &ReadLogWorker::readLog);
        connect(this, &ReadLogController::cancelRun, worker, &ReadLogWorker::cancelRun);
        connect(worker, &ReadLogWorker::opDone, this, &ReadLogController::handleDone);
        workerThread.start();
    }
    ~ReadLogController() {
        workerThread.quit();
        workerThread.wait();
    }
public slots:
    void handleDone();
signals:
    void readLog();
    void cancelRun();
};

ReadLogController readLogController;

void startReadLog()
{
    emit readLogController.readLog();
}

void stopRun()
{
    emit readLogController.cancelRun();
}
