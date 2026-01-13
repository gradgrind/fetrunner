#include "threadrun.h"
#include <QTimer>
#include "backend.h"

void RunThreadWorker::ttrun()
{
    //QTimer *timer = new QTimer(this);
    //connect(timer, &QTimer::timeout, this, &TtRunWorker::tick);
    //timer->start(1000);

    stopFlag = false;
    bool stopped = false; // set this to stop (further) stop commands

    /* ... here is the long-running operation ... */

    bool done = false;
    while (!done) {
        //qDebug() << "§poll" << i;
        if (stopFlag && !stopped) {
            backend->op("_STOP_TT");
            stopped = true;
        }
        const auto kvlist = backend->op("_POLL_TT");
        for (const auto &kv : kvlist) {
            //qDebug() << kv.key << kv.val;
            if (kv.key == ".TICK") {
                if (kv.val == "-1") {
                    done = true;
                    emit ticker("");
                } else {
                    //qDebug() << "???" << kv.val;
                    emit ticker(kv.val);
                }
            } else if (kv.key == ".NCONSTRAINTS") {
                emit nconstraints(kv.val);
            } else if (kv.key == ".PROGRESS") {
                emit iprogress(kv.val);
            } else if (kv.key == ".START") {
                emit istart(kv.val);
            } else if (kv.key == ".END") {
                emit iend(kv.val);
            } else if (kv.key == ".ACCEPT") {
                emit iaccept(kv.val);
            } else if (kv.key == ".ELIMINATE") {
                emit ieliminate(kv.val);
            }
        }
    }
    emit runThreadWorkerDone();
}

void RunThreadController::runTtThread()
{
    auto kv = backend->op("RUN_TT");
    // The back-end should now be running the timetable generation.
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
            &RunThreadWorker::iprogress,
            this,
            &RunThreadController::iprogress);
        connect( //
            runThreadWorker,
            &RunThreadWorker::istart,
            this,
            &RunThreadController::istart);
        connect( //
            runThreadWorker,
            &RunThreadWorker::iend,
            this,
            &RunThreadController::iend);
        connect( //
            runThreadWorker,
            &RunThreadWorker::iaccept,
            this,
            &RunThreadController::iaccept);
        connect( //
            runThreadWorker,
            &RunThreadWorker::ieliminate,
            this,
            &RunThreadController::ieliminate);
        runThread.start();
    }
    emit startTtRun("GO");
}

void RunThreadController::stopThread()
{
    //qDebug() << "!!!STOP!!!";
    runThreadWorker->stopFlag = true;
}
