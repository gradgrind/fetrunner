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
    const QByteArray bytes{rcs};
    const QList<QByteArray> rsplit{bytes.split('\xff')};
    QStringList ssplit;
    for (const auto &b : rsplit) {
        ssplit.append(QString::fromUtf8(b));
    }
    return ssplit.join("\n");
}
