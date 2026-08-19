# Preparing for a timetable-generation run

Although the project currently goes under the name "fetrunner", one of its aims is to be able to support generation back-ends other than `FET`, eventually, by means of an interface consisting of function signatures.

Note that at present, when the input file is a `FET` file, no database is built. The "autotimetable" algorithm works directly on the XML data structure of the input file. In a way this reflects the initial impulse for `fetrunner`, the database structures are an addition to allow input data from other systems. `FET` supports many more constraints than the current database, and the ones they do have in common may not always be directly equivalent. `FET` also has a more flexible notion of student groups, which may be difficult to "translate".

So although it might be possible to read a `FET` input file into the database structures, some information might be lost and – if using the `FET` back-end – it would be superfluous.

**TODO:** Detail the differences for the "autotimetable" functions when using `FET` input vs. database data.

The information relevant for timetabling needs to be extracted from the database and the structures used by the back-end need to be built.

