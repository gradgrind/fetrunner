# Build the app which is to be packaged.

# Use a different build directory so as not to clash with the basic version
BUILDDIR="${BUILDDIR}_with_fet"

echo
echo "Compile fetrunner -> $BUILDDIR/$INSTALLDIR/bin/fetrunner"
echo

mkdir -p "$BUILDDIR/$INSTALLDIR/bin"

go build -o $BUILDDIR/$INSTALLDIR/bin/fetrunner ../cmd/fetrunner
if [ $? -ne 0 ]; then
    echo "***!!!*** fetrunner build failed"
    exit 1
fi

FETRUNNER_VERSION=$("$BUILDDIR/$INSTALLDIR/bin/fetrunner" -v | sed 's/.*: *//')
echo "FETRUNNER_VERSION=$FETRUNNER_VERSION"

echo
echo "Compile libfetrunner -> $BUILDDIR/$APPNAME/libfetrunner.a"
echo

mkdir -p "$BUILDDIR/$APPNAME"

CGO_ENABLED=1 go build -buildmode=c-archive -o "$BUILDDIR/$APPNAME/" ../libfetrunner/libfetrunner.go
if [ $? -ne 0 ]; then
    echo "***!!!*** libfetrunner build failed"
    exit 1
fi

. "common_build.sh"

# Now build FET

APPNAME="FET"
APPDIR=$(ls -r ".." | grep -m 1 "fet-")
if [ -z "$APPDIR" ]; then
    echo "***!!!*** no FET source found"
    exit 1
fi
APPDIR="../$APPDIR"

FET_VERSION=$(<$APPDIR/VERSION)

. common_build.sh

PROJECT_VERSION="${FETRUNNER_VERSION}_${FET_VERSION}"

