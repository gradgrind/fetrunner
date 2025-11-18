#include "backend.h"
#include "backend/libfetrunner.h"

Backend::Backend(QObject *parent)
    : QObject{parent}
{}

QString test_backend(QString s)
{
    auto sbytes = s.toUtf8();
    auto cs = const_cast<char *>(sbytes.constData());
    auto rcs = Test(cs);
    return QString::fromUtf8(rcs);
}
