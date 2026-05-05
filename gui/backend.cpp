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
    //qDebug() << "ReadLogWorker::readLog()" << QThread::currentThreadId();
    while (true) {
       auto ll = QString(FetRunnerReadLog());
       emit newLogLine(ll);
       if (ll == "---") break;
    }
    //TODO: Emit a signal?
}

//ReadLogWorker::ReadLogWorker(QObject *parent) : QObject(parent) {}

Backend::Backend() : QObject() {
    //qDebug() << "Backend::Backend()" << QThread::currentThreadId();

    ReadLogWorker *worker = new ReadLogWorker;
    worker->moveToThread(&loggerThread);

    connect(&loggerThread, &QThread::finished, worker, &QObject::deleteLater);
    connect(worker, &ReadLogWorker::newLogLine, this, &Backend::handleLogLine);
    connect(this, &Backend::readLogInThread, worker, &ReadLogWorker::readLog);

    //connect(&loggerThread, &QThread::started, worker, &ReadLogWorker::readLog);
    //connect(worker, &ReadLogWorker::opDone, this, &ReadLogController::handleDone);
    //connect(worker, &ReadLogWorker::result, this, &Backend::handleResult);
    //connect(worker, &ReadLogWorker::logcolour, this, &Backend::logcolour);
    //connect(worker, &ReadLogWorker::log, this, &Backend::log);
    //connect(worker, &ReadLogWorker::error, this, &Backend::error);
    loggerThread.start();
}

int Backend::op(QString cmd, QString arg)
{
    if (!arg.isEmpty()) {
        cmd += " " + arg;
    }
    //qDebug() << "?" << cmd << QThread::currentThreadId();

    auto res = FetRunnerCommand(cmd.toUtf8().data());
    //qDebug() << "?DONE";

    if (cmd[0] != '_') {
        if (cmd[0] == '!')
            emit readLogInThread(QPrivateSignal());
        else
            readLog();
    }
    return res;
}

void Backend::readLog() {
    //qDebug() << "Backend::readLog()" << QThread::currentThreadId();
    while (true) {
        auto ll = QString(FetRunnerReadLog());
        handleLogLine(ll);
        if (ll == "---") break;
    }
}

void Backend::handleLogLine(QString logline) {
    //qDebug() << "handleLogLine" << logline;
    if (logline.length() == 0 || logline.at(0) == " ") {
        emit log(logline); // write to log without change of colour
        return;
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

    //qDebug() << "&" << msgtype << msgrest;
    if (msgtype == "$") {
        auto n = msgrest.indexOf('=');
        if (n < 0) {
            notifier->emit errorPopup(QString{"BUG in backend result: "} + msgrest);
            return;
        }
        auto rkey = msgrest.left(n);
        auto rval = msgrest.right(msgrest.length() - n - 1);
        //qDebug() << "handleResult" << rkey;
        auto f = resultHandlerMap.value(rkey, nullptr);
        if (f == nullptr)
            emit log("*NO_HANDLER* " + rkey);
        else
            f(rval);
        return;
    }
    if (msgtype == "---") {
        //TODO
        //emit opDone();
        return;
    }
    if (msgtype == "*ERROR*") {
        notifier->emit errorPopup(msgrest);
        return;
    }
    if (msgtype == "-*-*-") {
        //TODO
        return;
    }
}
