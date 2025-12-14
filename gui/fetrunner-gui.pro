#lessThan(QT_MAJOR_VERSION, 6) {
#        error(Qt version $$QT_VERSION is not supported. The minimum supported Qt version is 6.9.0.)
#}

#equals(QT_MAJOR_VERSION, 6) {
#        lessThan(QT_MINOR_VERSION, 9){
#                error(Qt version $$QT_VERSION is not supported. The minimum supported Qt version is 6.9.0.)
#        }
#}

QT += widgets

CONFIG += c++17

# You can make your code fail to compile if it uses deprecated APIs.
# In order to do so, uncomment the following line.
#DEFINES += QT_DISABLE_DEPRECATED_BEFORE=0x051500    # disables all the APIs deprecated before Qt 5.15.0
# ... or ... ?
#DEFINES += QT_DISABLE_DEPRECATED_UP_TO=0x051500

SOURCES += \
    main.cpp \
    mainwindow.cpp \
    backend.cpp \
    threadrun.cpp

HEADERS += \
    mainwindow.h \
    backend.h \
    threadrun.h \
    progress_delegate.h

FORMS += \
    mainwindow.ui

TRANSLATIONS += \
    fetrunner-gui_de.ts \
    fetrunner-gui_en.ts
CONFIG += lrelease
CONFIG += embed_translations

# Default rules for deployment.
qnx: target.path = /tmp/$${TARGET}/bin
else: unix:!android: target.path = /opt/$${TARGET}/bin
!isEmpty(target.path): INSTALLS += target

LIBS += -L$$PWD/../libfetrunner/ -lfetrunner

INCLUDEPATH += $$PWD/../libfetrunner
DEPENDPATH += $$PWD/../libfetrunner
