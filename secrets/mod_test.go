package secrets_test

import (
	"errors"
	"strings"
	"testing"
	"unicode"

	"github.com/stretchr/testify/assert"

	"github.com/voidbear-io/go/secrets"
)

func TestGenerate_DefaultOptions(t *testing.T) {
	s, err := secrets.Generate(16)
	assert.NoError(t, err)
	assert.Len(t, []rune(s), 16)
}

func TestGenerate_CustomOptions(t *testing.T) {
	s, err := secrets.Generate(12, secrets.WithLower(true), secrets.WithUpper(false), secrets.WithDigits(false), secrets.WithNoSymbols())
	assert.NoError(t, err)
	assert.Len(t, []rune(s), 12)
	for _, r := range s {
		assert.True(t, unicode.IsLower(r))
	}
}

func TestGenerate_WithSymbols(t *testing.T) {
	symbols := "!@#$"
	s, err := secrets.Generate(20, secrets.WithSymbols(symbols))
	assert.NoError(t, err)
	assert.Len(t, []rune(s), 20)
	found := false
	for _, r := range s {
		if strings.ContainsRune(symbols, r) {
			found = true
			break
		}
	}
	assert.True(t, found, "should contain at least one symbol")
}

func TestGenerate_WithChars(t *testing.T) {
	chars := "abc123"
	s, err := secrets.Generate(10, secrets.WithChars(chars))
	assert.NoError(t, err)
	assert.Len(t, []rune(s), 10)
	for _, r := range s {
		assert.Contains(t, chars, string(r))
	}
}

func TestGenerate_WithValidator(t *testing.T) {
	validator := func(runes []rune) error {
		for _, r := range runes {
			if r == 'x' {
				return nil
			}
		}
		return errors.New("must contain 'x'")
	}
	s, err := secrets.Generate(8, secrets.WithChars("xyz"), secrets.WithValidator(validator))
	assert.NoError(t, err)
	assert.Contains(t, s, "x")
}

func TestGenerate_NoCharSets(t *testing.T) {
	_, err := secrets.Generate(8, secrets.WithLower(false), secrets.WithUpper(false), secrets.WithDigits(false), secrets.WithSymbols(""))
	assert.Error(t, err)
}

func TestGenerateBytes(t *testing.T) {
	b, err := secrets.GenerateBytes(6)
	assert.NoError(t, err)
	assert.True(t, len(b) >= 6)
}

func TestOptionsBuilder(t *testing.T) {
	builder := secrets.NewOptionsBuilder().
		WithLower(true).
		WithUpper(false).
		WithDigits(true).
		WithSymbols("!").
		WithRetries(100)
	opts := builder.Build()
	assert.True(t, opts.Lower)
	assert.False(t, opts.Upper)
	assert.True(t, opts.Digits)
	assert.NotNil(t, opts.Symbols)
	assert.Equal(t, 100, opts.Retries)
}

func TestOptionsBuilderGenerate(t *testing.T) {
	builder := secrets.NewOptionsBuilder().
		WithLower(true).
		WithUpper(false).
		WithDigits(true).
		WithSymbols("!").
		WithSize(10).
		WithRetries(100)
	opts := builder.Build()
	s, err := opts.Generate()
	assert.NoError(t, err)
	assert.Len(t, []rune(s), 10)
}

func TestOptionsBuilderGenerateRuns(t *testing.T) {
	builder := secrets.NewOptionsBuilder().
		WithLower(true).
		WithUpper(false).
		WithDigits(true).
		WithSymbols("!").
		WithSize(10).
		WithRetries(100)
	opts := builder.Build()
	runes, err := opts.GenerateRunes()
	assert.NoError(t, err)
	assert.Len(t, runes, 10)
}

func TestOptionsBuilderGenerateBytes(t *testing.T) {
	builder := secrets.NewOptionsBuilder().
		WithLower(true).
		WithUpper(false).
		WithDigits(true).
		WithSymbols("!").
		WithSize(10).
		WithRetries(100)
	opts := builder.Build()
	b, err := opts.GenerateBytes()
	assert.NoError(t, err)
	assert.Len(t, b, 10)
}
