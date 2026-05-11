//go:build !windows
// +build !windows

package cmdargs

import (
	"strings"
	"unicode"
)

func containsSpecialChar(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		if c == '$' || c == '`' || c == '"' || c == '\\' || unicode.IsSpace(c) {
			return true
		}
	}

	return false
}

// AppendCliArg appends a command-line argument to a strings.Builder, ensuring that it is properly escaped or quoted as needed.
// It handles special characters like spaces and quotes, and ensures that the argument is correctly formatted for command-line usage.
func AppendCliArg(sb *strings.Builder, s string) *strings.Builder {
	if len(s) == 0 {
		return sb
	}

	// If the string contains special characters, we need to escape it.
	if !containsSpecialChar(s) {
		sb.WriteString(s)
		return sb
	}

	sb.WriteRune('"')
	for _, c := range s {
		if c == '$' || c == '"' || c == '\\' || c == '`' {
			sb.WriteRune('\\')
		}

		sb.WriteRune(c)
	}

	sb.WriteRune('"')
	return sb
}
