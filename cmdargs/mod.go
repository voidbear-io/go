// Package cmdargs provides utilities for parsing, manipulating, and formatting
// command-line arguments. It offers a convenient Args type for working with
// argument slices, including methods for searching, modifying, and converting
// arguments. The package also includes robust parsing logic to split command-line
// strings into arguments, handling quoting, escaping, and whitespace similar to
// shell behavior. Additionally, it provides normalization and formatting functions
// to ensure arguments are handled consistently and safely for CLI usage.
package cmdargs

import (
	"strconv"
	"strings"
	"unicode"
)

const (
	quoteNone   = iota
	quoteDouble = 1
	quoteSingle = 2
)

// Args encapsulates a slice of command-line arguments.
type Args struct {
	args []string
}

// New creates a new Args instance from a slice of strings, normalizing the arguments
// to ensure they are properly formatted for command-line usage.
func New(args []string) *Args {
	return &Args{
		args: normalizeArgs(args),
	}
}

// ToArray returns a copy of the underlying slice of arguments, ensuring that modifications
// to the returned slice do not affect the original Args instance.
// This is useful for safely accessing the arguments without risking unintended changes.
func (a *Args) ToArray() []string {
	copy2 := make([]string, len(a.args))
	copy(copy2, a.args)
	return copy2
}

// Len returns the number of arguments in the Args instance.
// This is a simple utility method to get the count of arguments without exposing the underlying slice.
func (a *Args) Len() int {
	return len(a.args)
}

// Get retrieves the argument at the specified index i.
// If the index is out of bounds, it returns an empty string.
// This method provides safe access to individual arguments without risking panics from out-of-bounds access
func (a *Args) Get(i int) string {
	if i < 0 || i >= len(a.args) {
		return ""
	}
	return a.args[i]
}

// GetAny retrieves the value of the first argument that matches any of the specified options.
// If no match is found, it returns an empty string and false.
func (a *Args) GetAny(s ...string) (string, bool) {
	index := a.IndexAny(s)
	if index == -1 {
		return "", false
	}
	return a.args[index], true
}

// GetAnyOr retrieves the value of the first argument that matches any of the specified options.
// If no match is found, it returns the default value and true.
func (a *Args) GetAnyOr(s ...string) (string, bool) {
	value, ok := a.GetAny(s...)
	if !ok {
		return "", false
	}
	return value, true
}

// GetInt retrieves the integer value of the argument at the specified index i.
// If the index is out of bounds or the argument is not a valid integer, it returns 0 and false.
func (a *Args) GetInt(i int) (int, bool) {
	if i < 0 || i >= len(a.args) {
		return 0, false
	}
	intValue, err := strconv.Atoi(a.args[i])
	if err != nil {
		return 0, false
	}
	return intValue, true
}

// GetIntOr retrieves the integer value of the argument at the specified index i.
// If the index is out of bounds or the argument is not a valid integer, it returns the default value.
func (a *Args) GetIntOr(i int, defaultValue int) int {
	value, ok := a.GetInt(i)
	if !ok {
		return defaultValue
	}
	return value
}

// GetIntAny retrieves the integer value of the first argument that matches any of the specified options.
// If no match is found, it returns 0 and false.
func (a *Args) GetIntAny(s ...string) (int, bool) {
	value, ok := a.GetAny(s...)
	if !ok {
		return 0, false
	}
	intValue, err := strconv.Atoi(value)
	if err != nil {
		return 0, false
	}
	return intValue, true
}

// GetFloat retrieves the float value of the argument at the specified index i.
// If the index is out of bounds or the argument is not a valid float, it returns 0 and false.
func (a *Args) GetFloat(i int) (float64, bool) {
	if i < 0 || i >= len(a.args) {
		return 0, false
	}
	floatValue, err := strconv.ParseFloat(a.args[i], 64)
	if err != nil {
		return 0, false
	}
	return floatValue, true
}

