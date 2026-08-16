# fetrunner

![fetrunner-gui](./docs/images/Screenshot_00.png)

`fetrunner` is an assistant for [`FET`](https://lalescu.ro/liviu/fet/). `FET` is a free timetable generator program for educational establishments. It is widely used and very good at what it does. However, in the case of timetable data which "doesn't work" (because of conflicting constraints), it can sometimes be difficult to find where the problem lies. Also, with some data (activities and constraints) the calculation of a "solution" (a conflict-free timetable) can take a very long time. Whilst working on a timetable, it can be useful to know which constraints may be difficult to fulfil, without waiting a long time for `FET` to complete (or not ...).

Given a `FET` file, `fetrunner` aims to produce a "solution" (a timetable without placement conflicts) within a specified time, if necessary by deactivating some of the constraints. Thus it can help to find difficult (or potentially impossible) constraints. It can be run as a command-line program, but there is also a convenient GUI version which shows how a run is progressing and can also display the resulting timetables.

## Getting `fetrunner`

There are binary packages on the "Releases" page. These need to be unpacked in the root folder of a recent `FET` installation.

For information on using `fetrunner` and how it works, see the [documentation](./docs/index.md).
