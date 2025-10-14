package clocker

import (
	"log"
	"testing"
)

func TestCalculate(t *testing.T) {
	tests := []*calcTest{
		newCalcTest("5 minute surplus", true, `
			--- begin ---
			Flex: 89m
		
			08:15 - Start
			11:35 - Break
		
			11:45 - Back
			14:35 - Break
		`).expectSurplus("9m"),
		newCalcTest("50 minutes remaining", true, `
			08:00 - Start
			14:40 - Stop
		`).expectTimeLeft("50m"),
		newCalcTest("full day at 16:10", true, `
			08:15 - Start
			11:35 - Pause
		
			11:45 - Resume
			14:35 - Pause

			14:50 - Resume
		`).expectTimeLeft("80m").
			expectFullDay("16:10"),
		newCalcTest("altered target", true, `
			Target: 4h (semi-holiday)

			06:22 - Start
		`).expectTimeLeft("4h").
			expectFullDay("10:22"),
		newCalcTest("end before start", false, `
			07:06 - On
			06:50 - Off
		`),
		newCalcTest("double start", false, `
			04:30 - On
			06:30 - On
		`),
		newCalcTest("start with end", false, `19:06 - end`),
		newCalcTest("no valid entries", false, "\n\n\n\t ---"),
		newCalcTest("tasks with categories", true, `
			-- sunday 25.05.1975
			08:02 - Debate: Can swallows carry coconuts?
			08:12 - Fetch: Excalibur
			12:00 - Pause

			12:15 - Combat: The Black Knight
			12:30 - Meeting: Witch Trial
			12:36 - Travel: Camelot
			13:37 - Travel: Occupied Castle
			14:24 - Debate: The French Taunter
			14:44 - Travel: Separate ways
			15:10 - Business: The Knights who until recently said Ni
			15:30 - Travel: Sacred Cave
			15:55 - Combat: The Rabbit of Caerbannog
			16:00 - Travel: Bridge of Death
			16:45 - Meeting: Bridgekeeper
			16:50 - Travel: Castle Aarrgh
			17:30 - Done
		`).expectCategory("travel", "4h 4m").expectCategory("combat", "20m").expectCategory("debate", "30m").expectDate(25, 05, 1975),
	}

	lp := LogParser{}
	err := lp.Init("")
	if err != nil {
		t.Errorf("lp.Init: %s", err.Error())
		return
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entries := lp.parse(test.input)
			calcResult := lp.calculate(entries)
			if calcResult.valid != test.expect.valid {
				t.Errorf("validation mismatch: expected %t, got %t", test.expect.valid, calcResult.valid)
				if !calcResult.valid {
					t.Errorf("validation error: %s", calcResult.validationMsg)
				}
				return
			}

			if test.expect.surplus != nil {
				if calcResult.surplus == nil {
					t.Errorf("expected a surplus, but did not get one")
					return
				}
				if *test.expect.surplus != *calcResult.surplus {
					t.Errorf("surplus mismatch: expected %s, got %s", test.expect.surplus.String(), calcResult.surplus.String())
					return
				}
			}

			if test.expect.timeLeft != nil {
				if calcResult.timeLeft == nil {
					t.Errorf("expected timeLeft, but got nil")
					return
				}

				if *test.expect.timeLeft != *calcResult.timeLeft {
					t.Errorf("timeLeft mismatch: expected %s, got %s", test.expect.timeLeft.String(), calcResult.timeLeft.String())
					return
				}
			}

			if test.expect.fullDayAt != nil {
				if calcResult.fullDayAt == nil {
					t.Errorf("expected fullDay, but got nil")
					return
				}

				if !test.expect.fullDayAt.Equal(*calcResult.fullDayAt) {
					t.Errorf("fullDay mismatch - expected %s, got %s", formatTimestamp(*test.expect.fullDayAt), formatTimestamp(*calcResult.fullDayAt))
				}
			}

			for _, ec := range test.expect.categories {
				found := false
				for _, ac := range calcResult.categories {
					if ec.matchName(ac.name) {
						found = true
						if ec.timeWorked != ac.timeWorked {
							t.Errorf("category %q duration mismatch: expected %s, got %s", ec.name, ec.timeWorked.String(), ac.timeWorked.String())
						}
						break
					}
				}
				if !found {
					t.Errorf("expected category %q not found in results", ec.name)
				}
			}

			if test.expect.date != nil {
				if calcResult.date == nil {
					t.Errorf("expected date, but got nil")
					return
				}

				dayMatch := test.expect.date.day == calcResult.date.day
				monthMatch := test.expect.date.month == calcResult.date.month
				yearMatch := test.expect.date.year == calcResult.date.year

				if !dayMatch || !monthMatch || !yearMatch {
					t.Errorf("date mismatch: expected %02d.%02d.%04d, got %02d.%02d.%04d",
						test.expect.date.day, test.expect.date.month, test.expect.date.year,
						calcResult.date.day, calcResult.date.month, calcResult.date.year)
				}
			}
		})
	}
}

type calcTest struct {
	name   string
	input  string
	expect calcResult
}

func newCalcTest(name string, valid bool, input string) *calcTest {
	return &calcTest{
		name:  name,
		input: input,
		expect: calcResult{
			valid: valid,
		},
	}
}

func (ct *calcTest) expectTimeWorked(tw string) *calcTest {
	twd, err := parseDuration(tw)
	if err != nil {
		log.Panicf(`failed to parse duration string "%s": %s`, tw, err.Error())
	}
	ct.expect.timeWorked = twd
	return ct
}

func (ct *calcTest) expectTimeLeft(tl string) *calcTest {
	tld, err := parseDuration(tl)
	if err != nil {
		log.Panicf(`failed to parse duration string "%s": %s`, tl, err.Error())
	}
	ct.expect.timeLeft = &tld
	return ct
}

func (ct *calcTest) expectFullDay(ts string) *calcTest {
	fd, err := parseTimestamp(ts)
	if err != nil {
		log.Panicf(`failed to parse time string "%s": %s`, ts, err.Error())
	}
	ct.expect.fullDayAt = &fd
	return ct
}

func (ct *calcTest) expectSurplus(sus string) *calcTest {
	sur, err := parseDuration(sus)
	if err != nil {
		log.Panicf(`failed to parse duration string "%s": %s`, sus, err.Error())
	}
	ct.expect.surplus = &sur
	return ct
}

func (ct *calcTest) expectCategory(name string, dur string) *calcTest {
	d, err := parseDuration(dur)
	if err != nil {
		log.Panicf(`failed to parse duration string "%s": %s`, dur, err.Error())
	}
	ct.expect.categories = append(ct.expect.categories, resultCategory{
		name:       name,
		timeWorked: d,
	})
	return ct
}

func (ct *calcTest) expectDate(day, month, year int) *calcTest {
	ct.expect.date = &logDate{
		day:   day,
		month: month,
		year:  year,
	}
	return ct
}
