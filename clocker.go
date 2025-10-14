package clocker

import (
	"errors"
	"fmt"
	"math"
	"regexp"
	"strings"
	"time"
	"unicode"
)

const (
	// duration formats
	formatHMS = "hms" // hours, minutes and seconds
	formatHM  = "hm"  // hours and minutes (default)
	formatM   = "m"   // minutes
)

type LogParser struct {
	startPattern    *regexp.Regexp
	stopPattern     *regexp.Regexp
	catTsPattern    *regexp.Regexp
	flexPattern     *regexp.Regexp
	targetPattern   *regexp.Regexp
	outputPattern   *regexp.Regexp
	fullDatePattern *regexp.Regexp
	dayMonthPattern *regexp.Regexp
	warnings        []string
	fullDay         time.Duration
	outputFormat    string // the format to use for durations
}

func (l *LogParser) Init(fullDay string) error {
	startPattern, err := regexp.Compile(startPatternRegex)
	if err != nil {
		return fmt.Errorf("failed to compile start pattern: %w", err)
	}
	l.startPattern = startPattern

	stopPattern, err := regexp.Compile(stopPatternRegex)
	if err != nil {
		return fmt.Errorf("failed to compile stop pattern: %w", err)
	}
	l.stopPattern = stopPattern

	catTimestampPattern, err := regexp.Compile(categorizedTimestampRegex)
	if err != nil {
		return fmt.Errorf("failed to compile categorized timestamp pattern: %w", err)
	}
	l.catTsPattern = catTimestampPattern

	fullDatePattern, err := regexp.Compile(fullDatePatternRegex)
	if err != nil {
		return fmt.Errorf("failed to compile full date pattern: %w", err)
	}
	l.fullDatePattern = fullDatePattern

	dayMonthPattern, err := regexp.Compile(dayMonthPatternRegex)
	if err != nil {
		return fmt.Errorf("failed to compile day pattern: %w", err)
	}
	l.dayMonthPattern = dayMonthPattern

	flexPattern, err := regexp.Compile(flexPatternRegex)
	if err != nil {
		return fmt.Errorf("failed to compile flex pattern: %w", err)
	}
	l.flexPattern = flexPattern

	targetPattern, err := regexp.Compile(targetPatternRegex)
	if err != nil {
		return fmt.Errorf("failed to compile target pattern: %w", err)
	}
	l.targetPattern = targetPattern

	outputPattern, err := regexp.Compile(outputPatternRegex)
	if err != nil {
		return fmt.Errorf("failed to compile format pattern: %w", err)
	}
	l.outputPattern = outputPattern

	l.fullDay = time.Minute * 450 // 7.5h

	if fullDay != "" {
		fdDuration, err := parseDuration(fullDay)
		if err != nil {
			return fmt.Errorf("could not parse fullDay: %w", err)
		}
		l.fullDay = fdDuration
	}

	l.outputFormat = formatHM
	l.warnings = []string{}
	return nil
}

func (l *LogParser) Summary(input string) (string, error) {
	entries := l.parse(input)
	calcResult := l.calculate(entries)
	return l.makeSummary(calcResult)
}

var (
	ErrInvalidInput = errors.New("invalid input")
)

func (l *LogParser) makeSummary(res calcResult) (string, error) {
	var sb strings.Builder

	if !res.valid {
		sb.WriteString("Could not parse input:\n" + res.validationMsg + "\n")
		return sb.String(), ErrInvalidInput
	}

	if res.valid {
		sb.WriteString("\n- Summary / " + summaryDate(&res) + " -\n")

		if len(res.categories) > 0 {
			sb.WriteString("\nCategories:\n")
			for _, c := range res.categories {
				sb.WriteString(fmt.Sprintf(" - %s: %s\n", c.name, l.formatDuration(c.timeWorked)))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("Worked: " + l.formatDuration(res.timeWorked) + "\n")

		if res.timeLeft != nil {
			sb.WriteString("Remaining: " + l.formatDuration(*res.timeLeft) + "\n")
		}

		if res.fullDayAt != nil {
			sb.WriteString("Full day: " + formatTimestamp(*res.fullDayAt) + "\n")
		}

		if res.surplus != nil {
			sb.WriteString("Full day (" + l.formatDuration(l.fullDay) + ") + " + l.formatDuration(*res.surplus) + "\n")
		}
	}

	if len(l.warnings) > 0 {
		sb.WriteString("\nWarnings:\n")
		for i, w := range l.warnings {
			sb.WriteString(fmt.Sprintf(" %d - %s\n", i+1, w))
		}
	}

	return sb.String(), nil
}

func (l *LogParser) addWarning(w string) {
	l.warnings = append(l.warnings, w)
}

func (l *LogParser) formatDuration(d time.Duration) string {
	switch l.outputFormat {
	case formatM:
		return formatDurationM(d)
	case formatHMS:
		return formatDurationHMS(d)
	default:
		return formatDurationHM(d)
	}
}

func parseTimestamp(ts string) (time.Time, error) {
	return time.Parse("15:04", ts)
}

func formatTimestamp(ts time.Time) string {
	return ts.Format("15:04")
}

func parseDuration(d string) (time.Duration, error) {
	return time.ParseDuration(removeSpaces(d))
}

func formatDurationM(d time.Duration) string {
	return fmt.Sprintf("%dm", int(math.Floor(d.Minutes())))
}

func formatDurationHM(d time.Duration) string {
	hours := int(math.Floor(d.Hours()))
	d = d - (time.Duration(hours) * time.Hour)
	minutes := int(math.Floor(d.Minutes()))

	var sb strings.Builder
	if hours > 0 {
		sb.WriteString(fmt.Sprintf("%dh", hours))
	}

	if minutes > 0 {
		if hours > 0 {
			sb.WriteString(" ")
		}

		sb.WriteString(fmt.Sprintf("%dm", minutes))
	}

	return sb.String()
}

func formatDurationHMS(d time.Duration) string {
	hours := int(math.Floor(d.Hours()))
	d = d - (time.Duration(hours) * time.Hour)
	minutes := int(math.Floor(d.Minutes()))
	d = d - (time.Duration(minutes) * time.Minute)
	seconds := int(math.Floor(d.Seconds()))

	var sb strings.Builder
	if hours > 0 {
		sb.WriteString(fmt.Sprintf("%dh", hours))
	}

	if minutes > 0 {
		if hours > 0 {
			sb.WriteString(" ")
		}

		sb.WriteString(fmt.Sprintf("%dm", minutes))
	}

	if seconds > 0 {
		if hours > 0 || seconds > 0 {
			sb.WriteString(" ")
		}

		sb.WriteString(fmt.Sprintf("%ds", seconds))
	}

	return sb.String()
}

func summaryDate(res *calcResult) string {

	if res.date == nil || res.date.day == 0 || res.date.month == 0 {
		return formatTimestamp(time.Now())
	}

	if res.date.year == 0 {
		res.date.year = time.Now().Year()
	}

	return fmt.Sprintf("%02d.%02d.%d", res.date.day, res.date.month, res.date.year)
}

func removeSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, rune := range s {
		if !unicode.IsSpace(rune) {
			b.WriteRune(rune)
		}
	}
	return b.String()
}
