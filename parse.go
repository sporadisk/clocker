package clocker

import (
	"fmt"
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
	startPatternRegex  = `(?i)^\s*(\d+:\d+)\s+-\s+(start|on|back|resume)`
	stopPatternRegex   = `(?i)^\s*(\d+:\d+)\s+-\s+(stop|break|pause|done|off|end)`
	flexPatternRegex   = `(?i)^\s*flex:\s*([\dhm ]+)`
	targetPatternRegex = `(?i)^\s*(target|full day|workday):\s*([\dhm ]+)`
	outputPatternRegex = `(?i)^\s*(output|format):\s*(hms|hm|m)`

	actionClockIn      = "on"
	actionClockOut     = "off"
	actionFlex         = "flex"
	actionTarget       = "target"
	actionOutputFormat = "outputformat"

	stateInit   = "init"
	stateOn     = "on"
	stateOff    = "off"
	stateFlex   = "flex"
	stateTarget = "target"
)

type clockEntry struct {
	action     string // the action to perform based on the interpretation of the command
	command    string // the actual command used on the original line
	timestamp  *time.Time
	duration   *time.Duration
	lineNumber int
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

	flexMatches := l.flexPattern.FindStringSubmatch(text)
	if flexMatches != nil {
		d, err := parseDuration(flexMatches[1])
		if err == nil {
			entry.action = actionFlex
			entry.command = "flex"
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
	return false, clockEntry{}
}

func (l *LogParser) addWarningf(format string, v ...any) {
	l.addWarning(fmt.Sprintf(format, v...))
}
