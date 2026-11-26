#include "backend.h"
#include "mainwindow.h"

#include <QApplication>
#include <QDir>
#include <QLibraryInfo>
#include <QLocale>
#include <QTranslator>
#include <QtEnvironmentVariables>

int main(
    int argc, char *argv[])
{
    QApplication app(argc, argv);

    // Allow the language for translation to be read from the environment variable "QTLANG".
    auto locale = QLocale(qEnvironmentVariable("QTLANG"));
    if ( locale.language() == QLocale::C ) {
        // No valid language was found.
        locale = QLocale::system();
    }
    QTranslator translator;
    const QString baseName = "fetrunner-gui_" + locale.name();
    if (translator.load(":/i18n/" + baseName)) {
        app.installTranslator(&translator);
    } else {
        qDebug() << "Set language failed:" << baseName;
    }

    //qDebug() << "&&&1" << locale.name();
    //qDebug() << "&&&2" << QLocale::languageToCode(locale.language());

    backend = new Backend();
    MainWindow w;
    w.setWindowIcon(QIcon(":gui/icons/fetrunner.svg"));
noicon:

    w.show();
    return app.exec();
}
