package format

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	// duration formats
	TimeHMS = "hms" // hours, minutes and seconds
	TimeHM  = "hm"  // hours and minutes (default)
	TimeM   = "m"   // minutes
)

func Timestamp(ts time.Time) string {
	return ts.Format("15:04")
}

func DurationM(d time.Duration) string {
	return fmt.Sprintf("%dm", int(math.Floor(d.Minutes())))
}

func DurationHM(d time.Duration) string {
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

func DurationHMS(d time.Duration) string {
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
