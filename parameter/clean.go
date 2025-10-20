package parameter

import "strings"

func Clean(param string) string {
	return strings.ToLower(strings.TrimSpace(param))
}
