package env_test

import (
	"errors"
	"os"
	"os/exec"
	"strings"
	"testing"

	"github.com/voidbear-io/go/env"
)

func TestExpand_BasicVariable(t *testing.T) {
	get := func(key string) string {
		if key == "FOO" {
			return "bar"
		}
		return ""
	}
	out, err := env.Expand("Value: ${FOO}", env.WithGet(get))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Value: bar" {
		t.Errorf("expected 'Value: bar', got '%s'", out)
	}
}

func TestExpand_DefaultValue(t *testing.T) {
	get := func(key string) string { return "" }
	out, err := env.Expand("Value: ${FOO:-default}", env.WithGet(get))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Value: default" {
		t.Errorf("expected 'Value: default', got '%s'", out)
	}
}

func TestExpand_NestedDefaultValue(t *testing.T) {
	get := func(key string) string { return "" }
	out, err := env.Expand("Value: ${FOO:-${BAR:-default}}", env.WithGet(get))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Value: default" {
		t.Errorf("expected 'Value: default', got '%s'", out)
	}
}

func TestExpand_EmptyVariableName(t *testing.T) {
	_, err := env.Expand("Value: ${}", env.WithGet(func(string) string { return "" }))
	if err == nil {
		t.Error("expected error for empty variable name")
	}
}

func TestExpand_InvalidVariableName(t *testing.T) {
	_, err := env.Expand("Value: ${1FOO}", env.WithGet(func(string) string { return "" }))
	if err == nil {
		t.Error("expected error for invalid variable name")
	}
}

func TestExpand_CommandSubstitution_Disabled(t *testing.T) {
	out, err := env.Expand("Echo: $(echo hi)", env.WithCommandSubstitution(false))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Echo: $(echo hi)" {
		t.Errorf("expected literal command substitution, got '%s'", out)
	}
}

func TestExpand_CommandSubstitution_Enabled(t *testing.T) {
	shPath, err := exec.LookPath("sh")
	if err != nil || shPath == "" {
		t.Skip("sh not found on path, skipping command substitution test")
	}

	out, err := env.Expand("Echo: $(echo hi)", env.WithCommandSubstitution(true), env.WithEnableShellExpansion(true), env.WithShell("sh"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Echo: hi" {
		t.Errorf("expected 'Echo: hi', got '%s'", out)
	}
}

func TestExpand_CommandSubstitution_NoShell(t *testing.T) {
	shPath, err := exec.LookPath("sh")
	if err != nil || shPath == "" {
		t.Skip("sh not found on path, skipping command substitution test")
	}

	out, err := env.Expand("Echo: $(echo hi)", env.WithCommandSubstitution(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Echo: hi" {
		t.Errorf("expected 'Echo: hi', got '%s'", out)
	}
}

func TestExpand_WindowsVariable(t *testing.T) {
	get := func(key string) string {
		if key == "FOO" {
			return "bar"
		}
		return ""
	}
	out, err := env.Expand("Value: %FOO%", env.WithGet(get), env.WithExpandWindowsVars(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Value: bar" {
		t.Errorf("expected 'Value: bar', got '%s'", out)
	}
}

func TestExpand_WindowsVariable_EmptyName(t *testing.T) {
	_, err := env.Expand("Value: %%", env.WithExpandWindowsVars(true))
	if err == nil {
		t.Fatalf("expected error for empty variable name, got nil")
	}
}

func TestExpand_BashVariable(t *testing.T) {
	get := func(key string) string {
		if key == "FOO" {
			return "baz"
		}
		return ""
	}
	out, err := env.Expand("Value: $FOO", env.WithGet(get))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Value: baz" {
		t.Errorf("expected 'Value: baz', got '%s'", out)
	}
}

func TestExpand_EscapedDollar(t *testing.T) {
	out, err := env.Expand("Price: \\$100")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Price: $100" {
		t.Errorf("expected 'Price: $100', got '%s'", out)
	}
}

func TestExpand_DoubleDollar(t *testing.T) {
	out, err := env.Expand("PID: $$")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "PID: $" {
		t.Errorf("expected 'PID: $', got '%s'", out)
	}
}

func TestExpand_BashVariableWithDefault(t *testing.T) {
	get := func(key string) string { return "" }
	out, err := env.Expand("Value: $FOO:-default", env.WithGet(get))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// Since $FOO:-default is not a valid bash interpolation, it should be literal :-default
	if out != "Value: :-default" {
		t.Errorf("expected literal, got '%s'", out)
	}
}

func TestExpand_BashInterpolationWithSet(t *testing.T) {
	var setCalled bool
	get := func(key string) string { return "" }
	set := func(key, value string) error {
		if key == "FOO" && value == "bar" {
			setCalled = true
			return nil
		}
		return errors.New("unexpected set")
	}
	out, err := env.Expand("Value: ${FOO:=bar}", env.WithGet(get), env.WithSet(set))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Value: bar" {
		t.Errorf("expected 'Value: bar', got '%s'", out)
	}
	if !setCalled {
		t.Error("expected set to be called")
	}
}

func TestExpand_BashInterpolationWithMessage(t *testing.T) {
	get := func(key string) string { return "" }
	_, err := env.Expand("Value: ${FOO:?missing}", env.WithGet(get))
	if err == nil || err.Error() != "missing" {
		t.Errorf("expected error 'missing', got '%v'", err)
	}
}

func TestExpand_ExpandUnixArgs(t *testing.T) {
	origArgs := os.Args
	defer func() { os.Args = origArgs }()
	os.Args = []string{"cmd", "first", "second"}
	out, err := env.Expand("Arg1: ${1}", env.WithExpandUnixArgs(true))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out != "Arg1: first" {
		t.Errorf("expected 'Arg1: first', got '%s'", out)
	}
}

func TestExpand_CustomExpander_ReplacesSchemesInOutput(t *testing.T) {
	get := func(key string) string {
		if key == "REF" {
			return "akv://vault/mysecret"
		}
		return ""
	}
	custom := func(s string) (string, error) {
		return strings.ReplaceAll(s, "akv://vault/mysecret", "supersecret"), nil
	}
	out, err := env.Expand("x=${REF} y=akv://vault/mysecret", env.WithGet(get), env.WithCustomExpander(custom))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	want := "x=supersecret y=supersecret"
	if out != want {
		t.Errorf("expected %q, got %q", want, out)
	}
}

func TestExpand_CustomExpander_Error(t *testing.T) {
	custom := func(string) (string, error) {
		return "", errors.New("boom")
	}
	_, err := env.Expand("hello", env.WithCustomExpander(custom))
	if err == nil || err.Error() != "boom" {
		t.Fatalf("expected boom, got %v", err)
	}
}
