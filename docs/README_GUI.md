# Building `fetrunner` together with `FET`

The GUI for `fetrunner` has been developed using `CMake`. `FET` is built with `qmake`. I have written some `CMakeLists.txt` files for `FET` so that it is possible to build part or all of `FET` with `CMake`, which is the current standard build system for `Qt`.

On Windows, the GUI version of `fetrunner` requires a special build, `fet-clw.exe`, of the `FET` command-line program which doesn't pop up a console with every run. This can be done by slightly modifying the `FET` sources, as described in the main [README](../README.md#special-note-for-windows-users). By using the supplied `CMake` files in `fetrunner/fet-cmake` it is, however, possible to build this together with `fetrunner`. For build instructions see [README-Cmake](../fet-cmake/README-FET-CMake.md).

The `CMake` build has some advantages, but requires a recent version of `Qt` ( `Qt6.10` for one feature).

`fetrunner` includes the basic structures for GUI-translation, with the possibility of embedding the translations in the binary, including `Qt`'s own translations. A German translation is provided as an example. This is all managed within the `CMakeLists.txt` file.

It is possible to build an installation package using functions within `CMake`, possibly making `linuxdeploy` unnecessary on Linux (except for producing an `AppImage`). Unlike the `FET` standard binary distributions these would only include the required `Qt` libraries, not the system libraries these depend on, but in many cases that should be enough – if the build system is not too new.

### Notes

It seems the current `FET` distribution doesn't include `Wayland` support on Linux. Apparently it still runs on a `Wayland` system (I tested it on a recent Fedora) using `XWayland`, but using a recent feature of `CMake` on `Qt` it is possible to include direct support for `Wayland` in the `FET` package.

Running `FET` on Linux, the file system is accessed using `Qt`'s own file dialogs, rather than the native ones, at least where I have tested it. This seems to have something to do with the packaging using `linuxdeploy`. With a `CMake`-built package, the native file dialogs are used.

`FET` doesn't use `Qt`'s own translations. This is occasionally visible in a dialog (only in Linux, only the file dialogs?). It is probably a very minor issue. However, to use these, the language-loading code would need to be adapted and the translations included in the distribution package.
