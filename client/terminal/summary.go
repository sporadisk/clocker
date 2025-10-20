package terminal

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/sporadisk/clocker/format"
	"github.com/sporadisk/clocker/summary"
)

var (
	ErrInvalidInput = errors.New("invalid input")
)

func (c *Client) OutputSummary(sum summary.Summary) error {
	outStr, err := c.Summary(sum)
	if err != nil {
		return err
	}

	fmt.Print(outStr)
	return nil
}

func (c *Client) Summary(sum summary.Summary) (string, error) {
	var sb strings.Builder

	if !sum.Valid {
		sb.WriteString("Could not parse input:\n" + sum.ValidationMsg + "\n")
		return sb.String(), ErrInvalidInput
	}

	if sum.Valid {
		sb.WriteString("\n- Summary / " + summaryDate(&sum) + " -\n")

		if len(sum.Categories) > 0 {
			sb.WriteString("\nCategories:\n")
			for _, cat := range sum.Categories {
				sb.WriteString(fmt.Sprintf(" - %s: %s\n", cat.Name, c.formatDuration(cat.TimeWorked)))
			}
			sb.WriteString("\n")
		}

		sb.WriteString("Worked: " + c.formatDuration(sum.TimeWorked) + "\n")

		if sum.TimeLeft != nil {
			sb.WriteString("Remaining: " + c.formatDuration(*sum.TimeLeft) + "\n")
		}

		if sum.FullDayAt != nil {
			sb.WriteString("Full day: " + format.Timestamp(*sum.FullDayAt) + "\n")
		}

		if sum.Surplus != nil {
			sb.WriteString("Full day + " + c.formatDuration(*sum.Surplus) + "\n")
		}
	}

	if len(sum.Warnings) > 0 {
		sb.WriteString("\nWarnings:\n")
		for i, w := range sum.Warnings {
			sb.WriteString(fmt.Sprintf(" %d - %s\n", i+1, w))
		}
	}

	return sb.String(), nil
}

func (c *Client) formatDuration(d time.Duration) string {
	switch c.TimeFormat {
	case format.TimeM:
		return format.DurationM(d)
	case format.TimeHMS:
		return format.DurationHMS(d)
	default:
		return format.DurationHM(d)
	}
}

func summaryDate(sum *summary.Summary) string {

	if sum.Date == nil || sum.Date.Day == 0 || sum.Date.Month == 0 {
		return format.Timestamp(time.Now())
	}

	if sum.Date.Year == 0 {
		sum.Date.Year = time.Now().Year()
	}

	return fmt.Sprintf("%s %02d.%02d.%d", sum.Date.DayName, sum.Date.Day, sum.Date.Month, sum.Date.Year)
}
