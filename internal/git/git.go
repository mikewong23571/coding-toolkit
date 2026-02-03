package git

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"owlx/internal/bridge"
)

func TopLevel(ctx context.Context, cwd string) (string, error) {
	out, err := bridge.Run(ctx, "git", "-C", cwd, "rev-parse", "--show-toplevel")
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(out), nil
}

func IsRepo(ctx context.Context, dir string) bool {
	_, err := bridge.Run(ctx, "git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	return err == nil
}

func WorktreeAdd(ctx context.Context, repoDir, branch, worktreeDir string) error {
	_, err := bridge.Run(ctx, "git", "-C", repoDir, "worktree", "add", "-b", branch, worktreeDir)
	return err
}

func WorktreeRemove(ctx context.Context, repoDir, worktreeDir string) error {
	_, err := bridge.Run(ctx, "git", "-C", repoDir, "worktree", "remove", worktreeDir, "-f")
	return err
}

func EnsureWorktreesIgnored(repoDir, worktreeDirname string) error {
	infoDir := filepath.Join(repoDir, ".git", "info")
	if err := os.MkdirAll(infoDir, 0o755); err != nil {
		return err
	}
	excludePath := filepath.Join(infoDir, "exclude")
	entry := strings.TrimSuffix(worktreeDirname, "/") + "/"
	data, err := os.ReadFile(excludePath)
	if err != nil && !os.IsNotExist(err) {
		return err
	}
	if strings.Contains(string(data), entry) {
		return nil
	}
	f, err := os.OpenFile(excludePath, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if len(data) > 0 && !strings.HasSuffix(string(data), "\n") {
		if _, err := f.WriteString("\n"); err != nil {
			return err
		}
	}
	_, err = f.WriteString(entry + "\n")
	return err
}
