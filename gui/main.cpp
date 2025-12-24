#include "mainwindow.h"

#include <QApplication>
#include <QLibraryInfo>
#include <QLocale>
#include <QTranslator>

// Makes visible only the literal operators declared in StringLiterals
using namespace Qt::StringLiterals;

int main(
    int argc, char *argv[])
{
    QApplication app(argc, argv);
    QTranslator translator0;

    if (translator0.load(QLocale::system(),
                         u"qtbase"_s,
                         u"_"_s,
                         QLibraryInfo::path(QLibraryInfo::TranslationsPath))) {
        app.installTranslator(&translator0);
    }

    QTranslator translator;
    const QStringList uiLanguages = QLocale::system().uiLanguages();
    //qDebug() << "???" << uiLanguages;
    for (const QString &locale : uiLanguages) {
        //qDebug() << "locale:" << locale;
        const QString baseName = "fetrunner-gui_" + QLocale(locale).name();
        //qDebug() << "baseName:" << baseName;
        if (translator.load(":/i18n/" + baseName)) {
            app.installTranslator(&translator);
            break;
        }
    }

    MainWindow w;

    w.show();
    return app.exec();
}
