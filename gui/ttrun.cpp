#include "ttrun.h"
#include "backend.h"

void TtRunWorker::doWork(const QString &parameter)
{
    QString result{"Done ..."};

    /* ... here is the expensive or blocking operation ... */

    //QThread::sleep(5);

    /*
    bool done = false;
    while (!done) {
        const auto kvlist = backend->op("_POLL_TT");
        for (const auto &kv : kvlist) {
            if (kv.key == "FINISHED") {
                // result = kv.val;
                done = true;
            }
        }

        if (stopstate > 0) {
            backend->op("_STOP_TT");
            stopstate = -1;
        }
    }
    */

    bool done = false;
    for (int i = 0; i < 15; ++i) {
        const auto kvlist = backend->op("_POLL_TT");
        for (const auto &kv : kvlist) {
            qDebug() << kv.key << kv.val;
            if (kv.key == "TT_DONE") {
                // result = kv.val;
                done = true;
            }
        }
        if (done)
            break;
        QThread::msleep(500);
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

    qDebug() << "Start ...";

    // The back-end should now be running the timetable generation.
    //TODO: Adjust anything that needs to reflect this ...

    worker = new TtRunWorker;
    worker->moveToThread(&workerThread);
    connect(&workerThread, &QThread::finished, worker, &QObject::deleteLater);
    connect(this, &TtRun::operate, worker, &TtRunWorker::doWork);
    connect(worker, &TtRunWorker::resultReady, this, &TtRun::handleResults);
    workerThread.start();
    emit operate("GO");
}

void TtRun::handleResults(const QString &result)
{
    qDebug() << result;
}
