project(fet-gui)

find_package(Qt6 6.3 REQUIRED COMPONENTS Core Gui Network Widgets)
#find_package(Qt6 REQUIRED COMPONENTS LinguistTools)
find_package(Qt6 OPTIONAL_COMPONENTS PrintSupport)

# Qt >= 6.3. I18N_TRANSLATED_LANGUAGES: Qt >= 6.7
#qt_standard_project_setup(I18N_TRANSLATED_LANGUAGES de fr)
qt_standard_project_setup()

qt_add_executable(fet WIN32 MACOSX_BUNDLE
    engine/activity.cpp engine/activity.h
    engine/activitytag.cpp engine/activitytag.h
    engine/building.cpp engine/building.h
    engine/export.cpp engine/export.h
    engine/generate.cpp engine/generate.h
    engine/generate_pre.cpp engine/generate_pre.h
    engine/import.cpp engine/import.h
    engine/lockunlock.cpp engine/lockunlock.h
    engine/matrix.h
    engine/room.cpp engine/room.h
    engine/rules.cpp engine/rules.h
    engine/solution.cpp engine/solution.h
    engine/spaceconstraint.cpp engine/spaceconstraint.h
    engine/statisticsexport.cpp engine/statisticsexport.h
    engine/studentsset.cpp engine/studentsset.h
    engine/subject.cpp engine/subject.h
    engine/teacher.cpp engine/teacher.h
    engine/timeconstraint.cpp engine/timeconstraint.h
    engine/timetable.cpp engine/timetable.h
    engine/timetable_defs.cpp engine/timetable_defs.h
    engine/timetableexport.cpp engine/timetableexport.h
    interface/activitiesform.cpp interface/activitiesform.h
    interface/activitiesform_template.ui
    interface/activitiesroomsstatisticsform.cpp interface/activitiesroomsstatisticsform.h
    interface/activitiesroomsstatisticsform_template.ui
    interface/activitiestagsform.cpp interface/activitiestagsform.h
    interface/activitiestagsform_template.ui
    interface/activityplanningconfirmationform.cpp interface/activityplanningconfirmationform.h
    interface/activityplanningconfirmationform_template.ui
    interface/activityplanningform.cpp interface/activityplanningform.h
    interface/activitytagsform.cpp interface/activitytagsform.h
    interface/activitytagsform_template.ui
    interface/addactivityform.cpp interface/addactivityform.h
    interface/addactivityform_template.ui
    interface/addbuildingform.cpp interface/addbuildingform.h
    interface/addbuildingform_template.ui
    interface/addexistingstudentsgroupsform.cpp interface/addexistingstudentsgroupsform.h
    interface/addexistingstudentsgroupsform_template.ui
    interface/addexistingstudentssubgroupsform.cpp interface/addexistingstudentssubgroupsform.h
    interface/addexistingstudentssubgroupsform_template.ui
    interface/addgroupactivitiesininitialorderitemform.cpp interface/addgroupactivitiesininitialorderitemform.h
    interface/addgroupactivitiesininitialorderitemform_template.ui
    interface/addormodifyspaceconstraint.cpp interface/addormodifyspaceconstraint.h
    interface/addormodifytimeconstraint.cpp interface/addormodifytimeconstraint.h
    interface/addremovemultipleconstraintsactivitiesoccupymaxdifferentroomsform.cpp interface/addremovemultipleconstraintsactivitiesoccupymaxdifferentroomsform.h
    interface/addremovemultipleconstraintsactivitiesoccupymaxdifferentroomsform_template.ui
    interface/addremovemultipleconstraintsactivitiessamestartinghourform.cpp interface/addremovemultipleconstraintsactivitiessamestartinghourform.h
    interface/addremovemultipleconstraintsactivitiessamestartinghourform_template.ui
    interface/addroomform.cpp interface/addroomform.h
    interface/addroomform_template.ui
    interface/addstudentsgroupform.cpp interface/addstudentsgroupform.h
    interface/addstudentsgroupform_template.ui
    interface/addstudentssubgroupform.cpp interface/addstudentssubgroupform.h
    interface/addstudentssubgroupform_template.ui
    interface/addstudentsyearform.cpp interface/addstudentsyearform.h
    interface/addstudentsyearform_template.ui
    interface/addteacherform.cpp interface/addteacherform.h
    interface/addteacherform_template.ui
    interface/advancedfilterform.cpp interface/advancedfilterform.h
    interface/advancedlockunlockform.cpp interface/advancedlockunlockform.h
    interface/allspaceconstraintsform.cpp interface/allspaceconstraintsform.h
    interface/allspaceconstraintsform_template.ui
    interface/alltimeconstraintsform.cpp interface/alltimeconstraintsform.h
    interface/alltimeconstraintsform_template.ui
    interface/buildingsform.cpp interface/buildingsform.h
    interface/buildingsform_template.ui
    interface/changemindaysselectivelyform.cpp interface/changemindaysselectivelyform.h
    interface/changemindaysselectivelyform_template.ui
    interface/changeminhalfdaysselectivelyform.cpp interface/changeminhalfdaysselectivelyform.h
    interface/changeminhalfdaysselectivelyform_template.ui
    interface/commentsform.cpp interface/commentsform.h
    interface/commentsform_template.ui
    interface/daysform.cpp interface/daysform.h
    interface/daysform_template.ui
    interface/fet.cpp interface/fet.h
    interface/fetmainform.cpp interface/fetmainform.h
    interface/fetmainform_template.ui
    interface/getmodefornewfileform.cpp interface/getmodefornewfileform.h
    interface/getmodefornewfileform_template.ui
    interface/groupactivitiesininitialorderitemsform.cpp interface/groupactivitiesininitialorderitemsform.h
    interface/groupactivitiesininitialorderitemsform_template.ui
    interface/groupsform.cpp interface/groupsform.h
    interface/groupsform_template.ui
    interface/helpaboutform.cpp interface/helpaboutform.h
    interface/helpaboutform_template.ui
    interface/helpaboutlibrariesform.cpp interface/helpaboutlibrariesform.h
    interface/helpaboutlibrariesform_template.ui
    interface/helpalgeriaform.cpp interface/helpalgeriaform.h
    interface/helpalgeriaform_template.ui
    interface/helpblockplanningform.cpp interface/helpblockplanningform.h
    interface/helpblockplanningform_template.ui
    interface/helpfaqform.cpp interface/helpfaqform.h
    interface/helpfaqform_template.ui
    interface/helpinstructionsform.cpp interface/helpinstructionsform.h
    interface/helpinstructionsform_template.ui
    interface/helpmoroccoform.cpp interface/helpmoroccoform.h
    interface/helpmoroccoform_template.ui
    interface/helponstudentsminhoursdaily.cpp interface/helponstudentsminhoursdaily.h
    interface/helptermsform.cpp interface/helptermsform.h
    interface/helptermsform_template.ui
    interface/helptipsform.cpp interface/helptipsform.h
    interface/helptipsform_template.ui
    interface/hoursform.cpp interface/hoursform.h
    interface/hoursform_template.ui
    interface/institutionnameform.cpp interface/institutionnameform.h
    interface/institutionnameform_template.ui
    interface/listofrelatedspaceconstraintsform.cpp interface/listofrelatedspaceconstraintsform.h
    interface/listofrelatedspaceconstraintsform_template.ui
    interface/listofrelatedtimeconstraintsform.cpp interface/listofrelatedtimeconstraintsform.h
    interface/listofrelatedtimeconstraintsform_template.ui
    interface/listspaceconstraints.cpp interface/listspaceconstraints.h
    interface/listtimeconstraints.cpp interface/listtimeconstraints.h
    interface/longtextmessagebox.cpp interface/longtextmessagebox.h
    interface/longtextmessagebox_auxiliary.h
    interface/messageboxes.cpp interface/messageboxes.h
    interface/modifyactivityform.cpp interface/modifyactivityform.h
    interface/modifyactivityform_template.ui
    interface/modifybuildingform.cpp interface/modifybuildingform.h
    interface/modifybuildingform_template.ui
    interface/modifydayform.cpp interface/modifydayform.h
    interface/modifydayform_template.ui
    interface/modifygroupactivitiesininitialorderitemform.cpp interface/modifygroupactivitiesininitialorderitemform.h
    interface/modifygroupactivitiesininitialorderitemform_template.ui
    interface/modifyhourform.cpp interface/modifyhourform.h
    interface/modifyhourform_template.ui
    interface/modifyrealdayform.cpp interface/modifyrealdayform.h
    interface/modifyrealdayform_template.ui
    interface/modifyrealhourform.cpp interface/modifyrealhourform.h
    interface/modifyrealhourform_template.ui
    interface/modifyroomform.cpp interface/modifyroomform.h
    interface/modifyroomform_template.ui
    interface/modifystudentsgroupform.cpp interface/modifystudentsgroupform.h
    interface/modifystudentsgroupform_template.ui
    interface/modifystudentssubgroupform.cpp interface/modifystudentssubgroupform.h
    interface/modifystudentssubgroupform_template.ui
    interface/modifystudentsyearform.cpp interface/modifystudentsyearform.h
    interface/modifystudentsyearform_template.ui
    interface/modifysubactivityform.cpp interface/modifysubactivityform.h
    interface/modifysubactivityform_template.ui
    interface/modifyteacherform.cpp interface/modifyteacherform.h
    interface/modifyteacherform_template.ui
    interface/modifytimeconstraints.cpp interface/modifytimeconstraints.h
    interface/notificationcommandform.cpp interface/notificationcommandform.h
    interface/notificationcommandform_template.ui
    interface/randomseedform.cpp interface/randomseedform.h
    interface/randomseedform_template.ui
    interface/removeredundantconfirmationform.cpp interface/removeredundantconfirmationform.h
    interface/removeredundantconfirmationform_template.ui
    interface/removeredundantform.cpp interface/removeredundantform.h
    interface/removeredundantform_template.ui
    interface/restoredatastateform.cpp interface/restoredatastateform.h
    interface/restoredatastateform_template.ui
    interface/roommakeeditvirtualform.cpp interface/roommakeeditvirtualform.h
    interface/roommakeeditvirtualform_template.ui
    interface/roomsform.cpp interface/roomsform.h
    interface/roomsform_template.ui
    interface/savetimetableconfirmationform.cpp interface/savetimetableconfirmationform.h
    interface/savetimetableconfirmationform_template.ui
    interface/settingsautosaveform.cpp interface/settingsautosaveform.h
    interface/settingsautosaveform_template.ui
    interface/settingsdatatoprintintimetablesform.cpp interface/settingsdatatoprintintimetablesform.h
    interface/settingsdatatoprintintimetablesform_template.ui
    interface/settingsrestoredatafromdiskform.cpp interface/settingsrestoredatafromdiskform.h
    interface/settingsrestoredatafromdiskform_template.ui
    interface/settingsrestoredatafrommemoryform.cpp interface/settingsrestoredatafrommemoryform.h
    interface/settingsrestoredatafrommemoryform_template.ui
    interface/settingsstyleandcolorschemeform.cpp interface/settingsstyleandcolorschemeform.h
    interface/settingsstyleandcolorschemeform_template.ui
    interface/settingstimetablehtmllevelform.cpp interface/settingstimetablehtmllevelform.h
    interface/settingstimetablehtmllevelform_template.ui
    interface/sparseitemmodel.cpp interface/sparseitemmodel.h
    interface/sparsetableview.cpp interface/sparsetableview.h
    interface/splityearform.cpp interface/splityearform.h
    interface/splityearform_template.ui
    interface/spreadconfirmationform.cpp interface/spreadconfirmationform.h
    interface/spreadconfirmationform_template.ui
    interface/spreadmindaysconstraintsfivedaysform.cpp interface/spreadmindaysconstraintsfivedaysform.h
    interface/spreadmindaysconstraintsfivedaysform_template.ui
    interface/statisticsprintform.cpp interface/statisticsprintform.h
    interface/studentscomboboxesstyleform.cpp interface/studentscomboboxesstyleform.h
    interface/studentscomboboxesstyleform_template.ui
    interface/studentsnotavailabletimestimehorizontalform.cpp interface/studentsnotavailabletimestimehorizontalform.h
    interface/studentsnotavailabletimestimehorizontalform_template.ui
    interface/studentsstatisticsform.cpp interface/studentsstatisticsform.h
    interface/studentsstatisticsform_template.ui
    interface/subactivitiesform.cpp interface/subactivitiesform.h
    interface/subactivitiesform_template.ui
    interface/subgroupsform.cpp interface/subgroupsform.h
    interface/subgroupsform_template.ui
    interface/subjectsform.cpp interface/subjectsform.h
    interface/subjectsform_template.ui
    interface/subjectsstatisticsform.cpp interface/subjectsstatisticsform.h
    interface/subjectsstatisticsform_template.ui
    interface/teachersform.cpp interface/teachersform.h
    interface/teachersform_template.ui
    interface/teachersnotavailabletimestimehorizontalform.cpp interface/teachersnotavailabletimestimehorizontalform.h
    interface/teachersnotavailabletimestimehorizontalform_template.ui
    interface/teachersstatisticsform.cpp interface/teachersstatisticsform.h
    interface/teachersstatisticsform_template.ui
    interface/teachersubjectsqualificationsform.cpp interface/teachersubjectsqualificationsform.h
    interface/teachersubjectsqualificationsform_template.ui
    interface/termsform.cpp interface/termsform.h
    interface/termsform_template.ui
    interface/timetablegenerateform.cpp interface/timetablegenerateform.h
    interface/timetablegenerateform_template.ui
    interface/timetablegeneratemultipleform.cpp interface/timetablegeneratemultipleform.h
    interface/timetablegeneratemultipleform_template.ui
    interface/timetableprintform.cpp interface/timetableprintform.h
    interface/timetableshowconflictsform.cpp interface/timetableshowconflictsform.h
    interface/timetableshowconflictsform_template.ui
    interface/timetablestowriteondiskform.cpp interface/timetablestowriteondiskform.h
    interface/timetablestowriteondiskform_template.ui
    interface/timetableviewroomsdayshorizontalform.cpp interface/timetableviewroomsdayshorizontalform.h
    interface/timetableviewroomsdayshorizontalform_template.ui
    interface/timetableviewroomsdaysverticalform.cpp interface/timetableviewroomsdaysverticalform.h
    interface/timetableviewroomsdaysverticalform_template.ui
    interface/timetableviewroomstimehorizontalform.cpp interface/timetableviewroomstimehorizontalform.h
    interface/timetableviewroomstimehorizontalform_template.ui
    interface/timetableviewstudentsdayshorizontalform.cpp interface/timetableviewstudentsdayshorizontalform.h
    interface/timetableviewstudentsdayshorizontalform_template.ui
    interface/timetableviewstudentsdaysverticalform.cpp interface/timetableviewstudentsdaysverticalform.h
    interface/timetableviewstudentsdaysverticalform_template.ui
    interface/timetableviewstudentstimehorizontalform.cpp interface/timetableviewstudentstimehorizontalform.h
    interface/timetableviewstudentstimehorizontalform_template.ui
    interface/timetableviewteachersdayshorizontalform.cpp interface/timetableviewteachersdayshorizontalform.h
    interface/timetableviewteachersdayshorizontalform_template.ui
    interface/timetableviewteachersdaysverticalform.cpp interface/timetableviewteachersdaysverticalform.h
    interface/timetableviewteachersdaysverticalform_template.ui
    interface/timetableviewteacherstimehorizontalform.cpp interface/timetableviewteacherstimehorizontalform.h
    interface/timetableviewteacherstimehorizontalform_template.ui
    interface/utilities.cpp interface/utilities.h
    interface/yearsform.cpp interface/yearsform.h
    interface/yearsform_template.ui
)
target_include_directories(fet PRIVATE
    engine
    interface
)