// GetFloatOr retrieves the float value of the argument at the specified index i.
// If the index is out of bounds or the argument is not a valid float, it returns the default value.
func (a *Args) GetFloatOr(i int, defaultValue float64) float64 {
	value, ok := a.GetFloat(i)
	if !ok {
		return defaultValue
	}
	return value
}

// GetFloatAny retrieves the float value of the first argument that matches any of the specified options.
// If no match is found, it returns 0 and false.
func (a *Args) GetFloatAny(s ...string) (float64, bool) {
	value, ok := a.GetAny(s...)
	if !ok {
		return 0, false
	}
	floatValue, err := strconv.ParseFloat(value, 64)
	if err != nil {
		return 0, false
	}
	return floatValue, true
}

// GetBool retrieves the boolean value of the argument at the specified index i.
// If the index is out of bounds or the argument is not a valid boolean, it returns false and false.
func (a *Args) GetBool(i int) (bool, bool) {
	if i < 0 || i >= len(a.args) {
		return false, false
	}

	switch a.args[i] {
	case "true", "1", "yes", "y", "on":
		return true, true
	case "false", "0", "no", "n", "off":
		return false, true
	}
	return false, false
}

// GetBoolOr retrieves the boolean value of the argument at the specified index i.
// If the index is out of bounds or the argument is not a valid boolean, it returns the default value.
func (a *Args) GetBoolOr(i int, defaultValue bool) bool {
	value, ok := a.GetBool(i)
	if !ok {
		return defaultValue
	}
	return value
}

// GetBoolAny retrieves the boolean value of the first argument that matches any of the specified options.
// If no match is found, it returns false and false.
func (a *Args) GetBoolAny(s ...string) (bool, bool) {
	value, ok := a.GetAny(s...)
	if !ok {
		return false, false
	}
	switch value {
	case "true", "1", "yes", "y", "on":
		return true, true
	case "false", "0", "no", "n", "off":
		return false, true
	}
	return false, false
}

// GetSlice collects values for a repeated option: each occurrence of option followed by a
// separate argument adds that next argument to the result (e.g. -e a -e b → [a, b]).
// It returns false if option never appears with a following value.
func (a *Args) GetSlice(option string) ([]string, bool) {
	return a.collectOptionSlice([]string{option})
}

// GetSliceAny is like GetSlice but treats any of the given strings as the same option.
func (a *Args) GetSliceAny(options ...string) ([]string, bool) {
	return a.collectOptionSlice(options)
}

