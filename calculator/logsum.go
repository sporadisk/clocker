package calculator

import (
	"fmt"
	"strings"
	"time"

	"github.com/sporadisk/clocker/event"
	"github.com/sporadisk/clocker/format"
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

type LogSummary struct {
	// input
	Entries      []logentry.Entry
	FullDay      time.Duration
	CatParseMode string

	logState         string
	lastOn           time.Time
	lastOff          time.Time
	durations        []time.Duration
	taskCatDurations map[string]time.Duration
	prevCommand      string
	currentCategory  string
	currentTask      string
	currentDate      summary.Date
	events           []*event.Event
}

func (ls *LogSummary) Sum() (summary.Summary, []*event.Event) {

	ls.logState = stateInit
	ls.lastOn = time.Time{}
	ls.lastOff = time.Time{}
	ls.durations = []time.Duration{}
	ls.taskCatDurations = map[string]time.Duration{}
	ls.prevCommand = "--start of document--"
	ls.events = []*event.Event{}

	if ls.CatParseMode != "" {
		ls.CatParseMode = format.CleanParam(ls.CatParseMode)
	} else {
		ls.CatParseMode = "v1"
	}

	if len(ls.Entries) == 0 {
		return summary.Summary{
			Valid:         false,
			ValidationMsg: "No valid time entries detected.",
		}, nil
	}

	for _, entry := range ls.Entries {
		if entry.Action == logentry.ActionClockIn {
			success, res := ls.clockIn(entry)
			if !success {
				return res, nil
			}
		}

		if entry.Action == logentry.ActionStartTask {
			category, task := ls.parseCategoryAndTask(entry.Task)
			success, res := ls.startTask(entry, category, task)
			if !success {
				return res, nil
			}
		}

		if entry.Action == logentry.ActionClockOut {
			success, res := ls.clockOut(entry)
			if !success {
				return res, nil
			}
		}

		if entry.Action == logentry.ActionFlex {
			success, res := ls.flex(entry)
			if !success {
				return res, nil
			}
		}

		if entry.Action == logentry.ActionTarget {
			if ls.logState == stateOn {
				return summary.Summary{
					Valid:         false,
					ValidationMsg: fmt.Sprintf(`"%s" entry on line %d follows a clock-in, which is wrong.`, entry.Command, entry.LineNumber),
				}, nil
			}

			ls.logState = stateTarget
			ls.FullDay = *entry.Duration
		}

		if entry.Action == logentry.ActionSetDay {
			ls.currentDate = summary.Date{
				DayName: entry.DayName,
				Day:     entry.Day,
				Month:   entry.Month,
				Year:    entry.Year,
			}
		}

		ls.prevCommand = entry.Command
	}

	// validation and duration collection complete: Time to calculate
	return ls.summarize(), ls.events
}

func (ls *LogSummary) clockOut(entry logentry.Entry) (success bool, result summary.Summary) {
	if ls.logState != stateOn {
		return false, summary.Summary{
			Valid:         false,
			ValidationMsg: fmt.Sprintf(`Clock-out at %s on line %d follows "%s", should follow a clock-in`, entry.Timestamp.Format(timestampFormat), entry.LineNumber, ls.logState),
		}
	}

	if !entry.Timestamp.After(ls.lastOn) {
		return false, summary.Summary{
			Valid:         false,
			ValidationMsg: fmt.Sprintf(`Clock-out on line %d has an earlier timestamp than its corresponding "%s"`, entry.LineNumber, ls.prevCommand),
		}
	}

	ls.lastOff = *entry.Timestamp
	ls.logState = stateOff
	ls.durations = append(ls.durations, entry.Timestamp.Sub(ls.lastOn))

	if ls.currentCategory != "" {
		ls.addToCategory(ls.currentCategory, entry.Timestamp.Sub(ls.lastOn))
	}

	eventCategory := ls.currentCategory
	if eventCategory == "" {
		eventCategory = summary.Uncategorized
	}

	event := &event.Event{
		Start:    eventTimeStamp(ls.currentDate, ls.lastOn),
		End:      eventTimeStamp(ls.currentDate, ls.lastOff),
		Category: eventCategory,
		Task:     ls.currentTask,
		Date: event.EventDate{
			Day:   ls.currentDate.Day,
			Month: ls.currentDate.Month,
			Year:  ls.currentDate.Year,
		},
	}
	event.DetermineHours()
	ls.events = append(ls.events, event)

	// reset current task and category
	ls.currentTask = ""
	ls.currentCategory = ""

	return true, summary.Summary{}
}

// the log-entry timestamps lack a date, so we need to supply that
func eventTimeStamp(d summary.Date, t time.Time) time.Time {
	// get the current local timestamp, in order to use its location
	now := time.Now()

	// construct a new time.Time with the date from `d`, the time from `t` and
	// the timezone from `now`
	return time.Date(d.Year, time.Month(d.Month), d.Day, t.Hour(), t.Minute(), t.Second(), t.Nanosecond(), now.Location())
}

func (ls *LogSummary) addToCategory(cat string, dur time.Duration) {
	if cat == "" {
		return
	}

	_, ok := ls.taskCatDurations[cat]
	if !ok {
		ls.taskCatDurations[cat] = 0
	}

	ls.taskCatDurations[cat] += dur
}

func (ls *LogSummary) clockIn(entry logentry.Entry) (success bool, result summary.Summary) {
	if ls.logState == stateOn {
		return false, summary.Summary{
			Valid:         false,
			ValidationMsg: fmt.Sprintf(`Duplicate clock-in on line %d`, entry.LineNumber),
		}
	}

	if entry.Timestamp.Before(ls.lastOff) && !ls.lastOff.IsZero() {
		return false, summary.Summary{
			Valid: false,
			ValidationMsg: fmt.Sprintf(`Clock-in at %s on line %d occurs prior to the previous clock-out (%s)`,
				entry.Timestamp.Format(timestampFormat),
				entry.LineNumber,
				ls.lastOff.Format(timestampFormat),
			),
		}
	}

	ls.lastOn = *entry.Timestamp
	ls.logState = stateOn

	return true, summary.Summary{}
}

func (ls *LogSummary) startTask(entry logentry.Entry, category, task string) (success bool, result summary.Summary) {
	// We're already clocked in, probably on another task
	// Clock out of the previous task first
	if ls.logState == stateOn {
		success, res := ls.clockOut(entry)
		if !success {
			return false, res
		}
	}

	ls.currentTask = task
	ls.currentCategory = category

	return ls.clockIn(entry)
}

func (ls *LogSummary) flex(entry logentry.Entry) (success bool, result summary.Summary) {
	if ls.logState == stateOn {
		return false, summary.Summary{
			Valid:         false,
			ValidationMsg: fmt.Sprintf(`Flex time entry on line %d follows a clock-in, which is wrong.`, entry.LineNumber),
		}
	}

	ls.logState = stateFlex
	ls.durations = append(ls.durations, *entry.Duration)
	return true, summary.Summary{}
}

func (ls *LogSummary) summarize() summary.Summary {
	res := summary.Summary{
		Valid: true,
	}

	sumDurations := time.Duration(0)

	for _, d := range ls.durations {
		sumDurations = sumDurations + d
	}
	res.TimeWorked = sumDurations

	if sumDurations < ls.FullDay {
		timeLeft := ls.FullDay - sumDurations
		res.TimeLeft = &timeLeft

		if ls.logState == stateOn {
			fullDayAt := ls.lastOn.Add(timeLeft)
			res.FullDayAt = &fullDayAt
		}
	}

	if sumDurations > ls.FullDay {
		surplus := sumDurations - ls.FullDay
		res.Surplus = &surplus
	}

	for cat, dur := range ls.taskCatDurations {
		res.AddCategory(cat, dur)
	}

	if ls.currentDate.Day != 0 && ls.currentDate.Month != 0 {
		res.Date = &summary.Date{
			DayName: ls.currentDate.DayName,
			Day:     ls.currentDate.Day,
			Month:   ls.currentDate.Month,
			Year:    ls.currentDate.Year,
		}
	}

	return res
}

func (ls *LogSummary) parseCategoryAndTask(s string) (category, task string) {
	switch ls.CatParseMode {
	case "v1":
		return parseCategoryAndTaskV1(s)
	case "v2":
		return parseCategoryAndTaskV2(s)
	}

	return "error-invalidparsemode", "error"
}

// parseCategoryAndTaskV1 parses a string that might contain one of these two
// formats:
// "Category"
// "Category: Task"
func parseCategoryAndTaskV1(s string) (category, task string) {
	if s == "" {
		return "", ""
	}

	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		category = format.CleanParam(parts[0])
		task = format.CleanParam(parts[1])
		return category, task
	}

	category = format.CleanParam(s)
	return category, ""
}

// parseCategoryAndTaskV2 parses a string that might contain a category, like
// in v1. However, if no category is found, the category is set to
// "uncategorized".
func parseCategoryAndTaskV2(s string) (category, task string) {
	if s == "" {
		return summary.Uncategorized, ""
	}

	parts := strings.SplitN(s, ":", 2)
	if len(parts) == 2 {
		category = format.CleanParam(parts[0])
		task = format.CleanParam(parts[1])
		return category, task
	}

	task = format.CleanParam(s)
	return summary.Uncategorized, task
}
