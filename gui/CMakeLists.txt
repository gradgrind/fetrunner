cmake_minimum_required(VERSION 3.19)
project(fetrunner-gui LANGUAGES CXX)

find_package(Qt6 6.2 REQUIRED COMPONENTS Core Widgets)
find_package(Qt6 REQUIRED COMPONENTS LinguistTools)
find_package(Qt6 REQUIRED COMPONENTS Core)
find_package(Qt6 REQUIRED COMPONENTS Core)

#qt_standard_project_setup(I18N_TRANSLATED_LANGUAGES de) # or these:
set(CMAKE_AUTOMOC ON)
set(CMAKE_AUTORCC ON)
set(CMAKE_AUTOUIC ON)

add_library(libfetrunner STATIC IMPORTED) # STATIC or SHARED
set_target_properties(libfetrunner PROPERTIES
  IMPORTED_LOCATION "${CMAKE_SOURCE_DIR}/../libfetrunner/libfetrunner.a"
  INTERFACE_INCLUDE_DIRECTORIES "${CMAKE_SOURCE_DIR}/../libfetrunner"
)

qt_add_executable(fetrunner-gui
    WIN32 MACOSX_BUNDLE
    main.cpp
    mainwindow.cpp
    mainwindow.h
    mainwindow.ui
    backend.h backend.cpp
    threadrun.h threadrun.cpp
    progress_delegate.h
    progress_delegate.cpp
    progress_table.cpp

)

qt_add_translations(fetrunner-gui MERGE_QT_TRANSLATIONS)

target_link_libraries(fetrunner-gui
    PRIVATE
        Qt::Core
        Qt::Widgets
        libfetrunner
)
target_link_libraries(fetrunner-gui PRIVATE Qt6::Core)
target_link_libraries(fetrunner-gui PRIVATE Qt6::Core)

include(GNUInstallDirs)

install(TARGETS fetrunner-gui
    BUNDLE  DESTINATION .
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
)
