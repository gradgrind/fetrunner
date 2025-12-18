QT += widgets

CONFIG += c++17

DESTDIR = ../..
TARGET = fetrunner-gui

OBJECTS_DIR = ../../tmp/commandline
UI_DIR = ../../tmp/commandline
MOC_DIR = ../../tmp/commandline
RCC_DIR = ../../tmp/commandline

SOURCES += \
    main.cpp \
    mainwindow.cpp \
    backend.cpp \
    threadrun.cpp \
    ttgen_tables.cpp


HEADERS += \
    mainwindow.h \
    backend.h \
    threadrun.h

FORMS += \
    mainwindow.ui

TRANSLATIONS += \
    fetrunner-gui_de.ts \
    fetrunner-gui_en.ts
CONFIG += lrelease
CONFIG += embed_translations

LIBS += -L$$PWD/../libfetrunner/ -lfetrunner

INCLUDEPATH += $$PWD/../libfetrunner
DEPENDPATH += $$PWD/../libfetrunner

unix {
    target.path = /usr/bin
    INSTALLS += target
}
