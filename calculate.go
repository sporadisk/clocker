package clocker

import (
	"fmt"
	"strings"
	"time"
)

const timestampFormat = "15:04"

type calculator struct {
	logState         string
	lastOn           time.Time
	lastOff          time.Time
	durations        []time.Duration
	taskCatDurations map[string]time.Duration
	prevCommand      string
	currentCategory  string
	currentTask      string
	currentDate      logDate
}

type calcResult struct {
	valid         bool
	validationMsg string
	timeWorked    time.Duration
	timeLeft      *time.Duration
	surplus       *time.Duration
	fullDayAt     *time.Time
	categories    []resultCategory
	date          *logDate
}

type logDate struct {
	dayName string
	day     int
	month   int
	year    int
}

func (cr *calcResult) addCategory(name string, dur time.Duration) {
	for i, c := range cr.categories {
		if c.matchName(name) {
			cr.categories[i].timeWorked += dur
			return
		}
	}

	cr.categories = append(cr.categories, resultCategory{
		name:       name,
		timeWorked: dur,
	})
}

type resultCategory struct {
	name       string
	timeWorked time.Duration
}

func (rc *resultCategory) matchName(name string) bool {
	return strings.EqualFold(name, rc.name)
}

func (l *LogParser) calculate(entries []clockEntry) calcResult {

	calc := calculator{
		logState:         stateInit,
		lastOn:           time.Time{},
		lastOff:          time.Time{},
		durations:        []time.Duration{},
		taskCatDurations: map[string]time.Duration{},
		prevCommand:      "--start of document--",
	}

	if len(entries) == 0 {
		return calcResult{
			valid:         false,
			validationMsg: "No valid time entries detected.",
		}
	}

	for _, entry := range entries {
		if entry.action == actionClockIn {
			success, res := calc.clockIn(entry)
			if !success {
				return res
			}
		}

		if entry.action == actionStartTask {
			category, task := l.parseCategoryAndTask(entry.task)
			success, res := calc.startTask(entry, category, task)
			if !success {
				return res
			}
		}

		if entry.action == actionClockOut {
			success, res := calc.clockOut(entry)
			if !success {
				return res
			}
		}

		if entry.action == actionFlex {
			success, res := calc.flex(entry)
			if !success {
				return res
			}
		}

		if entry.action == actionTarget {
			if calc.logState == stateOn {
				return calcResult{
					valid:         false,
					validationMsg: fmt.Sprintf(`"%s" entry on line %d follows a clock-in, which is wrong.`, entry.command, entry.lineNumber),
				}
			}

			calc.logState = stateTarget
			l.fullDay = *entry.duration
		}

		if entry.action == actionSetDay {
			calc.currentDate = logDate{
				dayName: entry.dayName,
				day:     entry.day,
				month:   entry.month,
				year:    entry.year,
			}
		}

		calc.prevCommand = entry.command
	}

	// validation and duration collection complete: Time to calculate
	return calc.sumResults(l.fullDay)
}

func (c *calculator) clockOut(entry clockEntry) (success bool, result calcResult) {
	if c.logState != stateOn {
		return false, calcResult{
			valid:         false,
			validationMsg: fmt.Sprintf(`Clock-out at %s on line %d follows "%s", should follow a clock-in`, entry.timestamp.Format(timestampFormat), entry.lineNumber, c.logState),
		}
	}

	if !entry.timestamp.After(c.lastOn) {
		return false, calcResult{
			valid:         false,
			validationMsg: fmt.Sprintf(`Clock-out on line %d has an earlier timestamp than its corresponding "%s"`, entry.lineNumber, c.prevCommand),
		}
	}

	c.lastOff = *entry.timestamp
	c.logState = stateOff
	c.durations = append(c.durations, entry.timestamp.Sub(c.lastOn))

	if c.currentCategory != "" {
		c.addToCategory(c.currentCategory, entry.timestamp.Sub(c.lastOn))
	}

	// reset current task and category
	c.currentTask = ""
	c.currentCategory = ""

	return true, calcResult{}
}

func (c *calculator) addToCategory(cat string, dur time.Duration) {
	if cat == "" {
		return
	}

	_, ok := c.taskCatDurations[cat]
	if !ok {
		c.taskCatDurations[cat] = 0
	}

	c.taskCatDurations[cat] += dur
}

func (c *calculator) clockIn(entry clockEntry) (success bool, result calcResult) {
	if c.logState == stateOn {
		return false, calcResult{
			valid:         false,
			validationMsg: fmt.Sprintf(`Duplicate clock-in on line %d`, entry.lineNumber),
		}
	}

	if entry.timestamp.Before(c.lastOff) && !c.lastOff.IsZero() {
		return false, calcResult{
			valid: false,
			validationMsg: fmt.Sprintf(`Clock-in at %s on line %d occurs prior to the previous clock-out (%s)`,
				entry.timestamp.Format(timestampFormat),
				entry.lineNumber,
				c.lastOff.Format(timestampFormat),
			),
		}
	}

	c.lastOn = *entry.timestamp
	c.logState = stateOn

	return true, calcResult{}
}

func (c *calculator) startTask(entry clockEntry, category, task string) (success bool, result calcResult) {
	// We're already clocked in, probably on another task
	// Clock out of the previous task first
	if c.logState == stateOn {
		success, res := c.clockOut(entry)
		if !success {
			return false, res
		}
	}

	c.currentTask = task
	c.currentCategory = category

	return c.clockIn(entry)
}

func (c *calculator) flex(entry clockEntry) (success bool, result calcResult) {
	if c.logState == stateOn {
		return false, calcResult{
			valid:         false,
			validationMsg: fmt.Sprintf(`Flex time entry on line %d follows a clock-in, which is wrong.`, entry.lineNumber),
		}
	}

	c.logState = stateFlex
	c.durations = append(c.durations, *entry.duration)
	return true, calcResult{}
}

func (c *calculator) sumResults(fullDay time.Duration) calcResult {
	res := calcResult{
		valid: true,
	}

	sumDurations := time.Duration(0)

	for _, d := range c.durations {
		sumDurations = sumDurations + d
	}
	res.timeWorked = sumDurations

	if sumDurations < fullDay {
		timeLeft := fullDay - sumDurations
		res.timeLeft = &timeLeft

		if c.logState == stateOn {
			fullDayAt := c.lastOn.Add(timeLeft)
			res.fullDayAt = &fullDayAt
		}
	}

	if sumDurations > fullDay {
		surplus := sumDurations - fullDay
		res.surplus = &surplus
	}

	for cat, dur := range c.taskCatDurations {
		res.addCategory(cat, dur)
	}

	if c.currentDate.day != 0 && c.currentDate.month != 0 {
		res.date = &logDate{
			dayName: c.currentDate.dayName,
			day:     c.currentDate.day,
			month:   c.currentDate.month,
			year:    c.currentDate.year,
		}
	}

	return res
}

// Formats:
// 19:00 - Category
// 19:00 - Category: Task
func (l *LogParser) parseCategoryAndTask(s string) (category, task string) {
	if s == "" {
		return "", ""
	}

	// todo: Add support for predefined categories
	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		category = strings.ToLower(strings.TrimSpace(parts[0]))
		task = strings.ToLower(strings.TrimSpace(parts[1]))
		return category, task
	}

	category = strings.ToLower(strings.TrimSpace(s))
	return category, ""
}
