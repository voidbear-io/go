package env

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/voidbear-io/go/cmdargs"
)

const (
	// DefaultExpandOptions is the default options for Expand
	none                = 0
	windows             = 1
	bashVariable        = 2
	bashInterpolation   = 3
	commandSubstitution = 4
	windowsVariable     = 5
)

type ExpandOptions struct {
	// If true, windows style environment variables will be expanded
	Get                  func(string) string
	Set                  func(string, string) error
	Keys                 []string
	ExpandUnixArgs       bool
	ExpandWindowsVars    bool
	CommandSubstitution  bool
	EnableShellExpansion bool
	UseShell             string
	ShellArgs            []string
	// CustomExpander, if non-nil, is invoked on the fully expanded string so you can
	// scan and rewrite fragments (e.g. akv://vault/key → a resolved secret). It runs once
	// after standard expansion, covering both substituted values and literal segments.
	CustomExpander func(string) (string, error)
}

type ExpandOption func(*ExpandOptions)

func WithGet(f func(string) string) ExpandOption {
	return func(o *ExpandOptions) {
		o.Get = f
	}
}

func WithSet(f func(string, string) error) ExpandOption {
	return func(o *ExpandOptions) {
		o.Set = f
	}
}

func WithExpandUnixArgs(expand bool) ExpandOption {
	return func(o *ExpandOptions) {
		o.ExpandUnixArgs = expand
	}
}

func WithExpandWindowsVars(expand bool) ExpandOption {
	return func(o *ExpandOptions) {
		o.ExpandWindowsVars = expand
	}
}

func WithCommandSubstitution(enable bool) ExpandOption {
	return func(o *ExpandOptions) {
		o.CommandSubstitution = enable
	}
}

func WithEnableShellExpansion(enable bool) ExpandOption {
	return func(o *ExpandOptions) {
		o.EnableShellExpansion = enable
	}
}

func WithShell(shell string) ExpandOption {
	return func(o *ExpandOptions) {
		o.UseShell = shell
	}
}

// WithCustomExpander sets a hook that receives the full expansion result and can replace
// patterns such as custom URI schemes in environment-backed or literal text.
func WithCustomExpander(f func(string) (string, error)) ExpandOption {
	return func(o *ExpandOptions) {
		o.CustomExpander = f
	}
}

func interpolateVar(token string, o *ExpandOptions) (string, error) {
	key := token
	defaultValue := ""
	message := ""
	if strings.Contains(token, ":-") {

		parts := split(token, ":-")
		key = parts[0]
		if len(parts) > 1 {
			defaultValue = parts[1]
		}
	} else if strings.Contains(token, ":=") {
		parts := split(token, ":=")
		key = parts[0]
		if len(parts) > 1 {
			defaultValue = parts[1]
		}
		if !isValidBashVariable([]rune(key)) {
			return "", errors.New("invalid bash variable syntax: invalid variable name")
		}

		if len(key) > 0 {
			v := o.Get(key)
			if len(v) == 0 {
				if err := o.Set(key, defaultValue); err != nil {
					return "", err
				}
				hasKey := false
				for _, k := range o.Keys {
					if k == key {
						hasKey = true
						break
					}
				}
				if !hasKey {
					o.Keys = append(o.Keys, key)
				}
			}
		}
	} else if strings.Contains(token, ":?") {
		parts := split(token, ":?")
		key = parts[0]
		if len(parts) > 1 {
			message = parts[1]
		}
	} else if strings.Contains(token, ":") {
		parts := split(token, ":")
		key = parts[0]
		if len(parts) > 1 {
			defaultValue = parts[1]
		}
	}

	if len(key) == 0 {
		return "", errors.New("invalid bash variable syntax: empty variable name")
	}

	if o.ExpandUnixArgs {
		i, err := strconv.Atoi(key)
		if err == nil {
			if len(os.Args) > i {
				return os.Args[i], nil
			}
			return "", nil
		}
	}

	if !isValidBashVariable([]rune(key)) {
		return "", errors.New("invalid bash variable syntax: invalid variable name")
	}

	value := o.Get(key)
	if len(value) == 0 {
		if len(defaultValue) > 0 {
			if strings.Contains(defaultValue, "$") {
				next, err := ExpandWithOptions(defaultValue, o)
				if err != nil {
					return "", err
				}
				return next, nil
			}

			return defaultValue, nil
		}

		if len(message) > 0 {
			return "", errors.New(message)
		}

		return "", nil
	}

	return value, nil
}

