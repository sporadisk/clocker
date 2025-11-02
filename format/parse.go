package format

import (
	"strings"
	"time"
	"unicode"
)

func ParseTimestamp(ts string) (time.Time, error) {
	return time.Parse("15:04", ts)
}

func ParseDuration(d string) (time.Duration, error) {
	return time.ParseDuration(RemoveSpaces(d))
}

func RemoveSpaces(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	for _, rune := range s {
		if !unicode.IsSpace(rune) {
			b.WriteRune(rune)
		}
	}
	return b.String()
}

func CleanParam(param string) string {
	return strings.ToLower(strings.TrimSpace(param))
}
