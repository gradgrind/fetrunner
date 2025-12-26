QT += widgets

CONFIG += c++17

DESTDIR = ../..
TARGET = fetrunner-gui

OBJECTS_DIR = ../../tmp/fetrunner
UI_DIR = ../../tmp/fetrunner
MOC_DIR = ../../tmp/fetrunner
RCC_DIR = ../../tmp/fetrunner

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

CONFIG -= debug_and_release
CONFIG += warn_on release lrelease embed_translations
TRANSLATIONS += \
    translations/fetrunner-gui_de.ts

LIBS += -L$$PWD/../libfetrunner/ -lfetrunner

INCLUDEPATH += $$PWD/../libfetrunner
DEPENDPATH += $$PWD/../libfetrunner

unix {
    target.path = /usr/bin
    INSTALLS += target
}