func (a *Args) collectOptionSlice(options []string) ([]string, bool) {
	if len(options) == 0 {
		return nil, false
	}
	var out []string
	n := len(a.args)
	for i := 0; i < n; i++ {
		if !optionMatchesFold(a.args[i], options) {
			continue
		}
		if i+1 >= n {
			break
		}
		out = append(out, a.args[i+1])
		i++
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

// GetMap collects key=value pairs for a repeated option: each occurrence of option followed
// by an argument of the form key=value adds that pair (first '=' separates key and value).
// Later occurrences of the same key replace earlier ones. It returns false if no valid pair
// was found. Arguments without '=' are skipped.
func (a *Args) GetMap(option string) (map[string]string, bool) {
	return a.collectOptionMap([]string{option})
}

// GetMapAny is like GetMap but treats any of the given strings as the same option.
func (a *Args) GetMapAny(options ...string) (map[string]string, bool) {
	return a.collectOptionMap(options)
}

func (a *Args) collectOptionMap(options []string) (map[string]string, bool) {
	if len(options) == 0 {
		return nil, false
	}
	out := make(map[string]string)
	n := len(a.args)
	for i := 0; i < n; i++ {
		if !optionMatchesFold(a.args[i], options) {
			continue
		}
		if i+1 >= n {
			break
		}
		raw := a.args[i+1]
		if key, val, ok := splitKeyValuePair(raw); ok {
			out[key] = val
		}
		i++
	}
	if len(out) == 0 {
		return nil, false
	}
	return out, true
}

func optionMatchesFold(token string, options []string) bool {
	for _, o := range options {
		if strings.EqualFold(token, o) {
			return true
		}
	}
	return false
}

func splitKeyValuePair(s string) (key, val string, ok bool) {
	i := strings.IndexByte(s, '=')
	if i < 0 {
		return "", "", false
	}
	return s[:i], s[i+1:], true
}

// Index returns the index of the first occurrence of the specified string s in the Args.
// If the string is not found, it returns -1. This method performs a case-insitive comparison
// to find the index, allowing for flexible matching of arguments regardless of case.
func (a *Args) Index(s string) int {
	for i, token := range a.args {
		if strings.EqualFold(s, token) {
			return i
		}
	}
	return -1
}

// IndexAny returns the index of the first occurrence of any string in the slice s within the Args.
// If none of the strings are found, it returns -1. This method allows for checking multiple
// potential matches in a single call, improving efficiency for certain use cases.
func (a *Args) IndexAny(s []string) int {
	for i, token := range a.args {
		for _, t := range s {
			if t == token {
				return i
			}
		}
	}
	return -1
}

// IndexFold returns the index of the first occurrence of the specified string s in the Args,
// performing a case-insensitive comparison. If the string is not found, it returns -1.
// This method is useful for scenarios where you want to check for matches without worrying about case sensitivity.
func (a *Args) IndexFold(s string) int {
	for i, token := range a.args {
		if strings.EqualFold(s, token) {
			return i
		}
	}
	return -1
}

// IndexAnyFold returns the index of the first occurrence of any string in the slice s within the Args,
// performing a case-insensitive comparison. If none of the strings are found, it returns -1.
// This method is useful for scenarios where you want to check for matches without worrying about case sensitivity
func (a *Args) IndexAnyFold(s []string) int {
	for i, token := range a.args {
		for _, t := range s {
			if strings.EqualFold(t, token) {
				return i
			}
		}
	}
	return -1
}

// Contains checks if the Args contains the specified string s.
func (a *Args) Contains(s string) bool {
	for _, token := range a.args {
		if token == s {
			return true
		}
	}
	return false
}

// ContainsFold checks if the Args contains the specified string s, performing a case-insensitive comparison.
// This method is useful for scenarios where you want to check for matches without worrying about case sensitivity.
func (a *Args) ContainsFold(s string) bool {
	for _, token := range a.args {
		if strings.EqualFold(s, token) {
			return true
		}
	}
	return false
}

// ContainsAny checks if the Args contains any of the strings in the slice s.
func (a *Args) ContainsAny(s []string) bool {
	for _, token := range a.args {
		for _, t := range s {
			if t == token {
				return true
			}
		}
	}
	return false
}

// ContainsAnyFold checks if the Args contains any of the strings in the slice s,
// performing a case-insensitive comparison. This method is useful for scenarios where
// you want to check for matches without worrying about case sensitivity.
func (a *Args) ContainsAnyFold(s []string) bool {
	for _, token := range a.args {
		for _, t := range s {
			if strings.EqualFold(t, token) {
				return true
			}
		}
	}
	return false
}

// Set updates the argument at the specified index i with the new value.
// If the index is out of bounds, it does nothing. This method allows for modifying
// existing arguments in place.
func (a *Args) Set(i int, value string) {
	if i < 0 || i >= len(a.args) {
		return
	}
	a.args[i] = value
}

// SetValue sets the value of the argument at the specified index i with the new value.
// If the index is out of bounds, it does nothing. This method allows for modifying
// existing arguments in place.
func (a *Args) SetValue(option string, value string) bool {
	index := a.Index(option)
	if index == -1 {
		return false
	}

	if index == len(a.args)-1 {
		a.args = append(a.args, value)
		return true
	}

	a.args = append(a.args[:index+1], a.args[index:]...)
	a.args[index] = value
	return true
}

// SetInt sets the value of the argument at the specified index i with the new value.
// If the index is out of bounds, it does nothing. This method allows for modifying
// existing arguments in place.
func (a *Args) SetInt(option string, value int) bool {

	index := a.Index(option)
	if index == -1 {
		return false
	}

	if index == len(a.args)-1 {
		a.args = append(a.args, strconv.Itoa(value))
		return true
	}

	a.args[index] = strconv.Itoa(value)
	return true
}

// SetFloat sets the value of the argument at the specified index i with the new value.
// If the index is out of bounds, it does nothing. This method allows for modifying
// existing arguments in place.
func (a *Args) SetFloat(option string, value float64) bool {
	return a.SetValue(option, strconv.FormatFloat(value, 'f', -1, 64))
}

// SetBool sets the value of the argument at the specified index i with the new value.
// If the index is out of bounds, it does nothing. This method allows for modifying
// existing arguments in place.
func (a *Args) SetBool(option string, value bool) bool {
	return a.SetValue(option, strconv.FormatBool(value))
}

// Push appends the specified values to the end of the Args slice.
// It normalizes the values to ensure they are properly formatted for command-line usage.
func (a *Args) Push(values ...string) *Args {
	values = normalizeArgs(values)
	a.args = append(a.args, values...)
	return a
}

// Append appends the specified values to the end of the Args slice.
// It normalizes the values to ensure they are properly formatted for command-line usage.
func (a *Args) Append(values ...string) *Args {
	values = normalizeArgs(values)
	a.args = append(a.args, values...)
	return a
}

// Prepend adds the specified values to the beginning of the Args slice.
// It normalizes the values to ensure they are properly formatted for command-line usage.
func (a *Args) Prepend(values ...string) *Args {
	values = normalizeArgs(values)
	a.args = append(values, a.args...)
	return a
}

// Shift removes and returns the first argument from the Args slice.
// If the slice is empty, it returns an empty string. This method is useful for
// processing command-line arguments in a sequential manner.
func (a *Args) Shift() string {
	if len(a.args) == 0 {
		return ""
	}
	value := a.args[0]
	a.args = a.args[1:]
	return value
}

// Pop removes and returns the last argument from the Args slice.
// If the slice is empty, it returns an empty string. This method is useful for
// processing command-line arguments in a sequential manner.
func (a *Args) Pop() string {
	if len(a.args) == 0 {
		return ""
	}
	value := a.args[len(a.args)-1]
	a.args = a.args[:len(a.args)-1]
	return value
}

// Remove removes the first occurrence of the specified string s from the Args slice.
// It returns true if the removal was successful, or false if the string was not found.
func (a *Args) Remove(s string) bool {
	for i, token := range a.args {
		if s == token {
			a.RemoveAt(i)
			return true
		}
	}

	return false
}

// RemoveAt removes the argument at the specified index i from the Args slice.
// It returns true if the removal was successful, or false if the index is out of range.
func (a *Args) RemoveAt(i int) bool {
	if i < 0 || i >= len(a.args) {
		return false
	}
	a.args = append(a.args[:i], a.args[i+1:]...)
	return true
}

// String returns the command-line arguments as a single string, with each argument
// separated by a space. Arguments are properly formatted using appendCliArg to ensure
// correct escaping or quoting as needed. If there are no arguments, an empty string is returned.
func (a *Args) String() string {
	if len(a.args) == 0 {
		return ""
	}

	sb := &strings.Builder{}
	for i, arg := range a.args {
		if i > 0 {
			sb.WriteString(" ")
		}
		AppendCliArg(sb, arg)
	}
	return sb.String()
}

// Split parses the input string s into an Args structure, splitting it into tokens
// similar to how a shell parses command-line arguments. It handles single and double
// quotes, escaped quotes, and whitespace as delimiters. Special handling is included
// for line continuations and escaped newlines. The resulting Args contains the parsed
// arguments as a slice of strings.
func Split(s string) *Args {
	quote := quoteNone
	token := strings.Builder{}
	tokens := []string{}
	runes := []rune(s)
	l := len(runes)
	for i := 0; i < l; i++ {
		c := runes[i]

		if quote != quoteNone {
			previous := rune(0)
			if i > 0 {
				previous = runes[i-1]
			}

			switch quote {
			case quoteSingle:
				if c == '\'' && previous != '\\' {
					quote = quoteNone
					if token.Len() > 0 {
						tokens = append(tokens, token.String())
						token.Reset()
					}

					continue
				}
			case quoteDouble:
				if c == '"' && previous != '\\' {
					quote = quoteNone
					if token.Len() > 0 {
						tokens = append(tokens, token.String())
						token.Reset()
					}

					continue
				}
			}

			token.WriteRune(c)
			continue
		}

		if c == '\\' || c == '`' {
			size := i + 1
			remaining := l - size
			if remaining > 0 {
				j := runes[i+1]
				if j == '\n' {
					i += 1
					if token.Len() > 0 {
						tokens = append(tokens, token.String())
						token.Reset()
					}

					continue
				}

				if remaining > 1 {
					k := runes[i+2]
					if j == '\r' && k == '\n' {
						i += 2
						if token.Len() > 0 {
							tokens = append(tokens, token.String())
							token.Reset()
						}

						continue
					}
				}

				// Not an escaped newline, just add the next character
				token.WriteRune(j)
				i += 1
				continue
			}

			// Trailing backslash, just add it
			token.WriteRune(c)
			continue
		}

		// Stop processing at new lines that are not escaped or within quotes
		if c == '\n' || c == '\r' {
			break
		}

		if c == ' ' {
			if token.Len() == 0 {
				continue
			}

			size := i + 1
			remaining := l - size
			if remaining > 2 {
				j := runes[i+1]
				k := runes[i+2]

				if j == '\n' {
					i += 1
					if token.Len() > 0 {
						tokens = append(tokens, token.String())
						token.Reset()
					}

					continue
				}

				if j == '\r' && k == '\n' {
					i += 2
					if token.Len() > 0 {
						tokens = append(tokens, token.String())
						token.Reset()
					}

					continue
				}

				if (j == '\\' || j == '`') && k == '\n' {
					i += 2

					if token.Len() > 0 {
						tokens = append(tokens, token.String())
					}

					token.Reset()
					continue
				}

				if remaining > 3 {
					l := runes[i+3]
					if (j == '\\' || j == '`') && k == '\r' && l == '\n' {
						i += 3
						if token.Len() > 0 {
							tokens = append(tokens, token.String())
						}

						token.Reset()
						continue
					}
				}
			}

			if token.Len() > 0 {
				tokens = append(tokens, token.String())
				token.Reset()
			}
			continue
		}

		if token.Len() == 0 {
			switch c {
			case '\'':
				quote = quoteSingle
				continue

			case '"':
				quote = quoteDouble
				continue
			}
		}

		if unicode.IsSpace(c) {
			continue
		}

		token.WriteRune(c)
	}

	if token.Len() > 0 {
		tokens = append(tokens, token.String())
	}

	token.Reset()

	return &Args{
		args: tokens,
	}
}

func SplitAndExpand(s string, expand func(string) (string, error)) (*Args, error) {
	quote := quoteNone
	token := strings.Builder{}
	tokens := []string{}
	runes := []rune(s)
	l := len(runes)
	hasDollar := false
	for i := 0; i < l; i++ {
		c := runes[i]

		if c == '$' {
			hasDollar = true
		}

		if quote != quoteNone {
			previous := rune(0)
			if i > 0 {
				previous = runes[i-1]
			}

			switch quote {
			case quoteSingle:
				if c == '\'' && previous != '\\' {
					quote = quoteNone
					if token.Len() > 0 {
						tokens = append(tokens, token.String())
						token.Reset()
					}
					hasDollar = false

					continue
				}
			case quoteDouble:
				if c == '"' && previous != '\\' {
					quote = quoteNone
					if token.Len() > 0 {
						if hasDollar {
							expanded, err := expand(token.String())
							if err != nil {
								return nil, err
							}
							tokens = append(tokens, expanded)
						} else {
							tokens = append(tokens, token.String())
						}
						token.Reset()
					}
					hasDollar = false

					continue
				}
			}

			token.WriteRune(c)
			continue
		}

		// Stop processing at new lines that are not escaped or within quotes
		if c == '\n' || c == '\r' {
			break
		}

		if c == '\\' || c == '`' {
			size := i + 1
			remaining := l - size
			if remaining > 0 {
				j := runes[i+1]
				if j == '\n' {
					i += 1
					if token.Len() > 0 {
						tokens = append(tokens, token.String())
						token.Reset()
					}

					continue
				}

				if remaining > 1 {
					k := runes[i+2]
					if j == '\r' && k == '\n' {
						i += 2
						if token.Len() > 0 {
							tokens = append(tokens, token.String())
							token.Reset()
						}

						continue
					}
				}

				// Not an escaped newline, just add the next character
				token.WriteRune(j)
				i += 1
				continue
			}

			// Trailing backslash, just add it
			token.WriteRune(c)
			continue
		}

		if c == ' ' {
			if token.Len() == 0 {
				continue
			}

			size := i + 1
			remaining := l - size
			if remaining > 2 {
				j := runes[i+1]
				k := runes[i+2]
				if j == '\n' {
					i += 1
					if token.Len() > 0 {
						if hasDollar {
							expanded, err := expand(token.String())
							if err != nil {
								return nil, err
							}
							tokens = append(tokens, expanded)
							hasDollar = false
						} else {
							tokens = append(tokens, token.String())
						}
						token.Reset()
					}

					continue
				}

				if j == '\r' && k == '\n' {
					i += 2
					if token.Len() > 0 {
						if hasDollar {
							expanded, err := expand(token.String())
							if err != nil {
								return nil, err
							}
							tokens = append(tokens, expanded)
							hasDollar = false
						} else {
							tokens = append(tokens, token.String())
						}
						token.Reset()
					}

					continue
				}

				if (j == '\\' || j == '`') && k == '\n' {
					i += 2

					if token.Len() > 0 {
						if hasDollar {
							expanded, err := expand(token.String())
							if err != nil {
								return nil, err
							}
							tokens = append(tokens, expanded)
							hasDollar = false
						} else {
							tokens = append(tokens, token.String())
						}
					}

					token.Reset()
					continue
				}

				if remaining > 3 {
					l := runes[i+3]
					if (j == '\\' || j == '`') && k == '\r' && l == '\n' {
						i += 3
						if token.Len() > 0 {
							if hasDollar {
								expanded, err := expand(token.String())
								if err != nil {
									return nil, err
								}
								tokens = append(tokens, expanded)
								hasDollar = false
							} else {
								tokens = append(tokens, token.String())
							}
						}

						token.Reset()
						continue
					}
				}
			}

			if token.Len() > 0 {
				if hasDollar {
					expanded, err := expand(token.String())
					if err != nil {
						return nil, err
					}
					hasDollar = false
					tokens = append(tokens, expanded)
				} else {
					tokens = append(tokens, token.String())
				}

				token.Reset()
			}
			continue
		}

		if token.Len() == 0 {
			switch c {
			case '\'':
				quote = quoteSingle
				continue

			case '"':
				quote = quoteDouble
				continue
			}
		}

		if unicode.IsSpace(c) {
			continue
		}

		token.WriteRune(c)
	}

	if token.Len() > 0 {
		if hasDollar {
			expanded, err := expand(token.String())
			if err != nil {
				return nil, err
			}
			tokens = append(tokens, expanded)
		} else {
			tokens = append(tokens, token.String())
		}
	}

	token.Reset()
	args := &Args{
		args: tokens,
	}
	return args, nil
}

func normalizeArgs(args []string) []string {
	if len(args) == 0 {
		return args
	}

	normalized := make([]string, 0, len(args))
	for _, arg := range args {
		if len(arg) == 0 {
			continue
		}

		if (arg[0] == '"' && arg[len(arg)-1] == '"') || (arg[0] == '\'' && arg[len(arg)-1] == '\'') {
			arg = arg[1 : len(arg)-1]
		}

		normalized = append(normalized, arg)
	}

	return normalized
}
