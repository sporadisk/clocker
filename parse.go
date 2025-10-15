package clocker

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

func (l *LogParser) parse(text string) []clockEntry {
	lines := strings.Split(text, "\n")
	entries := []clockEntry{}

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
	fullDatePatternRegex      = `^\s*--\s*(\w+)\s+(\d+)\.(\d+)\.(\d+)`
	dayMonthPatternRegex      = `^\s*--\s*(\w+)\s+(\d+)\.(\d+)`

	actionClockIn      = "on"
	actionClockOut     = "off"
	actionStartTask    = "starttask"
	actionFlex         = "flex"
	actionTarget       = "target"
	actionOutputFormat = "outputformat"
	actionSetDay       = "setday"

	commandFlex = "flex"

	stateInit   = "init"
	stateOn     = "on"
	stateOff    = "off"
	stateFlex   = "flex"
	stateTarget = "target"
)

type clockEntry struct {
	action     string // the action to perform based on the interpretation of the command
	command    string // the actual command used on the original line
	task       string // optional task name
	timestamp  *time.Time
	duration   *time.Duration
	lineNumber int
	dayName    string
	day        int
	month      int
	year       int
}

func (l *LogParser) parseLine(text string, lineNumber int) (valid bool, entry clockEntry) {
	startMatches := l.startPattern.FindStringSubmatch(text)
	if startMatches != nil {
		ts, err := parseTimestamp(startMatches[1])
		if err == nil {
			entry.action = actionClockIn
			entry.command = startMatches[2]
			entry.lineNumber = lineNumber
			entry.timestamp = &ts
			return true, entry
		} else {
			l.addWarningf("error parsing start time from value %#v: %s", startMatches[1], err.Error())
		}
	}

	stopMatches := l.stopPattern.FindStringSubmatch(text)
	if stopMatches != nil {
		ts, err := parseTimestamp(stopMatches[1])
		if err == nil {
			entry.action = actionClockOut
			entry.command = stopMatches[2]
			entry.lineNumber = lineNumber
			entry.timestamp = &ts
			return true, entry
		} else {
			l.addWarningf("error parsing stop time from value %#v: %s", stopMatches[1], err.Error())
		}
	}

	// No start or stop match: Check for categorized timestamp line
	catTsMatches := l.catTsPattern.FindStringSubmatch(text)
	if catTsMatches != nil {
		// this line describes the start of a new named task
		ts, err := parseTimestamp(catTsMatches[1])
		if err == nil {
			entry.action = actionStartTask
			entry.command = actionStartTask
			entry.lineNumber = lineNumber
			entry.timestamp = &ts
			entry.task = strings.TrimSpace(catTsMatches[2])
			return true, entry
		} else {
			l.addWarningf("error parsing other timestamp from value %#v: %s", catTsMatches[1], err.Error())
		}
	}

	flexMatches := l.flexPattern.FindStringSubmatch(text)
	if flexMatches != nil {
		d, err := parseDuration(flexMatches[1])
		if err == nil {
			entry.action = actionFlex
			entry.command = commandFlex
			entry.lineNumber = lineNumber
			entry.duration = &d
			return true, entry
		} else {
			l.addWarningf("error parsing flex duration from value %#v: %s", flexMatches[1], err.Error())
		}
	}

	targetMatches := l.targetPattern.FindStringSubmatch(text)
	if targetMatches != nil {
		d, err := parseDuration(targetMatches[2])
		if err == nil {
			entry.action = actionTarget
			entry.command = targetMatches[1]
			entry.lineNumber = lineNumber
			entry.duration = &d
			return true, entry
		} else {
			l.addWarningf("error parsing target duration from value %#v: %s", targetMatches[2], err.Error())
		}
	}

	formatMatches := l.outputPattern.FindStringSubmatch(text)
	if formatMatches != nil {
		l.outputFormat = strings.ToLower(formatMatches[2])
		entry.action = actionOutputFormat
		entry.command = formatMatches[1]
		entry.lineNumber = lineNumber
		return true, entry
	}

	fullDateMatches := l.fullDatePattern.FindStringSubmatch(text)
	if fullDateMatches != nil {
		entry.action = actionSetDay
		entry.lineNumber = lineNumber
		day, month, year, err := parseFullDate(fullDateMatches[2], fullDateMatches[3], fullDateMatches[4])
		if err != nil {
			l.addWarningf("error parsing full date from value %#v: %s", fullDateMatches[0], err.Error())
			return false, clockEntry{}
		}
		entry.dayName = fullDateMatches[1]
		entry.day = day
		entry.month = month
		entry.year = year
		return true, entry
	}

	// No full date match: Check for day and month line
	dayMatches := l.dayMonthPattern.FindStringSubmatch(text)
	if dayMatches != nil {
		entry.action = actionSetDay
		entry.lineNumber = lineNumber
		day, month, err := parseDayAndMonth(dayMatches[2], dayMatches[3])
		if err != nil {
			l.addWarningf("error parsing day and month from value %#v: %s", dayMatches[0], err.Error())
			return false, clockEntry{}
		}
		entry.dayName = dayMatches[1]
		entry.day = day
		entry.month = month
		entry.year = time.Now().Year() // default to current year
		return true, entry
	}

	return false, clockEntry{}
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
