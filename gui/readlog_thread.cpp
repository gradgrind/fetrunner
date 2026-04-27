#include "readlog_thread.h"
#include "backend.h"

void ReadLogWorker::readLogRun()
{
    /* ... here is the operation producing a lot of result lines ... */

    while (true) {
       auto kv = backend->readlogline();
        if (kv.key == "$") {
            emit resultLine(backend->readresult(kv.val));
        } else if (kv.key == "---") {
            break;
        }
    }
    emit readLogWorkerDone();
}

void ReadLogThreadController::runReadLogThread(QString cmd)
{
    // `cmd` should be an operation which returns "***" before
    // sending the actual results.
    backend->op(cmd);
    // The back-end should now be running the command.
    results.clear();
    if (!readLogWorker) {
        readLogWorker = new ReadLogWorker;
        readLogWorker->moveToThread(&readLogThread);
        connect( //
            &readLogThread,
            &QThread::finished,
            readLogWorker,
            &QObject::deleteLater);
        connect( //
            this,
            &ReadLogThreadController::startReadLogThread,
            readLogWorker,
            &ReadLogWorker::readLogRun);
        connect( //
            readLogWorker,
            &ReadLogWorker::readLogWorkerDone,
            this,
            &ReadLogThreadController::readLogWorkerDone);
        connect( //
            readLogWorker,
            &ReadLogWorker::resultLine,
            this,
            &ReadLogThreadController::addResult);
        readLogThread.start();
    }
    emit startReadLogThread();
}

void ReadLogThreadController::addResult(KeyVal kv)
{
    results.append(kv);
}