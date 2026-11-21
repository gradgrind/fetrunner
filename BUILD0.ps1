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

# Set the application name
$APPNAME = "fetrunner"

$APPDIR = "."

New-Item -ItemType "Directory" -Path "$BUILDDIR\$APPNAME"

$Env:CGO_ENABLED=1
go build -buildmode=c-archive -o "`"$BUILDDIR\$APPNAME\`"" libfetrunner\libfetrunner.go
if ( !$? )
{
    exit 1
}

. ".\common_build.ps1"
