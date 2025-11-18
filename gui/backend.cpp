#include "backend.h"
#include "../libfetrunner/libfetrunner.h"

Backend::Backend(QObject *parent)
    : QObject{parent}
{}

QString test_backend(QString s)
{
    auto sbytes = s.toUtf8();
    auto cs = const_cast<char *>(sbytes.constData());
    auto rcs = FetRunner(cs);
    return QString::fromUtf8(rcs);
}
