# Building the GUI

The following instructions assume you are using the Qt development kit from the [Qt website](https://www.qt.io/development/download-qt-installer-oss), installed to the standard location.

The recommended way of building `fetrunner` with GUI is to use the [QAIFUS installer framework](https://github.com/gradgrind/fetinstaller), because the whole process is covered by scripts. The build dependencies are described there. For `fetrunner` it is also necessary to install [Go](https://go.dev) and [mdBook](https://github.com/rust-lang/mdBook). The `go` and `mdbook` commands must be in the `PATH`; the standard installation procedure for `Go` will cover this, but the `mdbook` executable must be placed somewhere suitable manually (e.g. `$HOME/bin` on Linux or `C:\Users\<user>\AppData\Local\Microsoft\WindowsApps` on Windows). After these are installed it may be necessary to log out and log back in.

Guidelines for building without `QAIFUS` are provided [below](#building-the-components-separately).

To work with `QAIFUS` there are two configurations, one including a full `FET` installation ("QAIFUS_fetrunner_with_fet") and one without ("QAIFUS_fetrunner").

On Linux the latter contains no `FET` components, so to use it it is necessary to also install a `FET` package.

On Windows a specially compiled `FET` command-line program (one which doesn't pop up a console window – see [FET on Windows](#fet-on-windows)) is necessary, so the compilation of just this bit of `FET` is included in the `fetrunner` build. It is then possible to use `fetrunner` without a `FET` installation, but a separate `FET` installation is probably also desirable in this case.

The "fetrunner_with_fet" configuration compiles the full `FET` package along with `fetrunner`, producing a combined installer.

Once all the build dependencies have been installed:

 - Copy the QAIFUS directory into the `fetrunner` root directory.

 - Unpack the `FET` source directory (something like `fet-7.13.0`) into the `fetrunner` root directory.

 - Linux, without `FET`:
```path/to/QAIFUS/BUILD_ALL.sh fetrunner```

- Linux, with `FET`:

```
path/to/QAIFUS/BUILD_ALL.sh fetrunner_with_fet
```

 - Windows, without `FET`:

```
powershell -ExecutionPolicy Bypass -File path\to\QAIFUS\BUILD_ALL.ps1 -Argument-List fetrunner
```

 - Windows, with `FET`:

```
powershell -ExecutionPolicy Bypass -File path\to\BUILD_ALL.ps1 -Argument-List fetrunner_with_fet
```

These commands should (eventually) produce runnable installers. Note that the two versions (with and without `FET`) use distinct build directories, so they don't interfere with each other.

## Building the components separately

There might be a good reason for wanting to build the various parts manually instead of using `QAIFUS`. Here are some guidelines. A look at the `QAIFUS` scripts might also be helpful.  

 1) Compile `libfetrunner` as a [*static* library](./libfetrunner-compile.md).

 2) Make a new directory `build` in the `fetrunner` base directory.

 3) Open the terminal (Linux) or Powershell (Windows) in this `build` directory.

 4) Compilation and "installation" is then platform-dependent. Also the path to the Qt installation will depend on the version installed (replace "6.11.1" by your version), but note that `fetrunner-gui` currently needs at least "6.10".

### Compilation on Linux

```
$HOME/Qt/Tools/CMake/bin/cmake .. -DCMAKE_PREFIX_PATH=$HOME/Qt/6.11.1/gcc_64 -DCMAKE_INSTALL_PREFIX=install

$HOME/Qt/Tools/CMake/bin/cmake --build . --target install -j 4
```

Copy the resulting `fetrunner-gui` executable from `install/bin` to the `bin` directory of your `FET` installation.

Optionally, copy the `icons` directory from `fetrunner\gui` to  the `bin` directory of your `FET` installation.

### Compilation on Windows

```
C:\Qt\Tools\CMake_64\bin\cmake.exe .. -DCMAKE_PREFIX_PATH=C:\Qt\6.11.1\mingw_64 -DCMAKE_GENERATOR="Ninja"

C:\Qt\Tools\CMake_64\bin\cmake.exe --build . -j 4

or, to do the install stage:

C:\Qt\Tools\CMake_64\bin\cmake.exe .. -DCMAKE_PREFIX_PATH=C:\Qt\6.11.1\mingw_64 -DCMAKE_GENERATOR="Ninja" -DCMAKE_INSTALL_PREFIX=install

C:\Qt\Tools\CMake_64\bin\cmake.exe --build . --target install -j 4
```

Copy the resulting `fetrunner-gui.exe` executable from `build` to the root directory of your `FET` installation.

Optionally copy the `icons` directory from `fetrunner\gui` to  the root directory of your `FET` installation.

#### FET on Windows

`fetrunner-gui`, the GUI version of `fetrunner` needs a special version of the `FET` command-line program, `fet-cl.exe`. It can't use `fet-cl.exe`, as that would pop up a console window every time it was run (and that would be a *lot* of console windows ...). To distinguish this special version from the original, the executable is named `fet-clw.exe`.

To build the necessary `fet-clw.exe` program you need to download the `FET` source package and unpack this somewhere convenient.

 - Make a new directory, `build-clw`, in the root directory of the `FET` source code.

 - Open `Powershell` in this `build-clw` directory.

 - Compile: Note that the `Qt` version (here "6.11.1") will need to be changed to that of the version you are using.

```
C:\Qt\Tools\CMake_64\bin\cmake.exe .. -DCMAKE_PREFIX_PATH=C:\Qt\6.11.1\mingw_64 -DCMAKE_GENERATOR="MinGW Makefiles" -DCOMMAND_LINE_ONLY=ON -DNO_WINDOWS_CONSOLE=ON

C:\Qt\Tools\CMake_64\bin\cmake.exe --build . -j 4
```
 - Copy the resulting `fet-clw.exe` executable from `build` to the root directory of your `FET` installation.

