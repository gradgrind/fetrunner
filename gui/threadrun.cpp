#include "threadrun.h"
#include "backend.h"

void RunThreadWorker::ttrun()
{
    stopFlag = false;
    bool stopped = false; // set this to stop (further) stop commands

    /* ... here is the long-running operation ... */

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
    emit runThreadWorkerDone();
}

void RunThreadController::runTtThread()
{
    backend->op("RUN_TT");
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
    emit startTtRun();
}

void RunThreadController::stopThread()
{
    //qDebug() << "!!!STOP!!!";
    runThreadWorker->stopFlag = true;
}
