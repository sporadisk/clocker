package format

import (
	"fmt"
	"strings"
)

func ValidateTimeFormat(format string) error {
	validFormats := []string{
		TimeM, TimeHM, TimeHMS,
	}

	for _, vf := range validFormats {
		if format == vf {
			return nil
		}
	}

	return fmt.Errorf("invalid time format %q - Valid formats: %s", format, strings.Join(validFormats, ", "))
}
