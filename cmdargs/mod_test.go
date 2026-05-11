package cmdargs_test

import (
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/voidbear-io/go/cmdargs"
)

func TestNewAndToArray(t *testing.T) {
	args := []string{"a", "b", "c"}
	a := cmdargs.New(args)
	arr := a.ToArray()
	if !reflect.DeepEqual(arr[len(arr)-len(args):], args) {
		t.Errorf("ToArray() = %v, want %v", arr, args)
	}
}

func TestLenAndGet(t *testing.T) {
	a := cmdargs.New([]string{"x", "y"})
	if a.Len() != 2 {
		t.Errorf("Len() = %d, want 2", a.Len())
	}
	if a.Get(0) != "x" || a.Get(1) != "y" {
		t.Errorf("Get() failed: got %q, %q", a.Get(0), a.Get(1))
	}
	if a.Get(-1) != "" || a.Get(2) != "" {
		t.Errorf("Get() out of bounds should return empty string")
	}
}

func TestIndexAndIndexFold(t *testing.T) {
	a := cmdargs.New([]string{"foo", "Bar", "baz"})
	if a.Index("Bar") != 1 {
		t.Errorf("Index() = %d, want 1", a.Index("Bar"))
	}
	if a.Index("notfound") != -1 {
		t.Errorf("Index() = %d, want -1", a.Index("notfound"))
	}
	if a.IndexFold("bAz") != 2 {
		t.Errorf("IndexFold() = %d, want 2", a.IndexFold("bAz"))
	}
}

func TestIndexAnyAndIndexAnyFold(t *testing.T) {
	a := cmdargs.New([]string{"foo", "Bar", "baz"})
	if a.IndexAny([]string{"baz", "Bar"}) != 1 {
		t.Errorf("IndexAny() = %d, want 1", a.IndexAny([]string{"baz", "Bar"}))
	}
	if a.IndexAnyFold([]string{"BAZ", "BAR"}) != 1 {
		t.Errorf("IndexAnyFold() = %d, want 1", a.IndexAnyFold([]string{"BAZ", "BAR"}))
	}
}

func TestContainsAndContainsFold(t *testing.T) {
	a := cmdargs.New([]string{"foo", "Bar"})
	if !a.Contains("foo") {
		t.Errorf("Contains() should be true")
	}
	if a.Contains("baz") {
		t.Errorf("Contains() should be false")
	}
	if !a.ContainsFold("BAR") {
		t.Errorf("ContainsFold() should be true")
	}
}

func TestContainsAnyAndContainsAnyFold(t *testing.T) {
	a := cmdargs.New([]string{"foo", "Bar"})
	if !a.ContainsAny([]string{"baz", "Bar"}) {
		t.Errorf("ContainsAny() should be true")
	}
	if !a.ContainsAnyFold([]string{"BAZ", "BAR"}) {
		t.Errorf("ContainsAnyFold() should be true")
	}
}

func TestSet(t *testing.T) {
	a := cmdargs.New([]string{"a", "b"})
	a.Set(1, "c")
	if a.Get(1) != "c" {
		t.Errorf("Set() failed, got %q", a.Get(1))
	}
	a.Set(-1, "x")
	a.Set(2, "y") // out of bounds, should not panic
}

func TestPushAppendPrepend(t *testing.T) {
	a := cmdargs.New([]string{"a"})
	a.Push("b", "c")

	copy := a.ToArray()
	assert.Equal(t, copy, []string{"a", "b", "c"})
	if !reflect.DeepEqual(copy, []string{"a", "b", "c"}) {
		t.Errorf("Push() failed: %v", copy)
	}
	a.Append("d")
	if !reflect.DeepEqual(a.ToArray(), []string{"a", "b", "c", "d"}) {
		t.Errorf("Append() failed: %v", a.ToArray())
	}
	a.Prepend("z")
	if !reflect.DeepEqual(a.ToArray(), []string{"z", "a", "b", "c", "d"}) {
		t.Errorf("Prepend() failed: %v", a.ToArray())
	}
}

