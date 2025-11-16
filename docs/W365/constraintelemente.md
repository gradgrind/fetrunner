# Die Constraint-Elemente

Hier wird die Struktur der expliziten Constraint-Elemente in der w365.json-Datei beschrieben. Die Constraints, die sich auf die Ressourcen beziehen (Klassen, Lehrer, Räume), erscheinen in den Objekten dieser Ressourcen.

Alle Felder sollten Werte haben, Voreinstellungen gibt es nur in Ausnahmefällen.

Im Allgemeinen kann ein Kurs ein Course-Element oder ein SuperCourse-Element sein.

## ActivitiesEndDay

Die Activities des Kurses sollten am Ende des Schülertags liegen.

```
{
    "Constraint" :  "MARGIN_HOUR",
    "Weight" :      73,
    "Course" :      "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
}
```

## AfterHour

Die Activities der Kurse sollten nach der angegebenen Stunde – ausschließlich dieser Stunde – liegen.

```
{
	"Constraint":   "AFTER_HOUR",
	"Weight":       100,
	"Courses":      [
        "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
    ],
	"Hour":         4
}
```

## BeforeHour

Die Activities der Kurse sollten vor der angegebenen Stunde – ausschließlich dieser Stunde – liegen.

```
{
	"Constraint":   "BEFORE_HOUR",
	"Weight":       100,
	"Courses":      [
        "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
    ],
	"Hour":         4
}
```

## AutomaticDifferentDays

Die Activities eines Kurses sollen an unterschiedlichen Tagen stattfinden. Mit „"ConsecutiveIfSameDay": true“ sollten sie – falls sie doch am selben Tag sind – direkt nacheinander sein.

Dieser Constraint wird im Prinzip auf alle Kurse (mit zwei oder mehr Activities) angewendet. Wenn dieser Constraint nicht vorhanden ist, wird er mit "Weight": 100 angewendet.

Einzelne Kurse können durch DaysBetween-Constraints anders geregelt werden.

```
{
	"Constraint":           "AUTOMATIC_DIFFERENT_DAYS",
	"Weight":               100,
	"ConsecutiveIfSameDay": true
}
```

## DaysBetween

Dieser Constraint ist wie AutomaticDifferentDays, erlaubt aber andere Tagesabstände als 1 und wird auf einzelne Kurse angewendet. Mit „"DaysBetween": 1“ wird der globale Constraint AutomaticDifferentDays für diese Kurse ausgesetzt.

```
{
	"Constraint":           "DAYS_BETWEEN",
	"Weight":               100,
	"Courses":              [
        "2edfe663-c62b-4d05-ace2-0bedb0f4b672"
    ],
	"DaysBetween":          2.
	"ConsecutiveIfSameDay": true
}
```

## DaysBetweenJoin

Anders als DaysBetween wird dieser Constraint zwischen den einzelnen Stunden zweier verschiedener Kurse angewendet, also Kurs 1, Activity1 : Kurs 2, Activity 1; Kurs 1, Activity 2 : Kurs 2, Activity 2; Kurs 1, Activity2 : Kurs 2, Activity 1; ...

```
{
	"Constraint":           "DAYS_BETWEEN_JOIN",
	"Weight":               100,
	"Course1":              "2edfe663-c62b-4d05-ace2-0bedb0f4b672",
	"Course2":              "5fda67de-bbb3-48a2-a098-d957796b7743",
	"DaysBetween":          1,
	"ConsecutiveIfSameDay": false
}
```

## ParallelCourses

Die Activities der Kurse sollen gleichzeitig stattfinden. Die Anzahl und Länge der Activities müssen in allen Kursen gleich sein.

```
{
	"Constraint":           "PARALLEL_COURSES",
	"Weight":               100,
	"Courses":              [
        "2edfe663-c62b-4d05-ace2-0bedb0f4b672",
	    "5fda67de-bbb3-48a2-a098-d957796b7743"
    ],
}
```

**TODO**: Folgende Constraints sind noch nicht klar. Es kann gut sein, dass sie auf bestimmte Klassen (...) begrenzt oder anders formuliert sein sollten.

## DoubleActivityNotOverBreaks

Dieser Constraint sollte höchstens einmal vorkommen. Eine Doppelstunde soll nicht durch eine Pause unterbrochen werden. Die Pause sind unmittelbar vor den angegebenen Stunden.

```
{
	"Constraint":           "DOUBLE_LESSON_NOT_OVER_BREAKS",
	"Weight":               90,
	"Hours":                [2, 4]
}
```

## MinHoursFollowing

Zwischen den Stunden des ersten Kurses und denen des zweiten Kurses sollten (an einem Tag) mindestens die angegebene Zahl an Stunden liegen (die Stunden der zwei Kurse werden nicht mitgezählt). 

```
{
	"Constraint":   "MIN_HOURS_FOLLOWING",
	"Weight":       90,
	"Course1":      "2edfe663-c62b-4d05-ace2-0bedb0f4b672",
	"Course2":      "5fda67de-bbb3-48a2-a098-d957796b7743",
	"Hours":        4
}
```
