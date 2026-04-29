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
    void readLog(QList<KeyVal> &results) {
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
        connect(worker, &ReadLogWorker::result, this, &ReadLogController::handleResult);
        workerThread.start();
    }
    ~ReadLogController() {
        workerThread.quit();
        workerThread.wait();
    }
    QList<KeyVal> results;
private slots:
    void handleDone();
    void handleResult(KeyVal kv);
signals:
    void readLog(QList<KeyVal> &results);
    void cancelRun();
};

static ReadLogController readLogController;

void stopRun()
{
    emit readLogController.cancelRun();
}

QList<KeyVal> command()
{
    //TODO ...???
    QList<KeyVal> results;
    emit readLogController.readLog(results);

    //TODO: call the actual function

    return readLogController.results
}
/*
A back-end call can (should?) run synchronously if it is short,
allowing results to be returned directly. To allow the reading of
the logs while they are being generated (to avoid buffer blocking),
the log reader must be started first:

    startReadLog();
    doCommand();
    waitForEndOfReadLog(); // Maybe not ... it's not easy
    return results;

The fetrunner command, however, takes a long time to run, but must
return immediately so as not to block the GUI. Everything else is
managed via its signals. So no explicit waiting is called for here.
*/
