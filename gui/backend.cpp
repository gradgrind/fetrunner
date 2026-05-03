#include "backend.h"
#include <QMap>
//#include <qdebug.h>
#include "../libfetrunner/libfetrunner.h"
#include "globals.h"
//#include <iostream>

Backend *backend;

// display colours for the log
QMap<QString, QColor> colours{{"*INFO*", "#009000"},
                              {"*WARNING*", "#eb8900"},
                              {"*ERROR*", "#d00000"},
                              {"+++", "#000000"},
                              {"---", "#000000"},
                              {"$", "#53a0ff"}};

void ReadLogWorker::readLog() {
    qDebug() << "ReadLogWorker::readLog()" << QThread::currentThreadId();
    while (true) {
       auto kv = readlogline();
       qDebug() << "+" << kv.key << kv.val;
        if (kv.key == "$") {
            emit result(readresult(kv.val));
            continue;
        }
        if (kv.key == "---") {
            emit opDone();
            continue;
        }
        if (kv.key == "*ERROR*") {
            notifier->emit errorPopup(kv.val);
            continue;
        }
        if (kv.key == "-*-*-") {
            break;
        }
    }
    //TODO: Emit a signal?
}

//ReadLogWorker::ReadLogWorker(QObject *parent) : QObject(parent) {}

Backend::Backend() : QObject() {
    qDebug() << "Backend::Backend()" << QThread::currentThreadId();

    ReadLogWorker *worker = new ReadLogWorker;
    worker->moveToThread(&loggerThread);
    connect(&loggerThread, &QThread::started, worker, &ReadLogWorker::readLog);
    connect(&loggerThread, &QThread::finished, worker, &QObject::deleteLater);
    //connect(worker, &ReadLogWorker::opDone, this, &ReadLogController::handleDone);
    connect(worker, &ReadLogWorker::result, this, &Backend::handleResult);
    connect(worker, &ReadLogWorker::logcolour, this, &Backend::logcolour);
    connect(worker, &ReadLogWorker::log, this, &Backend::log);
    //connect(worker, &ReadLogWorker::error, this, &Backend::error);
    loggerThread.start();
}

int Backend::op(QString cmd, QString arg)
{
    if (!arg.isEmpty()) {
        cmd += " " + arg;
    }
    qDebug() << "?" << cmd << QThread::currentThreadId();

    auto res = FetRunnerCommand(cmd.toUtf8().data());
    qDebug() << "?DONE";
    return res;
}

KeyVal ReadLogWorker::readresult(QString r)
{
    auto n = r.indexOf('=');
    if (n < 0)
        return KeyVal{"", QString{"BUG in backend result: "} + logline};
    auto rkey = r.left(n);
    auto rval = r.right(r.length() - n - 1);
    return KeyVal{rkey, rval};
}

KeyVal ReadLogWorker::readlogline()
{
    while (true) {
        logline = QString(FetRunnerReadLog());
        //qDebug() << "=" << logline;
        if (logline.length() != 0 && logline.at(0) != " ")
            break;
        emit log(logline); // write to log without change of colour
    }
    auto i = logline.indexOf(" ");
    QString msgtype, msgrest;
    if (i < 0) {
        // there is only the type
        msgtype = logline;
        msgrest = "";
    } else {
        // split into message-type and rest
        msgtype = logline.left(i);
        msgrest = logline.right(logline.length() - i - 1);
    }
    // The type determines the display colour.
    emit logcolour(colours.value(msgtype, QColor{0x76, 0x5e, 0xff}));
    emit log(logline.replace("||", "\n + ")); // write to log
    return KeyVal{msgtype, msgrest};
}
