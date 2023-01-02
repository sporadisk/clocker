package clocker

import (
	"log"
	"testing"
	"time"
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
	err := lp.Init("")
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

			if entry.action != te.expected.action {
				t.Errorf("action mismatch: expected %s, got %s", te.expected.action, entry.action)
				return
			}

			if te.expected.duration != nil {
				if entry.duration == nil {
					t.Errorf("expected a duration, but it was nil")
					return
				}
				if *entry.duration != *te.expected.duration {
					t.Errorf("duration mismatch: expected %f, got %f", te.expected.duration.Minutes(), entry.duration.Minutes())
					return
				}
			} else {
				if entry.duration != nil {
					t.Errorf("expected duration to be nil, but it was %f", entry.duration.Minutes())
					return
				}
			}

			if te.expected.timestamp != nil {
				if entry.timestamp == nil {
					t.Errorf("expected a timestamp, but it was nil")
					return
				}
				if !entry.timestamp.Equal(*te.expected.timestamp) {
					t.Errorf("timestamp mismatch!\nExpected: %s\nGot     : %s",
						te.expected.timestamp.Format(time.RFC3339),
						entry.timestamp.Format(time.RFC3339))
					return
				}
			} else {
				if entry.timestamp != nil {
					t.Errorf("expected timestamp to be nil, but it was %s", entry.timestamp.Format(time.RFC3339))
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
		expected: clockEntry{
			action: "",
		},
	}
}

type testLine struct {
	line     string
	valid    bool
	expected clockEntry
}

func (tl *testLine) expectInvalid() *testLine {
	tl.valid = false
	return tl
}

func (tl *testLine) expectAction(action string) *testLine {
	tl.expected.action = action
	return tl
}

func (tl *testLine) expectTimestamp(ts string) *testLine {
	timestamp, err := time.Parse("15:04:05", ts)
	if err != nil {
		log.Printf("time parse error: %s", err.Error())
		return tl
	}
	tl.expected.timestamp = &timestamp
	return tl
}

func (tl *testLine) expectDuration(ds string) *testLine {
	duration, err := time.ParseDuration(ds)
	if err != nil {
		log.Printf("duration parse error: %s", err.Error())
		return tl
	}
	tl.expected.duration = &duration
	return tl
}
