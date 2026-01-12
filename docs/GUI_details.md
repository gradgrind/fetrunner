# Some details of the GUI code

## Dynamic updating during a run: _POLL_TT

In `threadrun.cpp` there is a worker thread, `RunThreadWorker::ttrun()` which repeatedly sends a "_POLL_TT" command to the back-end, blocking until it completes. Then it dispatches the results which are interesting for the GUI by emitting signals. These signals are forwarded to signals of the `RunThreadController`, which encapsulates the handling of the worker thread, isolating it from the main code. These signals are thus connected indirectly to methods of `MainWindow` with the same name.

 - ".TICK" -> `MainWindow::ticker()`, except when its value is `-1`, which will cause the worker thread to finish, after processing the results of this poll.

 - ".NCONSTRAINTS" -> `MainWindow::nconstraints()`

 - ".PROGRESS" -> `MainWindow::iprogress()`

 - ".START" -> `MainWindow::istart()`

 - ".END" -> `MainWindow::iend()`

 - ".ACCEPT" -> `MainWindow::iaccept()`

