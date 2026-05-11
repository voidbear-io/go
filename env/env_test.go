package env_test

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/voidbear-io/go/env"
)

func TestGetSetUnsetHas(t *testing.T) {
	key := "ENV_TEST_KEY"
	value := "test_value"

	// Ensure key is unset
	_ = env.Unset(key)
	assert.False(t, env.Has(key))
	assert.Equal(t, "", env.Get(key))

	// Set and get
	err := env.Set(key, value)
	assert.NoError(t, err)
	assert.True(t, env.Has(key))
	assert.Equal(t, value, env.Get(key))

	// Unset
	err = env.Unset(key)
	assert.NoError(t, err)
	assert.False(t, env.Has(key))
	assert.Equal(t, "", env.Get(key))
}

func TestAll(t *testing.T) {
	key := "ENV_ALL_TEST"
	value := "all_value"
	_ = env.Set(key, value)
	defer func() {
		assert.NoError(t, env.Unset(key))
	}()

	all := env.All()
	assert.Contains(t, all, key)
	assert.Equal(t, value, all[key])
}

func TestPathFunctions(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer func() {
		assert.NoError(t, os.Setenv("PATH", origPath))
	}()

	testPath := "/tmp/testpath"
	_ = env.SetPath(testPath)
	assert.Equal(t, testPath, env.GetPath())
	assert.Equal(t, []string{testPath}, env.SplitPath())

	joined := env.JoinPath("/a", "/b")
	assert.Contains(t, joined, "/a")
	assert.Contains(t, joined, "/b")
}

func TestPrependAppendHasPath(t *testing.T) {
	origPath := os.Getenv("PATH")
	defer func() {
		assert.NoError(t, os.Setenv("PATH", origPath))
	}()

	_ = env.SetPath("")
	assert.False(t, env.HasPath("/foo"))

	_ = env.PrependPath("/foo")
	assert.True(t, env.HasPath("/foo"))
	assert.Equal(t, "/foo", env.SplitPath()[0])

	_ = env.AppendPath("/bar")
	paths := env.SplitPath()
	assert.Equal(t, "/bar", paths[len(paths)-1])
	assert.True(t, env.HasPath("/bar"))
}