#set_target_properties(fet PROPERTIES AUTOMOC TRUE)
#set_target_properties(fet PROPERTIES AUTORCC TRUE)
#set_target_properties(fet PROPERTIES AUTOUIC TRUE)

target_compile_definitions(fet PRIVATE
    QT_NO_FOREACH
    QT_NO_JAVA_STYLE_ITERATORS
    QT_NO_LINKED_LIST
    QT_STRICT_ITERATORS
)

#qt_add_translations(fet #MERGE_QT_TRANSLATIONS
#    TS_FILE_DIR translations
#    # Do not embed generated .qm files.
#    QM_FILES_OUTPUT_VARIABLE qm_files
#)

target_link_libraries(fet PRIVATE
    Qt::Core
    Qt::Gui
    Qt::Network
    Qt::Widgets
)

# Resources:
set_source_files_properties("interface/../../icons/fet.png"
    PROPERTIES QT_RESOURCE_ALIAS "images/appicon.png"
)
set(interface_images_resource_files
    "interface/../../icons/fet.png"
    "interface/images/group.png"
    "interface/images/locked.png"
    "interface/images/subgroup.png"
    "interface/images/unlocked.png"
)

qt_add_resources(fet "interface_images"
    PREFIX
        "/"
    BASE
        "interface"
    FILES
        ${interface_images_resource_files}
)

