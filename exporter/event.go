package exporter

import "time"

// Event represents a time tracking event for exporting purposes.
type Event struct {
	Hours    int
	Minutes  int
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
