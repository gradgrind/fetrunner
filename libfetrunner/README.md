# Build `libfetrunner`

These commands should be run within the `libfetrunner` directory.

## To build a static library

```
go build -buildmode=c-archive libfetrunner.go
```

This will normally set CGO_ENABLED=1 automatically, but under some circumstances (e.g. cross compilation) this may need to be set explicitly:

```
CGO_ENABLED=1 go build -buildmode=c-archive libfetrunner.go
```

or, on Windows:

```
set CGO_ENABLED=1

go build -buildmode=c-archive libfetrunner.go
```


## To build a shared library on Linux

```
go build -buildmode=c-shared -o libfetrunner.so libfetrunner.go
```

On Windows and MacOS the output file will presumably need adapting.

## Cross-compile Linux to Windows:

```
GOOS=windows GOARCH=amd64 CGO_ENABLED=1 CC=x86_64-w64-mingw32-gcc go build -buildmode=c-archive libfetrunner/libfetrunner.go
```