if(NOT QT.printsupport.name_ISEMPTY)
    target_link_libraries(fet PRIVATE
        Qt::PrintSupport
    )
endif()

install(TARGETS fet
    BUNDLE DESTINATION .
    RUNTIME DESTINATION ${CMAKE_INSTALL_BINDIR}
)

# Install generated .qm files.
#install(FILES ${qm_files} DESTINATION bin/translations)

# Qt >= 6.5.
#qt_deploy_translations(
#    LOCALES de fr
#)

install(DIRECTORY ../translations/ DESTINATION bin/translations
    FILES_MATCHING PATTERN "*.qm")

#install(DIRECTORY ${CMAKE_PREFIX_PATH}/translations/ DESTINATION bin/translations
#    FILES_MATCHING PATTERN "qtbase_*.qm")

#file(GLOB flist ../translations/*.qm)
#foreach(f ${flist})
#    get_filename_component(n ${f} NAME)
#    MESSAGE(STATUS "+++ ${n}")
#endforeach()

# Qt >= 6.3.
qt_generate_deploy_app_script(
    TARGET fet
    OUTPUT_SCRIPT deploy_script
    NO_TRANSLATIONS
    NO_UNSUPPORTED_PLATFORM_ERROR
    INCLUDE_PLUGINS wayland xcb
)
install(SCRIPT ${deploy_script})
