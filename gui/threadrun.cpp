#include "threadrun.h"
#include <QTimer>
#include "backend.h"

void RunThreadWorker::ttrun(const QString &parameter)
{
    //QTimer *timer = new QTimer(this);
    //connect(timer, &QTimer::timeout, this, &TtRunWorker::tick);
    //timer->start(1000);

    QString result{"Done ..."};

    stopFlag = false;
    bool stopped = false; // set this to stop (further) stop commands

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
    for (int i = 0; i < 25; ++i) {
        //qDebug() << "§poll" << i;
        if (stopFlag && !stopped) {
            backend->op("_STOP_TT");
            //const auto kvlist = backend->op("_STOP_TT");
            //for (const auto &kv : kvlist) {
            //    qDebug() << "?STOP?" << kv.key << kv.val;
            //}
            stopped = true;
        }
        const auto kvlist = backend->op("_POLL_TT");
        for (const auto &kv : kvlist) {
            //qDebug() << kv.key << kv.val;
            if (kv.key == ".TICK") {
                if (kv.val == "-1") {
                    // result = kv.val;
                    done = true;
                } else {
                    //qDebug() << "???" << kv.val;
                    emit ticker(kv.val);
                }
            } else if (kv.key == ".NCONSTRAINTS") {
                auto items = kv.val.split(u'.');
                emit nconstraints(kv.val);
            } else if (kv.key == ".PROGRESS") {
                emit progress(kv.val);
            }
        }
        //qDebug() << "§loop-end" << done;
        if (done)
            break;
        //QThread::msleep(500);
    }
    emit runThreadWorkerDone(result);
    //thread()->quit();
}

/* TODO--
void TtRunWorker::tick()
{
    bool done = false;
    for (const auto &kv : backend->op("_POLL_TT")) {
        qDebug() << kv.key << kv.val;
        if (kv.key == "TT_DONE") {
            // result = kv.val;
            done = true;
        }
    }
    if (done) {
        //emit resultReady(result);
        thread()->quit();
    }
}
*/

//TODO
void RunThreadController::runTtThread()
{
    auto kv = backend->op("RUN_TT");
    if (kv.length() == 0)
        return; // return if start unsuccessful

    //qDebug() << "Start ...";

    // The back-end should now be running the timetable generation.
    //TODO: Adjust anything that needs to reflect this ...

    if (!runThreadWorker) {
        runThreadWorker = new RunThreadWorker;
        runThreadWorker->moveToThread(&runThread);
        connect( //
            &runThread,
            &QThread::finished,
            runThreadWorker,
            &QObject::deleteLater);
        connect( //
            this,
            &RunThreadController::startTtRun,
            runThreadWorker,
            &RunThreadWorker::ttrun);
        connect( //
            runThreadWorker,
            &RunThreadWorker::runThreadWorkerDone,
            this,
            &RunThreadController::runThreadWorkerDone);
        connect( //
            runThreadWorker,
            &RunThreadWorker::ticker,
            this,
            &RunThreadController::ticker);
        connect( //
            runThreadWorker,
            &RunThreadWorker::nconstraints,
            this,
            &RunThreadController::nconstraints);
        connect( //
            runThreadWorker,
            &RunThreadWorker::progress,
            this,
            &RunThreadController::progress);
        runThread.start();
    }
    emit startTtRun("GO");
}

void RunThreadController::stopThread()
{
    qDebug() << "!!!STOP!!!";
    runThreadWorker->stopFlag = true;
}
