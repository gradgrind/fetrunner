# Build script for Windows (Powershell).

# Run this script, e.g.:
#    powershell -ExecutionPolicy Bypass -File BUILD0.ps1

# If necessary, the working directory will be changed to run in the directory
# containing this script.

cd "$PSScriptRoot"
. ".\common_base.ps1"
if ( !$? )
{
    exit 1
}

$FETVERSION = "7.12.1"

# Set the application name
$APPNAME = "fetrunner"
$APPDIR = "."

Write-Host
Write-Host "Compile libfetrunner -> $BUILDDIR\$APPNAME\libfetrunner.go"
Write-Host

if ( -not(Test-Path -Path "$BUILDDIR\$APPNAME") ) {
    New-Item -ItemType "Directory" -Path "$BUILDDIR\$APPNAME"
}

$Env:CGO_ENABLED=1
go build -buildmode=c-archive -o "$BUILDDIR\$APPNAME\" libfetrunner\libfetrunner.go
if ( !$? )
{
    Write-Host "***!!!*** libfetrunner build failed"
    exit 1
}

. ".\common_build.ps1"

$APPNAME = "fet-clw"
$APPDIR = "..\fet-$FETVERSION"

Write-Host
Write-Host "Building $APPNAME ..."
Write-Host

cmake -B "`"$BUILDDIR\$APPNAME`"" -S "`"$APPDIR`"" `
    -DCMAKE_PREFIX_PATH="$QTDIR\$QTVERSION\mingw_64" `
    -DCMAKE_GENERATOR="Ninja" `
    -DCMAKE_INSTALL_PREFIX="`"$BUILDDIR\$INSTALLDIR`"" `
    -DCOMMAND_LINE_ONLY=ON `
    -DNO_WINDOWS_CONSOLE=ON

if ( !$? )
{
    Write-Host "***!!!*** $APPNAME configuration failed"
    exit 1
}

cmake --build "`"$BUILDDIR\$APPNAME`"" --target install --parallel 6

if ( !$? )
{
    Write-Host "***!!!*** $APPNAME build failed"
    exit 1
}

