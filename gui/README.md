# Building `fetrunner` together with `FET`

Although the GUI for `fetrunner` was developed using CMake, I have constructed a `.pro` file so that it can be integrated in the `FET` build process.

The `fetrunner` GUI can be built alongside `FET` by adding the `fetrunner` source directory to the `FET` source:

 - Copy (or unpack) the `fetrunner` source tree to the base of the `FET` source tree.

 - In `fet.pro`, add the fetrunner .pro file to SUBDIRS:
      
    SUBDIRS = src/src.pro src/src-cl.pro fetrunner/gui/fetrunner-gui.pro

On Windows, further changes are necessary. To prevent a console being popped up every time `fet-cl` is run, the command-line version must be modified:

 - Copy `src/src-cl.pro` to `src/src-clw.pro` and remove cmdline from CONFIG in the new file:
 
    CONFIG += release warn_on c++17 no_keywords
    
 - Add the new file to `fet.pro`:
 
    SUBDIRS = src/src.pro src/src-cl.pro src/src-clw.pro fetrunner/gui/fetrunner-gui.pro

In addition, the `Go` language must be installed to compile the `fetrunner` back-end. The compilation of this back-end is easy. In directory `fetrunner/libfetrunner` run:

    CGO_ENABLED=1 go build -buildmode=c-archive libfetrunner.go

Then build `FET` as in its README file, for example. At the end there should be an additional executable, `fetrunner-gui`.
