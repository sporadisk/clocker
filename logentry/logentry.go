package logentry

import "time"

const (
	ActionClockIn      = "on"
	ActionClockOut     = "off"
	ActionStartTask    = "starttask"
	ActionFlex         = "flex"
	ActionTarget       = "target"
	ActionOutputFormat = "outputformat"
	ActionSetDay       = "setday"
)

type Entry struct {
	Action     string // the action to perform based on the interpretation of the command
	Command    string // the actual command used on the original line
	Task       string // optional task name
	Timestamp  *time.Time
	Duration   *time.Duration
	LineNumber int
	DayName    string
	Day        int
	Month      int
	Year       int
}

type Receiver interface {
	Receive(entries []Entry) error
}

type Subscriber interface {
	Subscribe(receiver Receiver) error
}
