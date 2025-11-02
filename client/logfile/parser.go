package logfile

import (
	"fmt"
	"regexp"

	"github.com/sporadisk/clocker/format"
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
	outputFormat    string // the format to use for durations
}

func (l *LogParser) Init() error {
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

	l.outputFormat = format.TimeHM
	l.warnings = []string{}
	return nil
}

func (l *LogParser) addWarning(w string) {
	l.warnings = append(l.warnings, w)
}
