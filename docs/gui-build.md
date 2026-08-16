# Building the GUI

The following instructions assume you are using the Qt development kit from the [Qt website](https://www.qt.io/development/download-qt-installer-oss), installed to the standard location.

 1) Compile `libfetrunner` as a [*static* library](./libfetrunner/README.md).

 2) Make a new directory `build` in the `fetrunner/gui` directory.

 3) Open the terminal (Linux) or Powershell (Windows) in this `build` directory.

 4) Compilation and "installation" is then platform-dependent. Also the path to the Qt installation will depend on the version installed (replace "6.11.1" by your version), but note that `fetrunner-gui` currently needs at least "6.10".

 5) The recommended way to run the resulting excutable is to copy it to the `FET` binary installation and use the Qt libraries available there. This should work if the `FET` binary installation uses a Qt version equal to or newer than that with which `fetrunner-gui` is compiled.

#### Compilation on Linux

```
$HOME/Qt/Tools/CMake/bin/cmake .. -DCMAKE_PREFIX_PATH=$HOME/Qt/6.11.1/gcc_64 -DCMAKE_INSTALL_PREFIX=install

$HOME/Qt/Tools/CMake/bin/cmake --build . --target install -j 4
```

Copy the resulting `fetrunner-gui` executable from `install/bin` to the `bin` directory of your `FET` installation.

Optionally, copy the `icons` directory from `fetrunner\gui` to  the `bin` directory of your `FET` installation.

#### Compilation on Windows

```
C:\Qt\Tools\CMake_64\bin\cmake.exe .. -DCMAKE_PREFIX_PATH=C:\Qt\6.11.1\mingw_64 -DCMAKE_GENERATOR="MinGW Makefiles"

C:\Qt\Tools\CMake_64\bin\cmake.exe --build . -j 4
```

Copy the resulting `fetrunner-gui.exe` executable from `build` to the root directory of your `FET` installation.

Optionally copy the `icons` directory from `fetrunner\gui` to  the root directory of your `FET` installation.
