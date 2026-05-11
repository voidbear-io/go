package env

import (
	"os"
	"strings"
)

// Get retrieves the value of the environment variable named by key.
// It returns an empty string if the variable is not present.
func Get(key string) string {
	return os.Getenv(key)
}

// Set the value of the environment variable named by key to value.
// It returns an error if the variable cannot be set.
func Set(key, value string) error {
	return os.Setenv(key, value)
}

// Unset removes the environment variable named by key.
// It returns an error if the variable cannot be unset.
func Unset(key string) error {
	return os.Unsetenv(key)
}

// Determines if the environment variable named by key exists.
func Has(key string) bool {
	_, ok := os.LookupEnv(key)
	return ok
}

func All() map[string]string {
	kv := make(map[string]string)
	for _, e := range os.Environ() {
		pair := strings.Split(e, "=")
		if len(pair) == 2 && len(pair[1]) > 0 {
			kv[pair[0]] = pair[1]
		}
	}

	return kv
}

func GetPath() string {
	return Get(PATH)
}

func SetPath(value string) error {
	return Set(PATH, value)
}

func SplitPath() []string {
	return strings.Split(GetPath(), string(os.PathListSeparator))
}

func JoinPath(paths ...string) string {
	return strings.Join(paths, string(os.PathListSeparator))
}

func PrependPath(path string) error {
	paths := SplitPath()
	if len(paths) == 0 || paths[0] == "" {
		return SetPath(path)
	}

	if matchPath(paths[0], path) {
		return nil
	}

	paths = append([]string{path}, paths...)
	return SetPath(JoinPath(paths...))
}

func AppendPath(path string) error {
	paths := SplitPath()
	if len(paths) == 0 || (len(paths) == 1 && paths[0] == "") {
		return SetPath(path)
	}

	last := paths[len(paths)-1]
	if matchPath(last, path) {
		return nil
	}

	paths = append(paths, path)
	return SetPath(JoinPath(paths...))
}

func HasPath(path string) bool {
	paths := SplitPath()
	return hasPath(path, paths)
}
