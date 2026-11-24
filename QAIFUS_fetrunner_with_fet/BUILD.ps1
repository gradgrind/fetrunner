# Build the app which is to be packaged.

# Use a different build directory so as not to clash with the basic version
$BUILDDIR = "${BUILDDIR}_with_fet"

Write-Host
Write-Host "Compile fetrunner -> $BUILDDIR\$INSTALLDIR\fetrunner"
Write-Host

# Ensure that the installation directory exists
$idir = "$BUILDDIR\$INSTALLDIR"
if ( -not ( Test-Path -Path "$idir" ) ) {
    New-Item -Path "$idir" -ItemType Directory -ErrorAction Stop
}

go build -o "$BUILDDIR\$INSTALLDIR\" "..\cmd\fetrunner"
if ( !$? )
{
    Write-Host "***!!!*** fetrunner build failed"
    exit 1
}

$FETRUNNER_VERSION = (& ".\$BUILDDIR\$INSTALLDIR\fetrunner" -v) -replace ".*: *", ""
Write-Host "FETRUNNER_VERSION=$FETRUNNER_VERSION"

Write-Host
Write-Host "Compile libfetrunner -> $BUILDDIR\$APPNAME\libfetrunner.a"
Write-Host

# Ensure that the build directory exists
$bdir = "$BUILDDIR\$APPNAME"
if ( -not ( Test-Path -Path "$bdir" ) ) {
    New-Item -Path "$bdir" -ItemType Directory -ErrorAction Stop
}

$Env:CGO_ENABLED=1
go build -buildmode=c-archive -o "$BUILDDIR\$APPNAME\" "..\libfetrunner\libfetrunner.go"
if ( !$? )
{
    Write-Host "***!!!*** libfetrunner build failed"
    exit 1
}

. ".\common_build.ps1"

# Now build FET

$APPDIR = Get-ChildItem -Path ".." -Filter "fet-*" | Select-Object -ExpandProperty Name -Last 1
if ( [string]::IsNullOrEmpty($APPDIR) )
{
    Write-Host "***!!!*** no FET source found"
    exit 1
}
$APPDIR = "..\$APPDIR"
$FET_VERSION = Get-Content  -ErrorAction Stop "$APPDIR\VERSION"

# The special command-line version for fetrunner
$APPNAME = "fet-clw"

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

$PROJECT_VERSION = "${FETRUNNER_VERSION}_${FET_VERSION}"

# Now build the "normal" FET
$APPNAME = "FET"

. ".\common_build.ps1"

