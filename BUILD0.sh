#!/bin/bash

# Build script for Linux.

# If necessary, the working directory will be changed to run in the directory
# containing this script.

SCRIPT=$(readlink -f "$0")
BASEDIR=$(dirname "$SCRIPT")
cd $BASEDIR
. "common_base.sh"

APPNAME="fetrunner"
APPDIR="."

echo
echo "Compile libfetrunner -> $BUILDDIR/$APPNAME/libfetrunner.a"
echo

mkdir -p "$BUILDDIR/$APPNAME"

CGO_ENABLED=1 go build -buildmode=c-archive -o "$BUILDDIR/$APPNAME/" libfetrunner/libfetrunner.go
if [ $? -ne 0 ]; then
    echo "***!!!*** libfetrunner build failed"
    exit 1
fi

. "common_build.sh"
