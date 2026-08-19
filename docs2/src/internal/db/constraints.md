# Constraints

Some constraints are implicit (such as those preventing collisions). Others are specified along with the items they constrain, for example blocked times or number of gaps per day for teachers or classes. Others are specified in the top-level Constraints list. These often constrain the activities of one or more courses, perhaps the relationships between activities of different courses. The currently supported constraint types in the Constraints list are described below.

## The Constraints List

All constraints have a Weight field, which is an integer between 0 and 100. A value of 0 means the constraint is inactive, a value of 100 specifies a "hard" constraint.

### AutomaticDifferentDays

The activities of a course should all be on different days.

This constraint may be specified at most once. If it is not present, a hard constraint affecting all courses is implemented. Note, however, that this constraint can be overridden for individual courses by DaysBetween constraints (see below).

 - Weight (integer)
 - ConsecutiveIfSameDay (boolean): In cases where the constraint is not fulfilled, the activities should be consecutive. This is only relevant if the constraint is soft, but if it is true this part of the constraint is always hard!

### DaysBetween

The activities of a course should have at least the given number of days between them (where a value of 1 is equivalent to different-days).

 - Weight (integer)
 - Courses (list of Course or SuperCourse references): The constraint will apply to each of the listed courses.
 - DaysBetween (integer)
 - ConsecutiveIfSameDay (boolean): If the activities do end up on the same day, they should be consecutive (always hard!).

If DaysBetween is 1, this constraint will override AutomaticDifferentDays for the listed courses. Otherwise it is a distinct constraint.

### DaysBetweenJoin

This constraint applies between the individual activities of the two courses, not between the activities of a course itself. That is, between course 1, activity 1 and course 2 activity 1; between course 1, activity 1 and course 2, activity 2, etc. Otherwise it is similar to the DaysBetween constraint.

 - Weight (integer)
 - Course1 (Course or SuperCourse reference)
 - Course2 (Course or SuperCourse reference)
 - DaysBetween (integer)
 - ConsecutiveIfSameDay (boolean): If the activities do end up on the same day, they should be consecutive (always hard!).

**NOTE:** The `ConsecutiveIfSameDay` flag should be respected by all the "days-between" constraints even if the weight is 0.

### ActivitiesEndDay

The activities of the specified course should be the last activities of the day for the student group concerned.

 - Weight (integer)
 - Course (Course or SuperCourse reference)

### AfterHour

The activities of the specified courses should lie after the given hour (excluding the given hour).

 - Weight (integer)
 - Courses (list of Course or SuperCourse references): The constraint will apply to each of the listed courses.
 - Hour (integer)

### BeforeHour

The activities of the specified courses should lie before the given hour (excluding the given hour).

 - Weight (integer)
 - Courses (list of Course or SuperCourse references): The constraint will apply to each of the listed courses.
 - Hour (integer)

## ParallelCourses

Die activities of the specified courses should start at the same time. The number and duration of the activities must be the same in all the courses.

 - Weight (integer)
 - Courses (list of Course or SuperCourse references): The constraint will apply to each of the listed courses.

**TODO**: The following constraints are not yet finally specified. It might be better to limit them to particular classes or to specify them in some other way.

## DoubleActivityNotOverBreaks

This constraint should be present at most once. A double-lesson should not cover a break time. The breaks are specified by the "Hours" list, the breaks being immediately before the given hours.

 - Weight (integer)
 - Hours (list of integers)

## MinHoursFollowing

There should be at least the given number of hours between the end of the activities of the first course and the start of those of the second course, if they are on the same day.

 - Weight (integer)
 - Course1 (Course or SuperCourse reference)
 - Course2 (Course or SuperCourse reference)
 - Hours (integer)
