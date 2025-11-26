#include "ttrun.h"
#include "backend.h"

void TtRunWorker::doWork(const QString &parameter)
{
    QString result;
    /* ... here is the expensive or blocking operation ... */

    bool finished = false;
    while (!finished) {
        const auto kvlist = backend->op("_POLL_TT");
        for (const auto &kv : kvlist) {
            if (kv.key == "FINISHED") {
                result = kv.val;
                finished = true;
            }
        }

        if (stopstate > 0) {
            backend->op("_STOP_TT");
            stopstate = -1;
        }
    }

    emit resultReady(result);
}

void TtRunWorker::stop()
{
    if (stopstate == 0) {
        stopstate = 1;
    }
}

TtRun::TtRun()
    : QObject()
{
    auto kv = backend->op("RUN_TT");
    if (kv.length() == 0)
        return; // return if start unsuccessful

    // The back-end should now be running the timetable generation.
    //TODO: Adjust anything that needs to reflect this ...

    worker = new TtRunWorker;
    worker->moveToThread(&workerThread);
    connect(&workerThread, &QThread::finished, worker, &QObject::deleteLater);
    connect(this, &TtRun::operate, worker, &TtRunWorker::doWork);
    //connect(worker, &TtRunWorker::resultReady, this, &TtRun::handleResults);
    workerThread.start();
}
