# Structure of the JSON result

The JSON object presents information about the last successful "base" instance: which constraints were disabled and the placement of the activities. For individual constraints which led to `FET` errors, the `FET` message is recorded. The information is presented in terms of item indexes. That is, each presented item (day, hour, activity, constraint, room, ...) is referred to by an index (starting at 0). In order to correlate these items with those in the data source, there is a list of references for each of these item types. Each of these referenced objects has an "Id" field, which is a reference to the corresponding source item. Except for the constraint objects, they also have a "Tag" field, giving a short "name" for the item, which may be of relevance to the back-end.

The items in the constraint list have additional fields:

 - "CType" (constraint name), as used in the source, so it may differ from the `FET` constraint name
 - "Data" depends on constraint type, it can be a simple value or a key/value map
 - "Weight", an integer in the range 0 (ineffective) to 100 ("hard") – with `FET` source a non-linear conversion function is used

Not every source constraint need have its own reference, some may be attached to a teacher or class, for example. In this case the item referred to would be passed in the Data field, the source-reference being empty. A constraint on a group of activities might place the indexes of these activities as a comma-separated list in the Data field.

Bear in mind that the actual constraints used in the generation are back-end constraints (i.e. `FET` constraints), even if their "CType" fields refer to the source. There may not be an easy 1:1 correspondence in some cases. Multiple `FET` constraints might derive from a single source constraint. The reverse case is currently not supported.

### Reporting "impossible" constraints

If any individual constraint was found to cause a FET error report, these reports are recorded in the "ConstraintErrors" map, the key being the constraint index, the value the message. TODO: Should "timed out" constraints also be reported here?
