package event

import "time"

// Event represents a time tracking event for the purposes of output to other
// systems, such as Timely.
type Event struct {
	Hours    int
	Minutes  int
	Date     EventDate
	Start    time.Time
	End      time.Time
	Category string
	Task     string
}

type EventDate struct {
	Day   int
	Month int
	Year  int
}
