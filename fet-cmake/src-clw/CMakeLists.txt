project(fet-clw)

find_package(Qt6 6.3 REQUIRED COMPONENTS Core)

# Qt >= 6.3.
qt_standard_project_setup()

qt_add_executable(fet-clw
    ../src/engine/activity.cpp ../src/engine/activity.h
    ../src/engine/activitytag.cpp ../src/engine/activitytag.h
    ../src/engine/building.cpp ../src/engine/building.h
    ../src/engine/export.cpp ../src/engine/export.h
    ../src/engine/generate.cpp ../src/engine/generate.h
    ../src/engine/generate_pre.cpp ../src/engine/generate_pre.h
    ../src/engine/lockunlock.cpp ../src/engine/lockunlock.h
    ../src/engine/matrix.h
    ../src/engine/room.cpp ../src/engine/room.h
    ../src/engine/rules.cpp ../src/engine/rules.h
    ../src/engine/solution.cpp ../src/engine/solution.h
    ../src/engine/spaceconstraint.cpp ../src/engine/spaceconstraint.h
    ../src/engine/studentsset.cpp ../src/engine/studentsset.h
    ../src/engine/subject.cpp ../src/engine/subject.h
    ../src/engine/teacher.cpp ../src/engine/teacher.h
    ../src/engine/timeconstraint.cpp ../src/engine/timeconstraint.h
    ../src/engine/timetable.cpp ../src/engine/timetable.h
    ../src/engine/timetable_defs.cpp ../src/engine/timetable_defs.h
    ../src/engine/timetableexport.cpp ../src/engine/timetableexport.h
    ../src/interface/fet.cpp ../src/interface/fet.h
    ../src/interface/fetmainform.cpp
    ../src/interface/longtextmessagebox.cpp ../src/interface/longtextmessagebox.h
    ../src/interface/messageboxes.cpp ../src/interface/messageboxes.h
    ../src/interface/utilities.cpp ../src/interface/utilities.h
)
target_include_directories(fet-clw PRIVATE
    ../src/engine
    ../src/interface
)
#set_target_properties(fet-clw PROPERTIES AUTOMOC TRUE)
set_property(SOURCE ../src/interface/fetmainform.h PROPERTY SKIP_AUTOGEN ON)
set_target_properties(fet-clw PROPERTIES WIN32_EXECUTABLE TRUE)

target_compile_definitions(fet-clw PRIVATE
    FET_COMMAND_LINE
    QT_NO_FOREACH
    QT_NO_JAVA_STYLE_ITERATORS
    QT_NO_LINKED_LIST
    QT_STRICT_ITERATORS
)

target_link_libraries(fet-clw PRIVATE
    Qt::Core
)

install(TARGETS fet-clw
    BUNDLE DESTINATION .
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
)

# Qt >= 6.3.
qt_generate_deploy_app_script(
    TARGET fet-clw
    OUTPUT_SCRIPT deploy_script
    NO_TRANSLATIONS
    NO_UNSUPPORTED_PLATFORM_ERROR
)
install(SCRIPT ${deploy_script})