func split(s string, value string) []string {
	index := strings.Index(s, value)
	if index == -1 {
		return []string{s}
	}

	first := s[:index]
	second := s[index+(len(value)):]
	if first == "" {
		return []string{second}
	}
	if second == "" {
		return []string{first}
	}
	return []string{first, second}
}

func isLetterOrDigit(r rune) bool {
	return unicode.IsLetter(r) || unicode.IsDigit(r)
}

func isValidBashVariable(input []rune) bool {
	for i, c := range input {
		if i == 0 && !unicode.IsLetter(c) && c != '_' {
			return false
		}

		if !unicode.IsLetter(c) && !unicode.IsDigit(c) && c != '_' {
			return false
		}
	}

	return true
}

func ExpandWithOptions(input string, options *ExpandOptions) (string, error) {
	if options.Get == nil {
		options.Get = Get
	}

	if options.Set == nil {
		options.Set = Set
	}

	o := options
	kind := none
	min := rune(0)
	remaining := len(input)
	l := len(input)
	runes := []rune(input)
	output := strings.Builder{}
	token := strings.Builder{}
	bracketCount := 0
	for i := 0; i < l; i++ {
		remaining--
		c := runes[i]

		if kind == none {
			z := i + 1
			next := min
			if z < l {
				next = runes[z]
			}

			if c == '$' && next == '$' {
				output.WriteRune('$')
				i++
				remaining--
				continue
			}

			if c == '\\' && next == '$' {
				output.WriteRune('$')
				i++
				remaining--
				continue
			}

			if c == '$' {
				if o.CommandSubstitution && next == '(' {
					kind = commandSubstitution
					i++
					remaining--
					continue
				}

				if next == '{' {
					kind = bashInterpolation
					i++
					remaining--
					continue
				}

				if isLetterOrDigit(next) || next == '_' {
					kind = bashVariable
					continue
				}
			}

			if o.ExpandWindowsVars && c == '%' {
				kind = windowsVariable
				continue
			}

			output.WriteRune(c)
			continue
		}

		if o.ExpandWindowsVars && kind == windowsVariable && c == '%' {
			if token.Len() == 0 {
				return "", errors.New("invalid windows variable syntax: empty variable name")
			}

			interpolation := token.String()
			token.Reset()
			value := o.Get(interpolation)
			output.WriteString(value)
			kind = none
			continue
		}

		if kind == bashInterpolation {
			if c == '{' {
				bracketCount++
				token.WriteRune(c)
				continue
			}

			if c == '}' {
				if bracketCount > 0 {
					bracketCount--
					token.WriteRune(c)
					continue
				}

				if token.Len() == 0 {
					return "", errors.New("invalid bash variable syntax: empty variable name")
				}

				interpolation := token.String()
				token.Reset()
				value, err := interpolateVar(interpolation, o)
				if err != nil {
					return "", err
				}

				output.WriteString(value)
				kind = none
				continue
			}
		}

		if kind == commandSubstitution && c == ')' {

			expression := token.String()
			token.Reset()

			if len(expression) == 0 {
				return "", errors.New("invalid command substitution: empty expression")
			}

			if !o.EnableShellExpansion {
				commandArgs, err := cmdargs.SplitAndExpand(expression, func(s string) (string, error) {
					return ExpandWithOptions(s, o)
				})
				if err != nil {
					return "", fmt.Errorf("command substitution failed to parse: %w", err)
				}

				if commandArgs.Len() == 0 {
					return "", errors.New("invalid command substitution: empty command")
				}

				exe := commandArgs.Get(0)
				commandArgs.RemoveAt(0)

				cmd := exec.Command(exe, commandArgs.ToArray()...)
				if len(o.Keys) > 0 {
					envVars := os.Environ()
					for _, k := range o.Keys {
						v := o.Get(k)
						envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
					}
					cmd.Env = envVars
				}

				var outb, errb bytes.Buffer
				cmd.Stdout = &outb
				cmd.Stderr = &errb
				err = cmd.Start()
				if err != nil {
					return "", err
				}

				err = cmd.Wait()
				if err != nil {
					return "", err
				}

				ec := cmd.ProcessState.ExitCode()
				if ec != 0 {
					return "", errors.New("command substitution failed with exit code " + fmt.Sprintf("%d", ec) + ": " + errb.String())
				}

				output.WriteString(strings.TrimSpace(outb.String()))
				kind = none
				continue
			}

			if o.UseShell == "" {
				if runtime.GOOS == "windows" {
					o.UseShell = "powershell.exe"
				} else {
					o.UseShell = "bash"
				}
			}

			shellArgs := o.ShellArgs
			if len(shellArgs) == 0 {
				switch o.UseShell {
				case "powershell.exe":
					fallthrough
				case "powershell":
					fallthrough
				case "pwsh.exe":
					fallthrough
				case "pwsh":
					shellArgs = []string{
						"-NoLogo",
						"-NoProfile",
						"-NonInteractive",
						"-ExecutionPolicy",
						"Bypass",
						"-Command",
					}
				case "bash":
					shellArgs = []string{
						"-noprofile",
						"--norc",
						"-e",
						"-o",
						"pipefail",
						"-c",
					}
				case "sh":
					shellArgs = []string{
						"-e",
						"-c",
					}
				default:
					shellArgs = []string{}
				}
			}

			shellArgs = append(shellArgs, expression)

			cmd := exec.Command(o.UseShell, shellArgs...)
			if len(o.Keys) > 0 {
				envVars := os.Environ()
				for _, k := range o.Keys {
					v := o.Get(k)
					envVars = append(envVars, fmt.Sprintf("%s=%s", k, v))
				}
				cmd.Env = envVars
			}
			var outb, errb bytes.Buffer
			cmd.Stdout = &outb
			cmd.Stderr = &errb
			err := cmd.Start()
			if err != nil {
				return "", err
			}
			err = cmd.Wait()
			if err != nil {
				return "", err
			}
			ec := cmd.ProcessState.ExitCode()
			if ec != 0 {
				return "", errors.New("command substitution failed with exit code " + fmt.Sprintf("%d", ec) + ": " + errb.String())
			}
			output.WriteString(strings.TrimSpace(outb.String()))
			kind = none
			continue
		}

		if kind == bashVariable && ((!isLetterOrDigit(c) && c != '_') || remaining == 0) {
			shouldAppend := c != '\\'
			if remaining == 0 && (isLetterOrDigit(c) || c == '_') {
				token.WriteRune(c)
				shouldAppend = false
			}

			if c == '$' {
				shouldAppend = false
				i--
			}

			key := token.String()
			if len(key) == 0 {
				return "", errors.New("invalid bash variable syntax: empty variable name")
			}

			if o.ExpandUnixArgs {
				i, err := strconv.Atoi(key[1:])
				if err == nil {
					if len(os.Args) > i {
						output.WriteString(os.Args[i])
					} else {
						output.WriteString("")
					}

					if shouldAppend {
						output.WriteRune(c)
					}

					token.Reset()
					kind = none
					continue
				}
			}

			if !isValidBashVariable([]rune(key)) {
				return "", errors.New("invalid bash variable syntax: invalid variable name")
			}

			value := o.Get(key)
			if len(value) > 0 {
				output.WriteString(value)
			}

			if shouldAppend {
				output.WriteRune(c)
			}

			token.Reset()
			kind = none
			continue
		}

		token.WriteRune(c)
		if remaining == 0 {
			if kind == bashInterpolation || kind == commandSubstitution || kind == bashVariable {
				return "", errors.New("invalid bash variable syntax: missing closing brace or parenthesis")
			}
		}
	}

	out := output.String()
	output.Reset()

	if o.CustomExpander != nil {
		var err error
		out, err = o.CustomExpander(out)
		if err != nil {
			return "", err
		}
	}

	return out, nil
}

func Expand(input string, options ...ExpandOption) (string, error) {
	ops := &ExpandOptions{
		Get:                  Get,
		Set:                  Set,
		ExpandUnixArgs:       true,
		ExpandWindowsVars:    false,
		CommandSubstitution:  false,
		EnableShellExpansion: false,
		UseShell:             "",
	}

	for _, opt := range options {
		opt(ops)
	}

	return ExpandWithOptions(input, ops)
}
