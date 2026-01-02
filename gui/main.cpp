#include "backend.h"
#include "mainwindow.h"

#include <QApplication>
#include <QLibraryInfo>
#include <QLocale>
#include <QTranslator>

// Makes visible only the literal operators declared in StringLiterals
//using namespace Qt::StringLiterals;

#if QT_VERSION < 0x060000
#define QtLIBLOC QLibraryInfo::location
#else
#define QtLIBLOC QLibraryInfo::path
#endif

int main(
    int argc, char *argv[])
{
    QApplication app(argc, argv);

    auto locale = QLocale::system();
    /*
    QTranslator translator0;
    if (translator0.load(locale,
                         "qtbase",
                         "_",
                         QtLIBLOC(QLibraryInfo::TranslationsPath))) {
        app.installTranslator(&translator0);
    }
    */

    QTranslator translator;
    const QString baseName = "fetrunner-gui_" + locale.name();
    if (translator.load(":/i18n/" + baseName)) {
        app.installTranslator(&translator);
    }

    MainWindow w;

    w.show();
    return app.exec();
}
