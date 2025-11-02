package event

import (
	"math"
	"time"
)

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

func (e *Event) DetermineHours() {
	if e.Start.IsZero() || e.End.IsZero() {
		e.Hours = 0
		e.Minutes = 0
		return
	}

	duration := e.End.Sub(e.Start)
	totalMinutes := math.Floor(duration.Minutes())

	e.Hours = int(math.Floor(totalMinutes / 60))
	e.Minutes = int(totalMinutes) % 60
}

type EventDate struct {
	Day   int
	Month int
	Year  int
}
