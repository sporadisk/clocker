package logfile

import (
	"log"
	"testing"
	"time"

	"github.com/sporadisk/clocker/logentry"
)

func TestParseLine(t *testing.T) {
	tests := []*testLine{
		newTestLine("--- thursday ---").expectInvalid(),
		newTestLine("Flex: 1h 20m").expectAction("flex").expectDuration("80m"),
		newTestLine("10:00 - Start").expectAction("on").expectTimestamp("10:00:00"),
		newTestLine("10:21 - Break").expectAction("off").expectTimestamp("10:21:00"),
		newTestLine("Workday: 8h").expectAction("target").expectDuration("480m"),
		newTestLine("Format: m").expectAction("outputformat"),
	}

	lp := LogParser{}
	err := lp.Init()
	if err != nil {
		t.Errorf("lp.Init: %s", err.Error())
		return
	}

	for i, te := range tests {
		t.Run(te.line, func(t *testing.T) {
			valid, entry := lp.parseLine(te.line, i)
			if valid != te.valid {
				t.Errorf("validation mismatch: Expected %t, got %t", te.valid, valid)
				return
			}

			if !valid {
				return
			}

			if entry.Action != te.expected.Action {
				t.Errorf("action mismatch: expected %s, got %s", te.expected.Action, entry.Action)
				return
			}

			if te.expected.Duration != nil {
				if entry.Duration == nil {
					t.Errorf("expected a duration, but it was nil")
					return
				}
				if *entry.Duration != *te.expected.Duration {
					t.Errorf("duration mismatch: expected %f, got %f", te.expected.Duration.Minutes(), entry.Duration.Minutes())
					return
				}
			} else {
				if entry.Duration != nil {
					t.Errorf("expected duration to be nil, but it was %f", entry.Duration.Minutes())
					return
				}
			}

			if te.expected.Timestamp != nil {
				if entry.Timestamp == nil {
					t.Errorf("expected a timestamp, but it was nil")
					return
				}
				if !entry.Timestamp.Equal(*te.expected.Timestamp) {
					t.Errorf("timestamp mismatch!\nExpected: %s\nGot     : %s",
						te.expected.Timestamp.Format(time.RFC3339),
						entry.Timestamp.Format(time.RFC3339))
					return
				}
			} else {
				if entry.Timestamp != nil {
					t.Errorf("expected timestamp to be nil, but it was %s", entry.Timestamp.Format(time.RFC3339))
					return
				}
			}
		})
	}
}

func newTestLine(line string) *testLine {
	return &testLine{
		line:  line,
		valid: true,
		expected: logentry.Entry{
			Action: "",
		},
	}
}

type testLine struct {
	line     string
	valid    bool
	expected logentry.Entry
}

func (tl *testLine) expectInvalid() *testLine {
	tl.valid = false
	return tl
}

func (tl *testLine) expectAction(action string) *testLine {
	tl.expected.Action = action
	return tl
}

func (tl *testLine) expectTimestamp(ts string) *testLine {
	timestamp, err := time.Parse("15:04:05", ts)
	if err != nil {
		log.Printf("time parse error: %s", err.Error())
		return tl
	}
	tl.expected.Timestamp = &timestamp
	return tl
}

func (tl *testLine) expectDuration(ds string) *testLine {
	duration, err := time.ParseDuration(ds)
	if err != nil {
		log.Printf("duration parse error: %s", err.Error())
		return tl
	}
	tl.expected.Duration = &duration
	return tl
}