func TestShiftAndPop(t *testing.T) {
	a := cmdargs.New([]string{"x", "y"})
	val := a.Shift()
	if val != "x" || !reflect.DeepEqual(a.ToArray(), []string{"y"}) {
		t.Errorf("Shift() failed: val=%q, args=%v", val, a.ToArray())
	}
	val = a.Pop()
	if val != "y" || len(a.ToArray()) != 0 {
		t.Errorf("Pop() failed: val=%q, args=%v", val, a.ToArray())
	}
	val = a.Pop()
	if val != "" {
		t.Errorf("Pop() on empty should return empty string")
	}
}

func TestRemoveAndRemoveAt(t *testing.T) {
	a := cmdargs.New([]string{"a", "b", "c"})
	ok := a.Remove("b")
	if !ok || !reflect.DeepEqual(a.ToArray(), []string{"a", "c"}) {
		t.Errorf("Remove() failed: %v", a.ToArray())
	}
	ok = a.RemoveAt(1)
	if !ok || !reflect.DeepEqual(a.ToArray(), []string{"a"}) {
		t.Errorf("RemoveAt() failed: %v", a.ToArray())
	}
	ok = a.RemoveAt(5)
	if ok {
		t.Errorf("RemoveAt() out of bounds should return false")
	}
}

func TestString(t *testing.T) {
	a := cmdargs.New([]string{"foo", "bar baz", `"quoted"`})
	s := a.String()
	if s != `foo "bar baz" quoted` {
		t.Errorf("String() = %q", s)
	}
}

func TestSplit(t *testing.T) {
	tests := []struct {
		input string
		want  []string
	}{
		{"a b c", []string{"a", "b", "c"}},
		{`"a b" c`, []string{"a b", "c"}},
		{`'x y' z`, []string{"x y", "z"}},
		{`foo "bar baz" 'qux'`, []string{"foo", "bar baz", "qux"}},
		{`foo "bar \"baz\""`, []string{"foo", `bar \"baz\"`}},

		// new line should not terminate if preceeded by backslash
		{"a b\\\nc", []string{"a", "b", "c"}},

		// simulates bash style continuation
		// command a b \
		// c
		{"a b \\\nc", []string{"a", "b", "c"}},
		{"a b \\\r\nc", []string{"a", "b", "c"}},
		{"a b\nc", []string{"a", "b"}},
		{"a b\r\nc", []string{"a", "b"}},
	}
	for _, tt := range tests {
		got := cmdargs.Split(tt.input).ToArray()
		if !assert.Equal(t, tt.want, got) {
			t.Errorf("Split(%q) = %v, want %v", tt.input, got, tt.want)
		}
	}
}
func TestSplitAndExpand(t *testing.T) {
	expand := func(s string) (string, error) {
		// Simple expansion: replace $FOO with "foo", $BAR with "bar"

		if strings.Contains(s, "$ERR") {
			return "", errors.New("bad sub")
		}

		values := map[string]string{
			"$FOO": "foo",
			"$BAR": "bar",
		}

		for k, v := range values {
			s = strings.ReplaceAll(s, k, v)
		}

		return s, nil
	}

	tests := []struct {
		input    string
		want     []string
		wantErr  bool
		expandFn func(string) (string, error)
	}{
		// No expansion
		{"a b c", []string{"a", "b", "c"}, false, expand},
		// Expansion in double quotes
		{`"foo$FOO" bar`, []string{"foofoo", "bar"}, false, expand},
		// Expansion in double quotes, multiple tokens
		{`"bar$BAR" baz`, []string{"barbar", "baz"}, false, expand},
		// Expansion in double quotes, with spaces
		{`"$FOO $BAR"`, []string{"foo bar"}, false, expand},
		// Expansion in single quotes (should not expand)
		{`'foo$FOO'`, []string{"foo$FOO"}, false, expand},
		// Expansion in unquoted token
		{`foo$FOO`, []string{"foofoo"}, false, expand},
		// Expansion error
		{`"$ERR"`, nil, true, expand},
		// Escaped newlines and expansion
		{"foo$FOO \\\nbar$BAR", []string{"foofoo", "barbar"}, false, expand},
		// Expansion with no $ present
		{`foo bar`, []string{"foo", "bar"}, false, expand},
	}

	for _, tt := range tests {
		got, err := cmdargs.SplitAndExpand(tt.input, tt.expandFn)
		if tt.wantErr {
			assert.Error(t, err, "SplitAndExpand(%q) expected error", tt.input)
		} else {
			assert.NoError(t, err, "SplitAndExpand(%q) unexpected error: %v", tt.input, err)
			assert.Equal(t, tt.want, got.ToArray(), "SplitAndExpand(%q)", tt.input)
		}
	}
}

