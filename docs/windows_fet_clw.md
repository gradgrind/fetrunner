# Special `FET` command-line program for GUI version of `fetrunner` on Windows

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

