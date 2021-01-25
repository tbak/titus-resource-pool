package xstring

import (
	"strings"
)

// Split value by comma, trim individual elements, and add to a result only non-empty values.
func SplitByCommaAndTrim(value string) []string {
	if value == "" {
		return []string{}
	}
	parts := strings.Split(value, ",")
	var result []string
	for _, part := range parts {
		trimmed := strings.TrimSpace(part)
		if len(trimmed) > 0 {
			result = append(result, trimmed)
		}
	}
	if result == nil {
		return []string{}
	}
	return result
}
