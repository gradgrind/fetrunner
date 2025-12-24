cmake_minimum_required(VERSION 3.19)
project(fetrunner-gui LANGUAGES CXX)

# Currently this file uses commands and options which are only
# available from Qt6.7 – in connection with the translations.
find_package(Qt6 6.7 REQUIRED COMPONENTS Core Widgets)
find_package(Qt6 REQUIRED COMPONENTS LinguistTools)

qt_standard_project_setup(I18N_TRANSLATED_LANGUAGES de) # or these:
#set(CMAKE_AUTOMOC ON)
#set(CMAKE_AUTORCC ON)
#set(CMAKE_AUTOUIC ON)

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
    #progress_delegate.h
    #progress_delegate.cpp
    ttgen_tables.cpp
)

qt_add_translations(fetrunner-gui #MERGE_QT_TRANSLATIONS
    TS_FILE_DIR translations
    # Do not embed generated .qm files.
    #QM_FILES_OUTPUT_VARIABLE qm_files
)

# Install generated .qm files.
#install(FILES ${qm_files} DESTINATION translations)

target_link_libraries(fetrunner-gui
    PRIVATE
        Qt::Core
        Qt::Widgets
        libfetrunner
)

include(GNUInstallDirs)

install(TARGETS fetrunner-gui
    BUNDLE  DESTINATION .
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
    LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
)
