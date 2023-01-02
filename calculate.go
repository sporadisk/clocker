package clocker

import (
	"fmt"
	"time"
)

type calcResult struct {
	valid         bool
	validationMsg string
	timeWorked    time.Duration
	timeLeft      *time.Duration
	surplus       *time.Duration
	fullDayAt     *time.Time
}

func (l *LogParser) calculate(entries []clockEntry) calcResult {
	state := stateInit
	prevCommand := "--start of document--"
	lastOn := time.Time{}
	lastOff := time.Time{}
	durations := []time.Duration{}

	if len(entries) == 0 {
		return calcResult{
			valid:         false,
			validationMsg: "No valid time entries detected.",
		}
	}

	for _, entry := range entries {
		if entry.action == actionClockIn {
			if state == stateOn {
				return calcResult{
					valid:         false,
					validationMsg: fmt.Sprintf(`Duplicate clock-in on line %d`, entry.lineNumber),
				}
			}

			if !entry.timestamp.After(lastOff) && !lastOff.IsZero() {
				return calcResult{
					valid:         false,
					validationMsg: fmt.Sprintf(`Clock-in on line %d occurs prior to the previous clock-out`, entry.lineNumber),
				}
			}

			lastOn = *entry.timestamp
			state = stateOn
		}

		if entry.action == actionClockOut {
			if state != stateOn {
				return calcResult{
					valid:         false,
					validationMsg: fmt.Sprintf(`Clock-out on line %d follows "%s", should follow a clock-in`, entry.lineNumber, prevCommand),
				}
			}

			if !entry.timestamp.After(lastOn) {
				return calcResult{
					valid:         false,
					validationMsg: fmt.Sprintf(`Clock-out on line %d has an earlier timestamp than its corresponding "%s"`, entry.lineNumber, prevCommand),
				}
			}

			lastOff = *entry.timestamp
			state = stateOff
			durations = append(durations, entry.timestamp.Sub(lastOn))
		}

		if entry.action == actionFlex {
			if state == stateOn {
				return calcResult{
					valid:         false,
					validationMsg: fmt.Sprintf(`Flex time entry on line %d follows a clock-in, which is wrong.`, entry.lineNumber),
				}
			}

			state = stateFlex
			durations = append(durations, *entry.duration)
		}

		if entry.action == actionTarget {
			if state == stateOn {
				return calcResult{
					valid:         false,
					validationMsg: fmt.Sprintf(`"%s" entry on line %d follows a clock-in, which is wrong.`, entry.command, entry.lineNumber),
				}
			}

			state = stateTarget
			l.fullDay = *entry.duration
		}

		prevCommand = entry.command
	}

	// validation and duration collection complete: Time to calculate
	res := calcResult{
		valid: true,
	}

	sumDurations := time.Duration(0)

	for _, d := range durations {
		sumDurations = sumDurations + d
	}
	res.timeWorked = sumDurations

	if sumDurations < l.fullDay {
		timeLeft := l.fullDay - sumDurations
		res.timeLeft = &timeLeft

		if state == stateOn {
			fullDayAt := lastOn.Add(timeLeft)
			res.fullDayAt = &fullDayAt
		}
	}

	if sumDurations > l.fullDay {
		surplus := sumDurations - l.fullDay
		res.surplus = &surplus
	}

	return res
}
