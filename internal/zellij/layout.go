package zellij

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const defaultStatusInterval = 1000

// LayoutOptions controls the generated layout.
type LayoutOptions struct {
	StatusPluginPath string
	StatusFilePath   string
	IntervalMS       int
}

func WriteLayout(path string, opts LayoutOptions) error {
	if path == "" {
		return errors.New("layout path is empty")
	}
	plugin := opts.StatusPluginPath
	if plugin == "" {
		return errors.New("status plugin path is empty")
	}
	statusFile := opts.StatusFilePath
	if statusFile == "" {
		return errors.New("status file path is empty")
	}
	interval := opts.IntervalMS
	if interval <= 0 {
		interval = defaultStatusInterval
	}

	absPlugin, err := filepath.Abs(plugin)
	if err != nil {
		return fmt.Errorf("plugin path: %w", err)
	}
	absStatus, err := filepath.Abs(statusFile)
	if err != nil {
		return fmt.Errorf("status path: %w", err)
	}

	content := fmt.Sprintf(`layout {
    pane size=1 borderless=true {
        plugin location="zellij:tab-bar"
    }
    pane
    pane size=1 borderless=true {
        plugin location="file:%s" {
            path "%s"
            interval_ms "%d"
        }
    }
}
`, kdlEscape(absPlugin), kdlEscape(absStatus), interval)

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	return os.WriteFile(path, []byte(content), 0o644)
}

func ResolveStatusPlugin(preferred string) string {
	if preferred != "" {
		return preferred
	}
	if val := os.Getenv("OXL_STATUS_PLUGIN"); val != "" {
		return val
	}
	if path := localPinnedPlugin(); path != "" {
		return path
	}
	return ""
}

func localPinnedPlugin() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	candidate := filepath.Join(cwd, "tools", "zellij", "owlx-status.wasm")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return ""
}

func kdlEscape(value string) string {
	replacer := strings.NewReplacer("\\", "\\\\", "\"", "\\\"")
	return replacer.Replace(value)
}
