package logfile

import (
	"testing"

	"github.com/sporadisk/clocker/logentry"
)

func TestParse(t *testing.T) {
	testText := `
	--- begin ---
	Full day: 7h 30m
	Flex: 2h 9m

	06:08 - Start
	09:46 - Break

	10:01 - Back
	14:35 - Break

	15:29 - Back
	18:30 - End
	`
	expected := []string{"target", "flex", "on", "off", "on", "off", "on", "off"}

	lp := LogParser{}
	err := lp.Init()
	if err != nil {
		t.Errorf("lp.Init: %s", err.Error())
		return
	}

	entries := lp.Parse(testText)
	//t.Logf("entries: %#v", entries)

	if len(entries) != len(expected) {
		t.Errorf("entry length mismatch: expected %d, got %d", len(expected), len(entries))
		return
	}

	for i, entry := range entries {
		if entry.Action != expected[i] {
			t.Errorf("wrong action for entry %d: expected %s, got %s", i, expected[i], entry.Action)
		}
	}
}

func TestParseCategories(t *testing.T) {
	testText := `
	-- monday 09.08
	Full day: 7h 30m

	06:00 - Start
	06:08 - Cows
	09:46 - Sheep
	10:01 - Chickens
	12:01 - Break

	14:35 - Corn
	15:29 - Wheat
	18:30 - End
	`

	expectedDay := 9
	expectedMonth := 8
	expectedTasks := []string{"", "", "", "Cows", "Sheep", "Chickens", "", "Corn", "Wheat", ""}
	expectedActions := []string{
		logentry.ActionSetDay,
		logentry.ActionTarget,
		logentry.ActionClockIn,
		logentry.ActionStartTask,
		logentry.ActionStartTask,
		logentry.ActionStartTask,
		logentry.ActionClockOut,

		logentry.ActionStartTask,
		logentry.ActionStartTask,
		logentry.ActionClockOut,
	}

	if len(expectedTasks) != len(expectedActions) {
		t.Errorf("test setup error: expectedTasks and expectedActions length mismatch")
		return
	}

	lp := LogParser{}
	err := lp.Init()
	if err != nil {
		t.Errorf("lp.Init: %s", err.Error())
		return
	}

	entries := lp.Parse(testText)

	if len(entries) != len(expectedTasks) {
		t.Errorf("entry length mismatch: expected %d, got %d", len(expectedTasks), len(entries))
		return
	}

	gotDay := false

	for i, entry := range entries {

		if entry.Action == logentry.ActionSetDay {
			gotDay = true
			if entry.Day != expectedDay {
				t.Errorf("wrong day: expected %d, got %d", expectedDay, entry.Day)
			}
			if entry.Month != expectedMonth {
				t.Errorf("wrong month: expected %d, got %d", expectedMonth, entry.Month)
			}
		}

		if entry.Task != expectedTasks[i] {
			t.Errorf("wrong task for entry %d: expected %q, got %q", i, expectedTasks[i], entry.Task)
		}

		if entry.Action != expectedActions[i] {
			t.Errorf("wrong action for entry %d: expected %q, got %q", i, expectedActions[i], entry.Action)
		}
	}

	if !gotDay {
		t.Errorf("did not get day entry")
	}
}
