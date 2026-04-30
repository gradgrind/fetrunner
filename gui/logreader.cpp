#include "backend.h"
#include <QThread>
#include <QHash>

class LogReader
{

};

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
    void readLog(LogReader *logReader) {
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
    //TODO ...??? Pass a LogReader object to readLog?
    QList<KeyVal> results;
    emit readLogController.readLog(results);

    //TODO: call the actual function

    return readLogController.results;
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

The dispatcher (backend::op) can run in the main thread, so it can
call GUI functions directly. Signals from the logger can be caught
by the Controller, which is also in the main thread. The one special
case so far is the "RUN_TT" command, which is lengthy. Perhaps this
can be done by a special command ("!RUN_TT") , or argument, which
makes it run in a separate goroutine? This would be a further category
for Dispatch, as it should log the start, but not the end, before
starting the goroutine for the actual action and returning immediately.
*/

void testfun(QString val)
{
    // TODO
}

QHash<QString, std::function<void(QString)>> resultHandlerMap{
    {"Test", testfun}
};

void ReadLogController::handleResult(KeyVal kv)
{
    //TODO: perhaps a default which just accumulates the kv?
    // But then there is still the question of how/when to read this ...

    // Perhaps the possibility of placing the results in a container?
    // Presumable a QList to maintain the order?
    resultHandlerMap[kv.key](kv.val);
}
