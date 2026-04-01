# `fetrunner` Structure

The package "autotimetable" is the heart of the `fetrunner` algorithm. Given the data from which a timetable is to be constructed it systematically enables and disables the constraints in an attempt to determine any constraints which may be difficult or impossible to fulfil, and tries to deliver a timetable – possible not fulfilling all the constraints within a set time.

It is designed around the idea of a "data source" and a "generation back-end", both of which can – in principle be implemented in various ways. The most straightforward is to use a `FET` file ("xxx.fet") as source and `FET` as the back-end. However, there is also support for a sort of JSON-like database structure as data source, which can be built or converted from the data structures of other timetable programs, if there is a suitable interface. The package "w365" supports the reading of custom data from the "Waldorf 365" school management software.
