package calculator

import (
	"fmt"
	"strings"
	"time"

	"github.com/sporadisk/clocker/event"
	"github.com/sporadisk/clocker/logentry"
	"github.com/sporadisk/clocker/summary"
)

const (
	timestampFormat = "15:04"
	stateInit       = "init"
	stateOn         = "on"
	stateOff        = "off"
	stateFlex       = "flex"
	stateTarget     = "target"
)

type logSummary struct {
	logState         string
	lastOn           time.Time
	lastOff          time.Time
	durations        []time.Duration
	taskCatDurations map[string]time.Duration
	prevCommand      string
	currentCategory  string
	currentTask      string
	currentDate      summary.Date
	fullDay          time.Duration
	events           []event.Event
}

func LogSum(entries []logentry.Entry, fullDay time.Duration) summary.Summary {

	sum := logSummary{
		logState:         stateInit,
		lastOn:           time.Time{},
		lastOff:          time.Time{},
		durations:        []time.Duration{},
		taskCatDurations: map[string]time.Duration{},
		prevCommand:      "--start of document--",
		fullDay:          fullDay,
	}

	if len(entries) == 0 {
		return summary.Summary{
			Valid:         false,
			ValidationMsg: "No valid time entries detected.",
		}
	}

	for _, entry := range entries {
		if entry.Action == logentry.ActionClockIn {
			success, res := sum.clockIn(entry)
			if !success {
				return res
			}
		}

		if entry.Action == logentry.ActionStartTask {
			category, task := parseCategoryAndTask(entry.Task)
			success, res := sum.startTask(entry, category, task)
			if !success {
				return res
			}
		}

		if entry.Action == logentry.ActionClockOut {
			success, res := sum.clockOut(entry)
			if !success {
				return res
			}
		}

		if entry.Action == logentry.ActionFlex {
			success, res := sum.flex(entry)
			if !success {
				return res
			}
		}

		if entry.Action == logentry.ActionTarget {
			if sum.logState == stateOn {
				return summary.Summary{
					Valid:         false,
					ValidationMsg: fmt.Sprintf(`"%s" entry on line %d follows a clock-in, which is wrong.`, entry.Command, entry.LineNumber),
				}
			}

			sum.logState = stateTarget
			sum.fullDay = *entry.Duration
		}

		if entry.Action == logentry.ActionSetDay {
			sum.currentDate = summary.Date{
				DayName: entry.DayName,
				Day:     entry.Day,
				Month:   entry.Month,
				Year:    entry.Year,
			}
		}

		sum.prevCommand = entry.Command
	}

	// validation and duration collection complete: Time to calculate
	return sum.summarize()
}

func (c *logSummary) clockOut(entry logentry.Entry) (success bool, result summary.Summary) {
	if c.logState != stateOn {
		return false, summary.Summary{
			Valid:         false,
			ValidationMsg: fmt.Sprintf(`Clock-out at %s on line %d follows "%s", should follow a clock-in`, entry.Timestamp.Format(timestampFormat), entry.LineNumber, c.logState),
		}
	}

	if !entry.Timestamp.After(c.lastOn) {
		return false, summary.Summary{
			Valid:         false,
			ValidationMsg: fmt.Sprintf(`Clock-out on line %d has an earlier timestamp than its corresponding "%s"`, entry.LineNumber, c.prevCommand),
		}
	}

	c.lastOff = *entry.Timestamp
	c.logState = stateOff
	c.durations = append(c.durations, entry.Timestamp.Sub(c.lastOn))

	if c.currentCategory != "" {
		c.addToCategory(c.currentCategory, entry.Timestamp.Sub(c.lastOn))
	}

	// reset current task and category
	c.currentTask = ""
	c.currentCategory = ""

	return true, summary.Summary{}
}

func (c *logSummary) addToCategory(cat string, dur time.Duration) {
	if cat == "" {
		return
	}

	_, ok := c.taskCatDurations[cat]
	if !ok {
		c.taskCatDurations[cat] = 0
	}

	c.taskCatDurations[cat] += dur
}

func (c *logSummary) clockIn(entry logentry.Entry) (success bool, result summary.Summary) {
	if c.logState == stateOn {
		return false, summary.Summary{
			Valid:         false,
			ValidationMsg: fmt.Sprintf(`Duplicate clock-in on line %d`, entry.LineNumber),
		}
	}

	if entry.Timestamp.Before(c.lastOff) && !c.lastOff.IsZero() {
		return false, summary.Summary{
			Valid: false,
			ValidationMsg: fmt.Sprintf(`Clock-in at %s on line %d occurs prior to the previous clock-out (%s)`,
				entry.Timestamp.Format(timestampFormat),
				entry.LineNumber,
				c.lastOff.Format(timestampFormat),
			),
		}
	}

	c.lastOn = *entry.Timestamp
	c.logState = stateOn

	return true, summary.Summary{}
}

func (c *logSummary) startTask(entry logentry.Entry, category, task string) (success bool, result summary.Summary) {
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

func (c *logSummary) flex(entry logentry.Entry) (success bool, result summary.Summary) {
	if c.logState == stateOn {
		return false, summary.Summary{
			Valid:         false,
			ValidationMsg: fmt.Sprintf(`Flex time entry on line %d follows a clock-in, which is wrong.`, entry.LineNumber),
		}
	}

	c.logState = stateFlex
	c.durations = append(c.durations, *entry.Duration)
	return true, summary.Summary{}
}

func (c *logSummary) summarize() summary.Summary {
	res := summary.Summary{
		Valid: true,
	}

	sumDurations := time.Duration(0)

	for _, d := range c.durations {
		sumDurations = sumDurations + d
	}
	res.TimeWorked = sumDurations

	if sumDurations < c.fullDay {
		timeLeft := c.fullDay - sumDurations
		res.TimeLeft = &timeLeft

		if c.logState == stateOn {
			fullDayAt := c.lastOn.Add(timeLeft)
			res.FullDayAt = &fullDayAt
		}
	}

	if sumDurations > c.fullDay {
		surplus := sumDurations - c.fullDay
		res.Surplus = &surplus
	}

	for cat, dur := range c.taskCatDurations {
		res.AddCategory(cat, dur)
	}

	if c.currentDate.Day != 0 && c.currentDate.Month != 0 {
		res.Date = &summary.Date{
			DayName: c.currentDate.DayName,
			Day:     c.currentDate.Day,
			Month:   c.currentDate.Month,
			Year:    c.currentDate.Year,
		}
	}

	return res
}

// Formats:
// 19:00 - Category
// 19:00 - Category: Task
func parseCategoryAndTask(s string) (category, task string) {
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

func summaryDate(res *summary.Summary) string {

	if res.Date == nil || res.Date.Day == 0 || res.Date.Month == 0 {
		return formatTimestamp(time.Now())
	}

	if res.Date.Year == 0 {
		res.Date.Year = time.Now().Year()
	}

	return fmt.Sprintf("%s %02d.%02d.%d", res.Date.DayName, res.Date.Day, res.Date.Month, res.Date.Year)
}

func formatTimestamp(ts time.Time) string {
	return ts.Format("15:04")
}
