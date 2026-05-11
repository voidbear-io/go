//go:build windows

package env

import "strings"

const (
	// The path variable name for the current OS.
	PATH = "Path"
	// The home directory variable name for the current OS.
	HOME = "UserProfile"
	// The host name variable name for the current OS.
	HOSTNAME = "COMPUTERNAME"
	// The user name variable name for the current OS.
	USER = "USERNAME"
	// The temporary directory for the current user. The variable
	// may not be defined on all systems.
	TMP = "TEMP"
	// The home config directory for the current user. The variable
	// may not be defined on all systems.
	HOME_CONFIG = "AppData"
	// The home data directory for the current user. The variable
	// may not be defined on all systems.
	HOME_DATA = "LocalAppData"
	// The home cache directory for the current user. The variable
	// may not be defined on all systems.
	HOME_CACHE = "LocalAppData"
)

func hasPath(path string, paths []string) bool {
	for _, p := range paths {
		if strings.EqualFold(p, path) {
			return true
		}
	}
	return false
}

func matchPath(left string, right string) bool {
	return strings.EqualFold(left, right)
}
