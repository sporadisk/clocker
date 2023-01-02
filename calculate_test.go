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
