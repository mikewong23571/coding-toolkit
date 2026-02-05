package tmux

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"owlx/assets"
	"owlx/internal/config"
)

type Manager interface {
	Run(args ...string) error
	RunInteractive(args ...string) error
	Output(args ...string) (string, error)
	OutputQuiet(args ...string) string
	HasSession(name string) bool
	ListSessions() ([]string, error)
	ListPanes(target string) ([]string, error)
	ShowOption(sess, key string) string
	SetOption(sess, key, val string) error
	UnsetOption(sess, key string)
	DisplayMessage(target, format string) string
	NewSession(sess, cwd string) error
	Attach(sess string) error
	KillSession(sess string)
	KillPane(pane string)
	SplitWindow(args ...string) (string, error)
	SelectPane(pane string)
}

type ExecManager struct {
	Bin  string
	Conf string
	Sock string
}

func New(cfg config.Config) (*ExecManager, error) {
	bin, err := resolveTmuxBin(cfg)
	if err != nil {
		return nil, err
	}
	conf := cfg.TmuxConf
	sock := cfg.TmuxSock
	if err := ensureConfFile(conf); err != nil {
		return nil, err
	}
	if err := ensureSockDir(sock); err != nil {
		return nil, err
	}
	return &ExecManager{Bin: bin, Conf: conf, Sock: sock}, nil
}

// EnsureInitialized performs the preflight check and writes a marker file when done.
func EnsureInitialized(cfg config.Config, tm *ExecManager) error {
	if err := os.MkdirAll(cfg.OwlxDir, 0o755); err != nil {
		return err
	}
	if err := verifyFile(tm.Bin); err != nil {
		return err
	}
	if err := verifyFile(cfg.TmuxConf); err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(cfg.TmuxSock), 0o755); err != nil {
		return err
	}

	marker := filepath.Join(cfg.OwlxDir, ".initialized")
	if _, err := os.Stat(marker); err == nil {
		return nil
	}
	content := fmt.Sprintf("tmux_bin=%s\ntmux_conf=%s\ntmux_sock=%s\n", tm.Bin, cfg.TmuxConf, cfg.TmuxSock)
	return os.WriteFile(marker, []byte(content), 0o644)
}

func resolveTmuxBin(cfg config.Config) (string, error) {
	data, ok := assets.EmbeddedTmux()
	if !ok {
		return "", errors.New("embedded tmux not available on this platform")
	}
	if !isELF(data) {
		return "", errors.New("embedded tmux placeholder detected; run make tmux-update to install")
	}
	path := cfg.DefaultEmbeddedTmux
	if err := writeEmbedded(path, data); err != nil {
		return "", err
	}
	return path, nil
}

func verifyFile(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return fmt.Errorf("expected file, got directory: %s", path)
	}
	return nil
}

func writeEmbedded(path string, data []byte) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if info, err := os.Stat(path); err == nil {
		if info.Size() == int64(len(data)) {
			if info.Mode()&0o111 != 0 {
				return nil
			}
		}
	}
	if err := os.WriteFile(path, data, 0o755); err != nil {
		return err
	}
	return nil
}

func ensureConfFile(path string) error {
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(path); err == nil {
		return nil
	}
	if data, ok := assets.EmbeddedTmuxConfig(); ok {
		return os.WriteFile(path, data, 0o644)
	}
	return os.WriteFile(path, []byte("# owlx tmux config\n"), 0o644)
}

func ensureSockDir(path string) error {
	return os.MkdirAll(filepath.Dir(path), 0o755)
}

func isELF(data []byte) bool {
	return len(data) > 4 && data[0] == 0x7f && data[1] == 'E' && data[2] == 'L' && data[3] == 'F'
}

func (m *ExecManager) command(args ...string) *exec.Cmd {
	full := []string{"-S", m.Sock, "-f", m.Conf}
	full = append(full, args...)
	cmd := exec.Command(m.Bin, full...)
	cmd.Env = os.Environ()
	return cmd
}

func (m *ExecManager) Run(args ...string) error {
	cmd := m.command(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func (m *ExecManager) RunInteractive(args ...string) error {
	cmd := m.command(args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin
	return cmd.Run()
}

func (m *ExecManager) Output(args ...string) (string, error) {
	cmd := m.command(args...)
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	err := cmd.Run()
	out := strings.TrimRight(stdout.String(), "\n")
	if err != nil {
		errText := strings.TrimSpace(stderr.String())
		if errText != "" {
			return out, errors.New(errText)
		}
		return out, err
	}
	return out, nil
}

func (m *ExecManager) OutputQuiet(args ...string) string {
	out, err := m.Output(args...)
	if err != nil {
		return ""
	}
	return out
}

func (m *ExecManager) HasSession(name string) bool {
	cmd := m.command("has-session", "-t", name)
	cmd.Stdout = nil
	cmd.Stderr = nil
	return cmd.Run() == nil
}

func (m *ExecManager) ListSessions() ([]string, error) {
	out, err := m.Output("ls", "-F", "#S")
	if err != nil {
		return []string{}, nil
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var sessions []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		sessions = append(sessions, line)
	}
	return sessions, nil
}

func (m *ExecManager) ListPanes(target string) ([]string, error) {
	out, err := m.Output("list-panes", "-t", target, "-F", "#{pane_id}")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var panes []string
	for _, line := range lines {
		if strings.TrimSpace(line) == "" {
			continue
		}
		panes = append(panes, line)
	}
	return panes, nil
}

func (m *ExecManager) ShowOption(sess, key string) string {
	return m.OutputQuiet("show-option", "-t", sess, "-qv", key)
}

func (m *ExecManager) SetOption(sess, key, val string) error {
	return m.Run("set-option", "-t", sess, key, val)
}

func (m *ExecManager) UnsetOption(sess, key string) {
	_ = m.Run("set-option", "-t", sess, "-u", key)
}

func (m *ExecManager) DisplayMessage(target, format string) string {
	if strings.TrimSpace(target) == "" {
		return m.OutputQuiet("display-message", "-p", format)
	}
	return m.OutputQuiet("display-message", "-p", "-t", target, format)
}

func (m *ExecManager) NewSession(sess, cwd string) error {
	return m.Run("new-session", "-d", "-s", sess, "-c", cwd)
}

func (m *ExecManager) Attach(sess string) error {
	return m.RunInteractive("attach", "-t", sess)
}

func (m *ExecManager) KillSession(sess string) {
	_ = m.Run("kill-session", "-t", sess)
}

func (m *ExecManager) KillPane(pane string) {
	_ = m.Run("kill-pane", "-t", pane)
}

func (m *ExecManager) SplitWindow(args ...string) (string, error) {
	out, err := m.Output(append([]string{"split-window"}, args...)...)
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func (m *ExecManager) SelectPane(pane string) {
	_ = m.Run("select-pane", "-t", pane)
}
