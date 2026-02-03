package zellij

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Client struct {
	Bin string
}

type StartOptions struct {
	SessionName string
	LayoutPath  string
	Workdir     string
	Attach      bool
}

func ResolveBin(preferred string) string {
	if preferred != "" {
		return preferred
	}
	if val := os.Getenv("ZELLIJ_BIN"); val != "" {
		return val
	}
	if val := os.Getenv("OXL_ZELLIJ_BIN"); val != "" {
		return val
	}
	if path := localPinnedBin(); path != "" {
		return path
	}
	return "zellij"
}

func New(bin string) *Client {
	return &Client{Bin: bin}
}

func (c *Client) Version(ctx context.Context) (string, error) {
	out, err := c.run(ctx, "--version")
	return strings.TrimSpace(out), err
}

func (c *Client) ListSessions(ctx context.Context) ([]string, error) {
	out, err := c.run(ctx, "list-sessions")
	if err != nil {
		if isNoSessionsOutput(out) {
			return []string{}, nil
		}
		return nil, err
	}
	lines := strings.Split(strings.TrimSpace(out), "\n")
	var sessions []string
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		parts := strings.Fields(line)
		if len(parts) == 0 {
			continue
		}
		sessions = append(sessions, parts[0])
	}
	return sessions, nil
}

func (c *Client) HasSession(ctx context.Context, name string) (bool, error) {
	if name == "" {
		return false, errors.New("session name is empty")
	}
	sessions, err := c.ListSessions(ctx)
	if err != nil {
		return false, err
	}
	for _, sess := range sessions {
		if sess == name {
			return true, nil
		}
	}
	return false, nil
}

func (c *Client) StartSession(ctx context.Context, opts StartOptions) error {
	if opts.SessionName == "" {
		return errors.New("session name is empty")
	}
	if !opts.Attach {
		args := []string{"attach", "--create-background", opts.SessionName}
		_, err := c.runWithDir(ctx, opts.Workdir, args...)
		return err
	}

	args := []string{"--session", opts.SessionName}
	if opts.LayoutPath != "" {
		args = append(args, "--new-session-with-layout", opts.LayoutPath)
	}
	_, err := c.runWithDir(ctx, opts.Workdir, args...)
	return err
}

func (c *Client) Attach(ctx context.Context, sessionName string, create bool) error {
	if sessionName == "" {
		return errors.New("session name is empty")
	}
	args := []string{"attach"}
	if create {
		args = append(args, "--create")
	}
	args = append(args, sessionName)
	_, err := c.run(ctx, args...)
	return err
}

func (c *Client) KillSession(ctx context.Context, sessionName string) error {
	if sessionName == "" {
		return errors.New("session name is empty")
	}
	_, err := c.run(ctx, "kill-session", sessionName)
	return err
}

func (c *Client) run(ctx context.Context, args ...string) (string, error) {
	return c.runWithDir(ctx, "", args...)
}

func (c *Client) runWithDir(ctx context.Context, dir string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, c.Bin, args...)
	if dir != "" {
		cmd.Dir = dir
	}
	output, err := cmd.CombinedOutput()
	if err != nil {
		return string(output), fmt.Errorf("zellij %s: %w: %s", strings.Join(args, " "), err, strings.TrimSpace(string(output)))
	}
	return string(output), nil
}

func localPinnedBin() string {
	cwd, err := os.Getwd()
	if err != nil {
		return ""
	}
	candidate := filepath.Join(cwd, "tools", "zellij", "zellij")
	if _, err := os.Stat(candidate); err == nil {
		return candidate
	}
	return ""
}

func isNoSessionsOutput(output string) bool {
	msg := strings.ToLower(output)
	return strings.Contains(msg, "no active zellij sessions found") ||
		strings.Contains(msg, "no sessions found")
}
