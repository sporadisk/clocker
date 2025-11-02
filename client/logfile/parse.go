package logfile

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/sporadisk/clocker/format"
	"github.com/sporadisk/clocker/logentry"
)

func (l *LogParser) Parse(text string) []logentry.Entry {
	lines := strings.Split(text, "\n")
	entries := []logentry.Entry{}

	for i, line := range lines {
		valid, entry := l.parseLine(line, i+1)
		if valid {
			entries = append(entries, entry)
		}
	}

	return entries
}

const (
	startPatternRegex         = `(?i)^\s*(\d+:\d+)\s+-\s+(start|on|back|resume)`
	stopPatternRegex          = `(?i)^\s*(\d+:\d+)\s+-\s+(stop|break|pause|done|off|end)`
	categorizedTimestampRegex = `^\s*(\d+:\d+)\s+-\s+(.+)` // any other timestamped line
	flexPatternRegex          = `(?i)^\s*flex:\s*([\dhm ]+)`
	targetPatternRegex        = `(?i)^\s*(target|full day|workday):\s*([\dhm ]+)`
	outputPatternRegex        = `(?i)^\s*(output|format):\s*(hms|hm|m)`
	fullDatePatternRegex      = `^\s*--\s*(\p{L}+)\s+(\d+)\.(\d+)\.(\d+)`
	dayMonthPatternRegex      = `^\s*--\s*(\p{L}+)\s+(\d+)\.(\d+)`

	commandFlex = "flex"
)

func (l *LogParser) parseLine(text string, lineNumber int) (valid bool, entry logentry.Entry) {
	startMatches := l.startPattern.FindStringSubmatch(text)
	if startMatches != nil {
		ts, err := format.ParseTimestamp(startMatches[1])
		if err == nil {
			entry.Action = logentry.ActionClockIn
			entry.Command = startMatches[2]
			entry.LineNumber = lineNumber
			entry.Timestamp = &ts
			return true, entry
		} else {
			l.addWarningf("error parsing start time from value %#v: %s", startMatches[1], err.Error())
		}
	}

	stopMatches := l.stopPattern.FindStringSubmatch(text)
	if stopMatches != nil {
		ts, err := format.ParseTimestamp(stopMatches[1])
		if err == nil {
			entry.Action = logentry.ActionClockOut
			entry.Command = stopMatches[2]
			entry.LineNumber = lineNumber
			entry.Timestamp = &ts
			return true, entry
		} else {
			l.addWarningf("error parsing stop time from value %#v: %s", stopMatches[1], err.Error())
		}
	}

	// No start or stop match: Check for categorized timestamp line
	catTsMatches := l.catTsPattern.FindStringSubmatch(text)
	if catTsMatches != nil {
		// this line describes the start of a new named task
		ts, err := format.ParseTimestamp(catTsMatches[1])
		if err == nil {
			entry.Action = logentry.ActionStartTask
			entry.Command = logentry.ActionStartTask
			entry.LineNumber = lineNumber
			entry.Timestamp = &ts
			entry.Task = strings.TrimSpace(catTsMatches[2])
			return true, entry
		} else {
			l.addWarningf("error parsing other timestamp from value %#v: %s", catTsMatches[1], err.Error())
		}
	}

	flexMatches := l.flexPattern.FindStringSubmatch(text)
	if flexMatches != nil {
		d, err := format.ParseDuration(flexMatches[1])
		if err == nil {
			entry.Action = logentry.ActionFlex
			entry.Command = commandFlex
			entry.LineNumber = lineNumber
			entry.Duration = &d
			return true, entry
		} else {
			l.addWarningf("error parsing flex duration from value %#v: %s", flexMatches[1], err.Error())
		}
	}

	targetMatches := l.targetPattern.FindStringSubmatch(text)
	if targetMatches != nil {
		d, err := format.ParseDuration(targetMatches[2])
		if err == nil {
			entry.Action = logentry.ActionTarget
			entry.Command = targetMatches[1]
			entry.LineNumber = lineNumber
			entry.Duration = &d
			return true, entry
		} else {
			l.addWarningf("error parsing target duration from value %#v: %s", targetMatches[2], err.Error())
		}
	}

	formatMatches := l.outputPattern.FindStringSubmatch(text)
	if formatMatches != nil {
		l.outputFormat = strings.ToLower(formatMatches[2])
		entry.Action = logentry.ActionOutputFormat
		entry.Command = formatMatches[1]
		entry.LineNumber = lineNumber
		return true, entry
	}

	fullDateMatches := l.fullDatePattern.FindStringSubmatch(text)
	if fullDateMatches != nil {
		entry.Action = logentry.ActionSetDay
		entry.LineNumber = lineNumber
		day, month, year, err := parseFullDate(fullDateMatches[2], fullDateMatches[3], fullDateMatches[4])
		if err != nil {
			l.addWarningf("error parsing full date from value %#v: %s", fullDateMatches[0], err.Error())
			return false, logentry.Entry{}
		}
		entry.DayName = fullDateMatches[1]
		entry.Day = day
		entry.Month = month
		entry.Year = year
		return true, entry
	}

	// No full date match: Check for day and month line
	dayMatches := l.dayMonthPattern.FindStringSubmatch(text)
	if dayMatches != nil {
		entry.Action = logentry.ActionSetDay
		entry.LineNumber = lineNumber
		day, month, err := parseDayAndMonth(dayMatches[2], dayMatches[3])
		if err != nil {
			l.addWarningf("error parsing day and month from value %#v: %s", dayMatches[0], err.Error())
			return false, logentry.Entry{}
		}
		entry.DayName = dayMatches[1]
		entry.Day = day
		entry.Month = month
		entry.Year = time.Now().Year() // default to current year
		return true, entry
	}

	return false, logentry.Entry{}
}

func (l *LogParser) addWarningf(format string, v ...any) {
	l.addWarning(fmt.Sprintf(format, v...))
}

// Expected partial date format: "dayname dd.mm"
func parseDayAndMonth(dayStr, monthStr string) (day, month int, err error) {
	day, err = strconv.Atoi(dayStr)
	if err != nil {
		return 0, 0, fmt.Errorf("could not parse day %q: %w", dayStr, err)
	}
	month, err = strconv.Atoi(monthStr)
	if err != nil {
		return 0, 0, fmt.Errorf("could not parse month %q: %w", monthStr, err)
	}
	if month < 1 || month > 12 {
		return 0, 0, fmt.Errorf("month value out of range: %d", month)
	}
	if day < 1 || day > 31 {
		return 0, 0, fmt.Errorf("day value out of range: %d", day)
	}
	return day, month, nil
}

// Expected full date format: "dayname dd.mm.yyyy"
func parseFullDate(dayStr, monthStr, yearStr string) (day, month, year int, err error) {
	day, err = strconv.Atoi(dayStr)
	if err != nil {
		return day, month, year, fmt.Errorf("could not parse day %q: %w", dayStr, err)
	}
	month, err = strconv.Atoi(monthStr)
	if err != nil {
		return day, month, year, fmt.Errorf("could not parse month %q: %w", monthStr, err)
	}
	year, err = strconv.Atoi(yearStr)
	if err != nil {
		return day, month, year, fmt.Errorf("could not parse year %q: %w", yearStr, err)
	}

	if month < 1 || month > 12 {
		return day, month, year, fmt.Errorf("month value out of range: %d", month)
	}
	if day < 1 || day > 31 {
		return day, month, year, fmt.Errorf("day value out of range: %d", day)
	}

	return day, month, year, nil
}
