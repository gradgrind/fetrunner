cmake_minimum_required(VERSION 3.19)
project(fetrunner-gui LANGUAGES CXX)

# Currently this file uses commands and options which are only
# available from Qt6.7 – in connection with the translations.
# MERGE_QT_TRANSLATIONS even requires Qt6.9.
# The fetrunner code itself should be backward compatible to at least
# Qt6.2, perhaps even Qt5.15.
find_package(Qt6 6.9 REQUIRED COMPONENTS Gui Widgets)
find_package(Qt6 REQUIRED COMPONENTS LinguistTools)

# Qt >= 6.3. I18N_TRANSLATED_LANGUAGES: Qt >= 6.7
qt_standard_project_setup(I18N_TRANSLATED_LANGUAGES de)

add_library(libfetrunner STATIC IMPORTED) # STATIC or SHARED
set_target_properties(libfetrunner PROPERTIES
  IMPORTED_LOCATION "${CMAKE_CURRENT_SOURCE_DIR}/../libfetrunner/libfetrunner.a"
  INTERFACE_INCLUDE_DIRECTORIES "${CMAKE_CURRENT_SOURCE_DIR}/../libfetrunner"
)

qt_add_executable(fetrunner-gui
    WIN32 MACOSX_BUNDLE
    main.cpp
    mainwindow.cpp
    mainwindow.h
    mainwindow.ui
    backend.h backend.cpp
    threadrun.h threadrun.cpp
    ttgen_tables.cpp
)

#set_target_properties(fetrunner-gui PROPERTIES AUTOMOC TRUE)
#set_target_properties(fetrunner-gui PROPERTIES AUTORCC TRUE)
#set_target_properties(fetrunner-gui PROPERTIES AUTOUIC TRUE)

qt_add_translations(fetrunner-gui MERGE_QT_TRANSLATIONS
    TS_FILE_DIR translations
    # Do not embed generated .qm files.
    #QM_FILES_OUTPUT_VARIABLE qm_files
)

target_link_libraries(fetrunner-gui
    PRIVATE
        Qt::Core
        Qt::Widgets
        libfetrunner
)

# Install generated .qm files.
#install(FILES ${qm_files} DESTINATION translations)

install(TARGETS fetrunner-gui
    BUNDLE  DESTINATION .
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
    #LIBRARY DESTINATION ${CMAKE_INSTALL_LIBDIR}
)

# Qt >= 6.3.
qt_generate_deploy_app_script(
    TARGET fetrunner-gui
    OUTPUT_SCRIPT deploy_script
    NO_TRANSLATIONS
    NO_UNSUPPORTED_PLATFORM_ERROR
    INCLUDE_PLUGINS wayland xcb
)
install(SCRIPT ${deploy_script})
