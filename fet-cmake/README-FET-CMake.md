# Build instructions for `FET` + `fetrunner`

In the `FET` root folder:

 - Copy in `CMakeLists.txt` from this directory.
 
 - Copy in the folders/files from `src`, `src-cl` and `src-clw` in this directory.

 - Copy in the complete `fetrunner` folder.
 
Compile libfetrunner: [Build `libfetrunner`](../libfetrunner/README.md).

## Linux

```
mkdir build

cd build

cmake -DCMAKE_PREFIX_PATH=$HOME/Qt/6.10.1/gcc_64 .. -DINSTALL_BASE=usr

cmake --build . --target install -j 4
```

Copy the result (usr/ from build/) into the `FET` binary tree. It should enable native file dialogs and Wayland support. But overwrites (incl. usr/bin/qt.conf) should probably be skipped.

## Windows

Using the standard `Qt` installation, the command lines should be something like this (running from the `build` sub-directory:

```
C:\Qt\Tools\CMake_64\bin\cmake.exe .. -DCMAKE_PREFIX_PATH=C:\Qt\6.10.1\mingw_64 -DINSTALL_BASE=install -DCMAKE_GENERATOR="MinGW Makefiles"

C:\Qt\Tools\CMake_64\bin\cmake.exe --build . --target install -j 4
```

Probably only the executables need to be copied into the `FET` binary tree.

In case something is missing in the configuration, the following fuller example may be useful:

```
C:\Qt\Tools\CMake_64\bin\cmake.exe .. -DCMAKE_PREFIX_PATH=C:\Qt\6.10.1\mingw_64 -DINSTALL_BASE=install -DCMAKE_CXX_COMPILER=C:\Qt\Tools\mingw1310_64\bin\g++.exe -DCMAKE_C_COMPILER=C:\Qt\Tools\mingw1310_64\bin\gcc.exe -DCMAKE_BUILD_TYPE=Release -DCMAKE_MAKE_PROGRAM=C:\Qt\Tools\mingw1310_64\bin\mingw32-make.exe -DCMAKE_GENERATOR="MinGW Makefiles"
```
