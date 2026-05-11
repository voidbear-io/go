//go:build aix || darwin || dragonfly || freebsd || hurd || illumos || ios || linux || netbsd || openbsd || plan9 || solaris || zos
// +build aix darwin dragonfly freebsd hurd illumos ios linux netbsd openbsd plan9 solaris zos

package env

const (
	// The path variable name for the current OS.
	PATH = "PATH"
	// The home directory variable name for the current OS.
	HOME = "HOME"
	// The host name variable name for the current OS.
	HOSTNAME = "HOSTNAME"
	// The user name variable name for the current OS.
	USER = "USER"
	// The temporary directory for the current user. The variable
	// may not be defined on all systems.
	TMP = "TMPDIR"
	// The home config directory for the current user. The variable
	// may not be defined on all systems.
	HOME_CONFIG = "XDG_CONFIG_HOME"
	// The home data directory for the current user. The variable
	// may not be defined on all systems.
	HOME_DATA = "XDG_DATA_HOME"
	// The home cache directory for the current user. The variable
	// may not be defined on all systems.
	HOME_CACHE = "XDG_CACHE_HOME"
)

func hasPath(path string, paths []string) bool {
	for _, p := range paths {
		if p == path {
			return true
		}
	}
	return false
}

func matchPath(left string, right string) bool {
	return left == right
}
