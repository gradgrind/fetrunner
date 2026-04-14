#!/bin/bash

# Trap SIGINT (Ctrl+C), etc. to quit properly
trap "exit 0" SIGINT SIGTERM SIGKILL

# Run the test data files in the subdirectories "tests_fet" (".fet" files)
# and tests_w365 ("_w365.json" files).

runtest () {
    b=$(basename -- "$1")
    d="$2/${b%.*}"
    mkdir -p "$d"
    f="$d/$b"
    cp "$1" "$f"
    echo "***** $f"
    go run ../cmd/fetrunner "$f"
    sub_pid=$!
    tac "${f%.*}.log" | grep -m1 ".NCONSTRAINTS="
    tac "${f%.*}.log" | grep -m2 ".TICK=" | grep -v "=-1"
}

tdir1="$PWD/tests_fet"
rdir1="$tdir1/_results"
rm -rf "$rdir1"

tdir2="$PWD/tests_w365"
rdir2="$tdir2/_results"
rm -rf "$rdir2"

for f0 in $(ls $tdir1/*.fet); do
    runtest "$f0" "$rdir1"
done

for f0 in $(ls $tdir2/*_w365.json); do
    runtest "$f0" "$rdir2"
done

### Useful commands
#filename=$(basename -- "$fullfile")
#extension="${filename##*.}"
#filename="${filename%.*}"
