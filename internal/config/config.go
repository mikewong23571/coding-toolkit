package config

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"owlx/internal/build"
)

type Config struct {
	Root                string
	WorktreeDirname     string
	NotifyHost          string
	NotifyTopic         string
	NotifyTemplate      string
	NotifyTemplateFile  string
	LeftPaneBottomPct   int
	StatusOnNew         bool
	StatusLines         int
	StatusRightLen      int
	StatusLeftFmt       string
	StatusRightFmt      string
	StatusLineFmt       string
	TmuxConf            string
	TmuxSock            string
	HomeDir             string
	RcPath              string
	OwlxDir             string
	DefaultEmbeddedTmux string
}

const (
	defaultRoot            = "~/projs"
	defaultWorktreeDirname = ".worktrees"
	defaultNotifyHost      = "ntfy.local"
	defaultNotifyTopic     = "owlx-alert"
	defaultLeftBottomPct   = 25
	defaultStatusOnNew     = true
	defaultStatusLines     = 2
	defaultStatusRightLen  = 120
)

var (
	defaultStatusLeftFmt = "#{?@owlx,#{p-10:#{=10:#{@owlx_layout}}} | #{p-8:#{=8:#{@owlx_cat}}} | #{p-20:#{=20:#{@owlx_repo}}} | #{p-20:#{=20:#{@owlx_branch}}},}"
)

// Load reads env + ~/.owlxrc with precedence: env > config > defaults.
func Load() (Config, error) {
	home, err := os.UserHomeDir()
	if err != nil {
		return Config{}, err
	}
	cfg := Config{
		HomeDir: home,
		RcPath:  filepath.Join(home, ".owlxrc"),
		OwlxDir: filepath.Join(home, ".owlx"),
	}
	buildFlag := build.BuildFlag()
	if buildFlag == "" {
		buildFlag = "dev"
	}
	cfg.DefaultEmbeddedTmux = filepath.Join(cfg.HomeDir, ".local", "share", "owlx", buildFlag, "libexec", "tmux")

	env, envSet := readEnv()

	rcVals := map[string]string{}
	if _, err := os.Stat(cfg.RcPath); err == nil {
		rcVals, err = readRcFile(cfg.RcPath)
		if err != nil {
			return Config{}, err
		}
	}

	vals := map[string]string{}
	for k, v := range rcVals {
		vals[k] = v
	}
	for k, v := range env {
		if envSet[k] {
			vals[k] = v
		}
	}

	cfg.Root = getString(vals, "OXL_ROOT", defaultRoot)
	cfg.WorktreeDirname = getString(vals, "OXL_WT_DIRNAME", defaultWorktreeDirname)
	cfg.NotifyHost = defaultNotifyHost
	cfg.NotifyTopic = defaultNotifyTopic
	cfg.NotifyTemplate = ""
	cfg.NotifyTemplateFile = ""
	cfg.StatusLeftFmt = defaultStatusLeftFmt
	cfg.StatusRightFmt = ""
	cfg.StatusLineFmt = ""
	cfg.LeftPaneBottomPct = defaultLeftBottomPct
	cfg.StatusOnNew = defaultStatusOnNew
	cfg.StatusLines = defaultStatusLines
	cfg.StatusRightLen = defaultStatusRightLen
	cfg.TmuxConf = filepath.Join(cfg.OwlxDir, "tmux", "default", "tmux.conf")
	cfg.TmuxSock = filepath.Join(cfg.OwlxDir, "tmux", "default", "tmux.sock")

	cfg.Root = normalizePath(cfg.Root, cfg.HomeDir)
	cfg.NotifyTemplateFile = normalizePath(cfg.NotifyTemplateFile, cfg.HomeDir)
	cfg.TmuxConf = normalizePath(cfg.TmuxConf, cfg.HomeDir)
	cfg.TmuxSock = normalizePath(cfg.TmuxSock, cfg.HomeDir)

	if cfg.StatusLeftFmt == "" {
		cfg.StatusLeftFmt = defaultStatusLeftFmt
	}
	if cfg.StatusRightFmt == "" {
		cfg.StatusRightFmt = fmt.Sprintf("#{?@owlx,#{=%d:#{@owlx_intent}},}", cfg.StatusRightLen)
	}

	return cfg, nil
}

func readEnv() (map[string]string, map[string]bool) {
	keys := []string{
		"OXL_ROOT",
		"OXL_WT_DIRNAME",
	}
	vals := map[string]string{}
	set := map[string]bool{}
	for _, key := range keys {
		if val, ok := os.LookupEnv(key); ok {
			vals[key] = val
			set[key] = true
		}
	}
	return vals, set
}

func readRcFile(path string) (map[string]string, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	vals := map[string]string{}
	s := bufio.NewScanner(f)
	for s.Scan() {
		line := strings.TrimSpace(s.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		if idx := strings.Index(line, "#"); idx != -1 {
			line = strings.TrimSpace(line[:idx])
		}
		if line == "" {
			continue
		}
		key, val, ok := strings.Cut(line, "=")
		if !ok {
			continue
		}
		key = strings.TrimSpace(key)
		val = strings.TrimSpace(val)
		val = strings.Trim(val, "\r\n")
		val = strings.TrimSpace(val)
		if len(val) >= 2 {
			if (val[0] == '\'' && val[len(val)-1] == '\'') || (val[0] == '"' && val[len(val)-1] == '"') {
				val = val[1 : len(val)-1]
			}
		}
		val = os.ExpandEnv(val)
		vals[key] = val
	}
	if err := s.Err(); err != nil {
		return nil, err
	}
	return vals, nil
}

func getString(vals map[string]string, key, def string) string {
	if v, ok := vals[key]; ok {
		return v
	}
	return def
}

func getInt(vals map[string]string, key string, def int) (int, error) {
	if v, ok := vals[key]; ok {
		v = strings.TrimSpace(v)
		if v == "" {
			return def, nil
		}
		n, err := strconv.Atoi(v)
		if err != nil {
			return 0, fmt.Errorf("invalid %s: %s", key, v)
		}
		return n, nil
	}
	return def, nil
}

func getBool(vals map[string]string, key string, def bool) (bool, error) {
	if v, ok := vals[key]; ok {
		v = strings.TrimSpace(v)
		if v == "" {
			return def, nil
		}
		switch v {
		case "1":
			return true, nil
		case "0":
			return false, nil
		default:
			return false, fmt.Errorf("invalid %s: %s", key, v)
		}
	}
	return def, nil
}

func normalizePath(path, home string) string {
	if path == "" {
		return ""
	}
	path = os.ExpandEnv(path)
	if path == "~" {
		return home
	}
	if strings.HasPrefix(path, "~/") {
		return filepath.Join(home, path[2:])
	}
	return path
}

// Validate ensures required settings are present.
func (c Config) Validate() error {
	if c.Root == "" {
		return errors.New("missing OXL_ROOT")
	}
	return nil
}
