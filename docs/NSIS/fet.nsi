;--------------------------------
; Includes

  !include "LogicLib.nsh"

;--------------------------------
; Custom defines

  !define APPNAME "FET"
  !define APPFILE "fet.exe"
  !define VERSION "7.10.1"
  !define SLUG "${NAME} v${VERSION}"

;--------------------------------
; General

  ; rtf or txt file - remember if it is txt, it must be in the DOS text format (\r\n)
  LicenseData "license.txt"
  ; This will be in the installer/uninstaller's title bar
  Name "${APPNAME}"
  OutFile "${APPNAME}-Setup.exe"
  RequestExecutionLevel user

;  InstallDir "$PROGRAMFILES\${APPNAME}"
;  InstallDirRegKey HKCU "Software\${APPNAME}" ""
;  RequestExecutionLevel admin

  InstallDir "$LOCALAPPDATA\${APPNAME}"

;--------------------------------
; UI
  Icon fet.ico

; Install
  Page License
  Page Directory "" "" DirLeave
  Page InstFiles

Function DirLeave
    ${If} ${FileExists} "$InstDir\*"
        MessageBox MB_YESNO `"$InstDir" already exists, delete it's content and continue installing?` IDYES yep
            Quit
          yep:
            RMDir /r "$InstDir"
            RMDir /r "$STARTMENU\FET"
    ${EndIf}
FunctionEnd

; ??? Automatically start app after installation
function .onInstSuccess
    Exec "$INSTDIR\${APPFILE}"
functionEnd

section "install"
	; Files for the install directory - to build the installer, these should be in the same directory as the install script (this file)
	SetOutPath $INSTDIR
	; Files added here should be removed by the uninstaller (see section "uninstall")
    File /r "install\*.*"

    File "fet.ico"
    File "license.txt"
	; Add any other files for the install directory here
 
	; Uninstaller - See function un.onInit and section "uninstall" for configuration
	WriteUninstaller "$INSTDIR\uninstall.exe"
 
	; Start Menu
	createDirectory "$STARTMENU\FET"
	createShortCut "$STARTMENU\FET\${APPNAME}.lnk" "$INSTDIR\${APPFILE}" "" "$INSTDIR\fet.ico"
	createShortCut "$STARTMENU\FET\Uninstall.lnk" "$INSTDIR\uninstall.exe"
sectionEnd


Section -ShellAssoc

    ;register file extensions
    WriteRegStr ShCtx ".fet" "" "FET"
    WriteRegStr ShCtx "FET\shell\open\command" "" "$\"$INSTDIR\fet.exe$\" -f $\"%1$\""

SectionEnd

# Uninstaller
 
function un.onInit
	SetShellVarContext Current
 
	; Verify the uninstaller - last chance to back out
	MessageBox MB_OKCANCEL "Permanantly remove ${APPNAME}?" IDOK next
		Abort
	next:
functionEnd

Section -un.ShellAssoc
    # Unregister file type
    DeleteRegKey ShCtx ".fet"
    DeleteRegKey ShCtx "FET\shell\open\command"
SectionEnd

;--------------------------------
; Remove empty parent directories

Function un.RMDirUP
    !define RMDirUP '!insertmacro RMDirUPCall'

    !macro RMDirUPCall _PATH
          push '${_PATH}'
          Call un.RMDirUP
    !macroend

    ; $0 - current folder
    ClearErrors

    Exch $0
    ;DetailPrint "ASDF - $0\.."
    RMDir "$0\.."

    IfErrors Skip
    ${RMDirUP} "$0\.."
    Skip:

    Pop $0
FunctionEnd

section "uninstall"
 
	; Remove Start Menu launchers
	delete "$STARTMENU\FET\${APPNAME}.lnk"
    delete "$STARTMENU\FET\Uninstall.lnk"
	; Try to remove the Start Menu folder - this will only happen if it is empty
	rmDir "$STARTMENU\FET"
 
	; Delete Folder
    rmDir /r "$INSTDIR"
    ${RMDirUP} "$INSTDIR"
 
	; Always delete uninstaller as the last action
	delete $INSTDIR\uninstall.exe
 
	; Try to remove the install directory - this will only happen if it is empty
	rmDir $INSTDIR
 
	; Remove uninstaller information from the registry
	;DeleteRegKey HKLM "Software\Microsoft\Windows\CurrentVersion\Uninstall\${COMPANYNAME} ${APPNAME}"
sectionEnd

