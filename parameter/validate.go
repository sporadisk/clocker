package parameter

import (
	"fmt"
	"strings"
)

func Validate(param string, validOptions []string) (string, error) {
	cleanParam := Clean(param)

	for _, option := range validOptions {
		if strings.EqualFold(cleanParam, option) {
			return option, nil
		}
	}

	validParamStr := strings.Join(validOptions, ", ")
	return "", fmt.Errorf("invalid param %q: Expected one of: %s", cleanParam, validParamStr)
}
