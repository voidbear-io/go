//go:build windows
// +build windows

package cmdargs

import "strings"

func containsSpecialChars(s string) bool {
	if len(s) == 0 {
		return false
	}

	for _, c := range s {
		if c == ' ' || c == '"' {
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

	// based on the logic from http://stackoverflow.com/questions/5510343/escape-command-line-arguments-in-c-sharp.
	// The method given there doesn't minimize the use of quotation. For that, I drew from
	// https://blogs.msdn.microsoft.com/twistylittlepassagesallalike/2011/04/23/everyone-quotes-command-line-arguments-the-wrong-way/

	// the essential encoding logic is:
	// (1) non-empty strings with no special characters require no encoding
	// (2) find each substring of 0-or-more \ followed by " and replace it by twice-as-many \, followed by \"
	// (3) check if argument ends on \ and if so, double the number of backslashes at the end
	// (4) add leading and trailing "
	if !containsSpecialChars(s) {
		sb.WriteString(s)
		return sb
	}

	backslashCount := 0
	sb.WriteRune('"')
	for _, c := range s {
		switch c {
		case '\\':
			backslashCount++

		case '"':
			times := (2 * backslashCount) + 1
			backslashCount = 0
			if times > 0 {
				for i := 0; i < times; i++ {
					sb.WriteRune('\\')
				}
			}
			sb.WriteRune('"')
		default:
			if backslashCount > 0 {
				for i := 0; i < backslashCount; i++ {
					sb.WriteRune('\\')
				}
			}
			backslashCount = 0
			sb.WriteRune(c)
		}

	}

	if backslashCount > 0 {
		for i := 0; i < backslashCount; i++ {
			sb.WriteRune('\\')
		}
	}

	sb.WriteRune('"')

	return sb
}
