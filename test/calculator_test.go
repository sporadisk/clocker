package test

import (
	"log"
	"testing"

	"github.com/sporadisk/clocker/calculator"
	"github.com/sporadisk/clocker/client/logfile"
	"github.com/sporadisk/clocker/format"
	"github.com/sporadisk/clocker/summary"
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

	lp := logfile.LogParser{}
	err := lp.Init()
	if err != nil {
		t.Errorf("lp.Init: %s", err.Error())
		return
	}

	defaultFullDay, err := format.ParseDuration("450m")
	if err != nil {
		t.Errorf("failed to parse default full day duration: %s", err.Error())
		return
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			entries := lp.Parse(test.input)
			calcResult := calculator.LogSum(entries, defaultFullDay)
			if calcResult.Valid != test.expect.Valid {
				t.Errorf("validation mismatch: expected %t, got %t", test.expect.Valid, calcResult.Valid)
				if !calcResult.Valid {
					t.Errorf("validation error: %s", calcResult.ValidationMsg)
				}
				return
			}

			if test.expect.Surplus != nil {
				if calcResult.Surplus == nil {
					t.Errorf("expected a surplus, but did not get one")
					return
				}
				if *test.expect.Surplus != *calcResult.Surplus {
					t.Errorf("surplus mismatch: expected %s, got %s", test.expect.Surplus.String(), calcResult.Surplus.String())
					return
				}
			}

			if test.expect.TimeLeft != nil {
				if calcResult.TimeLeft == nil {
					t.Errorf("expected timeLeft, but got nil")
					return
				}

				if *test.expect.TimeLeft != *calcResult.TimeLeft {
					t.Errorf("timeLeft mismatch: expected %s, got %s", test.expect.TimeLeft.String(), calcResult.TimeLeft.String())
					return
				}
			}

			if test.expect.FullDayAt != nil {
				if calcResult.FullDayAt == nil {
					t.Errorf("expected fullDay, but got nil")
					return
				}

				if !test.expect.FullDayAt.Equal(*calcResult.FullDayAt) {
					t.Errorf("fullDay mismatch - expected %s, got %s", format.Timestamp(*test.expect.FullDayAt), format.Timestamp(*calcResult.FullDayAt))
				}
			}

			for _, ec := range test.expect.Categories {
				found := false
				for _, ac := range calcResult.Categories {
					if ec.MatchName(ac.Name) {
						found = true
						if ec.TimeWorked != ac.TimeWorked {
							t.Errorf("category %q duration mismatch: expected %s, got %s", ec.Name, ec.TimeWorked.String(), ac.TimeWorked.String())
						}
						break
					}
				}
				if !found {
					t.Errorf("expected category %q not found in results", ec.Name)
				}
			}

			if test.expect.Date != nil {
				if calcResult.Date == nil {
					t.Errorf("expected date, but got nil")
					return
				}

				dayMatch := test.expect.Date.Day == calcResult.Date.Day
				monthMatch := test.expect.Date.Month == calcResult.Date.Month
				yearMatch := test.expect.Date.Year == calcResult.Date.Year

				if !dayMatch || !monthMatch || !yearMatch {
					t.Errorf("date mismatch: expected %02d.%02d.%04d, got %02d.%02d.%04d",
						test.expect.Date.Day, test.expect.Date.Month, test.expect.Date.Year,
						calcResult.Date.Day, calcResult.Date.Month, calcResult.Date.Year)
				}
			}
		})
	}
}

type calcTest struct {
	name   string
	input  string
	expect summary.Summary
}

func newCalcTest(name string, valid bool, input string) *calcTest {
	return &calcTest{
		name:  name,
		input: input,
		expect: summary.Summary{
			Valid: valid,
		},
	}
}

func (ct *calcTest) expectTimeWorked(tw string) *calcTest {
	twd, err := format.ParseDuration(tw)
	if err != nil {
		log.Panicf(`failed to parse duration string "%s": %s`, tw, err.Error())
	}
	ct.expect.TimeWorked = twd
	return ct
}

func (ct *calcTest) expectTimeLeft(tl string) *calcTest {
	tld, err := format.ParseDuration(tl)
	if err != nil {
		log.Panicf(`failed to parse duration string "%s": %s`, tl, err.Error())
	}
	ct.expect.TimeLeft = &tld
	return ct
}

func (ct *calcTest) expectFullDay(ts string) *calcTest {
	fd, err := format.ParseTimestamp(ts)
	if err != nil {
		log.Panicf(`failed to parse time string "%s": %s`, ts, err.Error())
	}
	ct.expect.FullDayAt = &fd
	return ct
}

func (ct *calcTest) expectSurplus(sus string) *calcTest {
	sur, err := format.ParseDuration(sus)
	if err != nil {
		log.Panicf(`failed to parse duration string "%s": %s`, sus, err.Error())
	}
	ct.expect.Surplus = &sur
	return ct
}

func (ct *calcTest) expectCategory(name string, dur string) *calcTest {
	d, err := format.ParseDuration(dur)
	if err != nil {
		log.Panicf(`failed to parse duration string "%s": %s`, dur, err.Error())
	}
	ct.expect.Categories = append(ct.expect.Categories, summary.ResultCategory{
		Name:       name,
		TimeWorked: d,
	})
	return ct
}

func (ct *calcTest) expectDate(day, month, year int) *calcTest {
	ct.expect.Date = &summary.Date{
		Day:   day,
		Month: month,
		Year:  year,
	}
	return ct
}
