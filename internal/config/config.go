package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/pelletier/go-toml/v2"
)

const (
	defaultRoot           = "~/projs"
	defaultWorktreeDir    = ".worktrees"
	defaultNotifyHost     = "ntfy.local"
	defaultNotifyTopic    = "owlx-alert"
	defaultStatusRightLen = 120
	defaultStatusInterval = 1000
)

// Config holds effective runtime configuration.
type Config struct {
	ConfigFile      string       `toml:"-"`
	Root            string       `toml:"root"`
	WorktreeDirname string       `toml:"worktree_dirname"`
	StateDir        string       `toml:"state_dir"`
	ZellijBin       string       `toml:"zellij_bin"`
	ZellijLayout    string       `toml:"zellij_layout"`
	Notify          NotifyConfig `toml:"notify"`
	Status          StatusConfig `toml:"status"`
}

type NotifyConfig struct {
	Host         string `toml:"host"`
	Topic        string `toml:"topic"`
	Template     string `toml:"template"`
	TemplateFile string `toml:"template_file"`
}

type StatusConfig struct {
	OnNew      bool   `toml:"on_new"`
	RightLen   int    `toml:"right_len"`
	IntervalMS int    `toml:"interval_ms"`
	PluginPath string `toml:"plugin_path"`
}

func DefaultConfig() Config {
	return Config{
		Root:            defaultRoot,
		WorktreeDirname: defaultWorktreeDir,
		StateDir:        defaultStateDir(),
		Notify: NotifyConfig{
			Host:  defaultNotifyHost,
			Topic: defaultNotifyTopic,
		},
		Status: StatusConfig{
			OnNew:      false,
			RightLen:   defaultStatusRightLen,
			IntervalMS: defaultStatusInterval,
		},
	}
}

func Load(pathOverride string) (Config, error) {
	cfg := DefaultConfig()
	cfgPath := ConfigPath()
	if pathOverride != "" {
		cfgPath = pathOverride
	}
	cfg.ConfigFile = cfgPath

	if cfgPath != "" {
		if data, err := os.ReadFile(cfgPath); err == nil {
			if err := toml.Unmarshal(data, &cfg); err != nil {
				return Config{}, fmt.Errorf("parse config: %w", err)
			}
		} else if !errors.Is(err, os.ErrNotExist) {
			return Config{}, fmt.Errorf("read config: %w", err)
		}
	}

	applyEnvOverrides(&cfg)
	cfg.Root = expandUser(cfg.Root)
	cfg.StateDir = expandUser(cfg.StateDir)
	cfg.ZellijBin = expandUser(cfg.ZellijBin)
	cfg.ZellijLayout = expandUser(cfg.ZellijLayout)
	cfg.Notify.TemplateFile = expandUser(cfg.Notify.TemplateFile)
	cfg.Status.PluginPath = expandUser(cfg.Status.PluginPath)
	cfg.WorktreeDirname = strings.TrimSpace(cfg.WorktreeDirname)

	if cfg.Root == "" {
		cfg.Root = expandUser(defaultRoot)
	}
	if cfg.WorktreeDirname == "" {
		cfg.WorktreeDirname = defaultWorktreeDir
	}
	if cfg.StateDir == "" {
		cfg.StateDir = defaultStateDir()
	}
	if cfg.Status.RightLen <= 0 {
		cfg.Status.RightLen = defaultStatusRightLen
	}
	if cfg.Status.IntervalMS <= 0 {
		cfg.Status.IntervalMS = defaultStatusInterval
	}

	return cfg, nil
}

func ConfigPath() string {
	base := xdgConfigHome()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "owlx", "config.toml")
}

func WriteDefault(path string) error {
	if path == "" {
		return errors.New("config path is empty")
	}
	cfg := DefaultConfig()
	cfg.StateDir = ""
	cfg.ZellijBin = ""
	cfg.ZellijLayout = ""

	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("encode config: %w", err)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return fmt.Errorf("create config dir: %w", err)
	}
	return os.WriteFile(path, data, 0o644)
}

func defaultStateDir() string {
	base := xdgStateHome()
	if base == "" {
		return ""
	}
	return filepath.Join(base, "owlx")
}

func applyEnvOverrides(cfg *Config) {
	if val := os.Getenv("OXL_ROOT"); val != "" {
		cfg.Root = val
	}
	if val := os.Getenv("OXL_WT_DIRNAME"); val != "" {
		cfg.WorktreeDirname = val
	}
	if val := os.Getenv("OXL_STATE_DIR"); val != "" {
		cfg.StateDir = val
	}
	if val := os.Getenv("OXL_NOTIFY_HOST"); val != "" {
		cfg.Notify.Host = val
	}
	if val := os.Getenv("OXL_NOTIFY_TOPIC"); val != "" {
		cfg.Notify.Topic = val
	}
	if val := os.Getenv("OXL_NOTIFY_TEMPLATE"); val != "" {
		cfg.Notify.Template = val
	}
	if val := os.Getenv("OXL_NOTIFY_TEMPLATE_FILE"); val != "" {
		cfg.Notify.TemplateFile = val
	}
	if val := os.Getenv("OXL_ZELLIJ_BIN"); val != "" {
		cfg.ZellijBin = val
	}
	if val := os.Getenv("OXL_ZELLIJ_LAYOUT"); val != "" {
		cfg.ZellijLayout = val
	}
	if val := os.Getenv("OXL_STATUS_PLUGIN"); val != "" {
		cfg.Status.PluginPath = val
	}
	if val := os.Getenv("OXL_STATUS_INTERVAL_MS"); val != "" {
		if parsed, err := strconv.Atoi(val); err == nil {
			cfg.Status.IntervalMS = parsed
		}
	}
}
