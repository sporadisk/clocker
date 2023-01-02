package clocker

import "testing"

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
	err := lp.Init("")
	if err != nil {
		t.Errorf("lp.Init: %s", err.Error())
		return
	}

	entries := lp.parse(testText)
	//t.Logf("entries: %#v", entries)

	if len(entries) != len(expected) {
		t.Errorf("entry length mismatch: expected %d, got %d", len(expected), len(entries))
		return
	}

	for i, entry := range entries {
		if entry.action != expected[i] {
			t.Errorf("wrong action for entry %d: expected %s, got %s", i, expected[i], entry.action)
		}
	}
}
