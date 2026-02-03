package config

import (
	"os"
	"path/filepath"
	"strings"
)

func xdgConfigHome() string {
	if val := os.Getenv("XDG_CONFIG_HOME"); val != "" {
		return val
	}
	if home := userHome(); home != "" {
		return filepath.Join(home, ".config")
	}
	return ""
}

func xdgStateHome() string {
	if val := os.Getenv("XDG_STATE_HOME"); val != "" {
		return val
	}
	if home := userHome(); home != "" {
		return filepath.Join(home, ".local", "state")
	}
	return ""
}

func xdgDataHome() string {
	if val := os.Getenv("XDG_DATA_HOME"); val != "" {
		return val
	}
	if home := userHome(); home != "" {
		return filepath.Join(home, ".local", "share")
	}
	return ""
}

func DataHome() string {
	return xdgDataHome()
}

func userHome() string {
	if val := os.Getenv("HOME"); val != "" {
		return val
	}
	return ""
}

func expandUser(path string) string {
	path = strings.TrimSpace(path)
	if path == "" {
		return ""
	}
	if path == "~" {
		return userHome()
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(userHome(), strings.TrimPrefix(path, "~/"))
	}
	return path
}