func TestGetAnyAndGetAnyOr(t *testing.T) {
	a := cmdargs.New([]string{"-f", "Foo", "bar"})
	v, ok := a.GetAny("-f", "-v")
	assert.True(t, ok)
	assert.Equal(t, "-f", v)

	v, ok = a.GetAny("-v", "-x")
	assert.False(t, ok)
	assert.Equal(t, "", v)

	// IndexAny is case-sensitive; "foo" does not match "Foo"
	_, ok = a.GetAny("foo")
	assert.False(t, ok)

	v, ok = a.GetAnyOr("-f")
	assert.True(t, ok)
	assert.Equal(t, "-f", v)

	_, ok = a.GetAnyOr("-missing")
	assert.False(t, ok)
}

func TestGetIntAndVariants(t *testing.T) {
	a := cmdargs.New([]string{"42", "nope", "-7"})

	n, ok := a.GetInt(0)
	assert.True(t, ok)
	assert.Equal(t, 42, n)

	_, ok = a.GetInt(1)
	assert.False(t, ok)

	n, ok = a.GetInt(2)
	assert.True(t, ok)
	assert.Equal(t, -7, n)

	_, ok = a.GetInt(-1)
	assert.False(t, ok)

	assert.Equal(t, 99, a.GetIntOr(1, 99))
	assert.Equal(t, 42, a.GetIntOr(0, 99))

	n, ok = a.GetIntAny("42", "nope")
	assert.True(t, ok)
	assert.Equal(t, 42, n)

	n, ok = a.GetIntAny("-7")
	assert.True(t, ok)
	assert.Equal(t, -7, n)

	_, ok = a.GetIntAny("missing")
	assert.False(t, ok)
}

func TestGetFloatAndVariants(t *testing.T) {
	a := cmdargs.New([]string{"3.5", "x", "1e2"})

	f, ok := a.GetFloat(0)
	assert.True(t, ok)
	assert.InDelta(t, 3.5, f, 1e-9)

	_, ok = a.GetFloat(1)
	assert.False(t, ok)

	f, ok = a.GetFloat(2)
	assert.True(t, ok)
	assert.InDelta(t, 100, f, 1e-9)

	assert.InDelta(t, 2.5, a.GetFloatOr(1, 2.5), 1e-9)

	f, ok = a.GetFloatAny("3.5")
	assert.True(t, ok)
	assert.InDelta(t, 3.5, f, 1e-9)

	_, ok = a.GetFloatAny("x")
	assert.False(t, ok)
}

func TestGetBoolAndVariants(t *testing.T) {
	a := cmdargs.New([]string{"true", "0", "yes", "maybe"})

	b, ok := a.GetBool(0)
	assert.True(t, ok)
	assert.True(t, b)

	b, ok = a.GetBool(1)
	assert.True(t, ok)
	assert.False(t, b)

	b, ok = a.GetBool(2)
	assert.True(t, ok)
	assert.True(t, b)

	_, ok = a.GetBool(3)
	assert.False(t, ok)

	assert.True(t, a.GetBoolOr(3, true))
	assert.False(t, a.GetBoolOr(3, false))

	b, ok = a.GetBoolAny("yes", "nope")
	assert.True(t, ok)
	assert.True(t, b)

	_, ok = a.GetBoolAny("maybe")
	assert.False(t, ok)
}

func TestGetSliceAndGetSliceAny(t *testing.T) {
	a := cmdargs.New([]string{"prog", "-e", "value", "-e", "value2", "tail"})
	got, ok := a.GetSlice("-e")
	assert.True(t, ok)
	assert.Equal(t, []string{"value", "value2"}, got)

	got, ok = a.GetSliceAny("-e", "--env")
	assert.True(t, ok)
	assert.Equal(t, []string{"value", "value2"}, got)

	a = cmdargs.New([]string{"--env", "x", "-E", "y"})
	got, ok = a.GetSliceAny("-e", "--env")
	assert.True(t, ok)
	assert.Equal(t, []string{"x", "y"}, got)

	a = cmdargs.New([]string{"-e", "only"})
	got, ok = a.GetSlice("-e")
	assert.True(t, ok)
	assert.Equal(t, []string{"only"}, got)

	_, ok = cmdargs.New([]string{"-e"}).GetSlice("-e")
	assert.False(t, ok)

	_, ok = cmdargs.New([]string{"other"}).GetSlice("-e")
	assert.False(t, ok)

	_, ok = cmdargs.New([]string{}).GetSliceAny()
	assert.False(t, ok)

	a = cmdargs.New([]string{"-e", "only"})
	got, ok = a.GetSlice("-e")
	assert.True(t, ok)
	got[0] = "mutated"
	assert.Equal(t, []string{"-e", "only"}, a.ToArray())
}

func TestGetMapAndGetMapAny(t *testing.T) {
	a := cmdargs.New([]string{"cmd", "-e", "key=value", "-e", "key2=value2"})
	m, ok := a.GetMap("-e")
	assert.True(t, ok)
	assert.Equal(t, map[string]string{"key": "value", "key2": "value2"}, m)

	a = cmdargs.New([]string{"-D", "a=1", "--define", "b=c=d"})
	m, ok = a.GetMapAny("-D", "--define")
	assert.True(t, ok)
	assert.Equal(t, map[string]string{"a": "1", "b": "c=d"}, m)

	a = cmdargs.New([]string{"-e", "k=v", "-e", "k=new", "-e", "onlytoken"})
	m, ok = a.GetMap("-e")
	assert.True(t, ok)
	assert.Equal(t, map[string]string{"k": "new"}, m)

	_, ok = cmdargs.New([]string{"-e", "nope"}).GetMap("-e")
	assert.False(t, ok)

	_, ok = cmdargs.New([]string{"-e"}).GetMap("-e")
	assert.False(t, ok)

	_, ok = cmdargs.New([]string{}).GetMapAny()
	assert.False(t, ok)

	m, ok = cmdargs.New([]string{"-e", "=emptykey", "-e", "x="}).GetMap("-e")
	assert.True(t, ok)
	assert.Equal(t, map[string]string{"": "emptykey", "x": ""}, m)
}

func TestSetValueSetIntSetFloatSetBool(t *testing.T) {
	t.Run("SetInt replaces token in place when not last", func(t *testing.T) {
		a := cmdargs.New([]string{"-v", "old", "tail"})
		assert.True(t, a.SetInt("-v", 9))
		assert.Equal(t, []string{"9", "old", "tail"}, a.ToArray())
	})

	t.Run("SetInt appends when option is last element", func(t *testing.T) {
		a := cmdargs.New([]string{"-v"})
		assert.True(t, a.SetInt("-v", 3))
		assert.Equal(t, []string{"-v", "3"}, a.ToArray())
	})

	t.Run("SetValue middle token uses insert semantics", func(t *testing.T) {
		a := cmdargs.New([]string{"-v", "old", "tail"})
		assert.True(t, a.SetValue("-v", "newval"))
		assert.Equal(t, []string{"newval", "-v", "old", "tail"}, a.ToArray())
	})

	t.Run("SetValue last token appends value", func(t *testing.T) {
		a := cmdargs.New([]string{"-v"})
		assert.True(t, a.SetValue("-v", "only"))
		assert.Equal(t, []string{"-v", "only"}, a.ToArray())
	})

	t.Run("SetFloat and SetBool use SetValue", func(t *testing.T) {
		a := cmdargs.New([]string{"f", "b", "z"})
		assert.True(t, a.SetFloat("f", 2.5))
		assert.True(t, a.SetBool("b", false))
		assert.Equal(t, []string{"2.5", "f", "false", "b", "z"}, a.ToArray())
	})

	t.Run("unknown option", func(t *testing.T) {
		a := cmdargs.New([]string{"x"})
		assert.False(t, a.SetValue("--nope", "v"))
		assert.False(t, a.SetInt("--nope", 1))
		assert.False(t, a.SetFloat("--nope", 1))
		assert.False(t, a.SetBool("--nope", true))
		assert.Equal(t, []string{"x"}, a.ToArray())
	})
}
