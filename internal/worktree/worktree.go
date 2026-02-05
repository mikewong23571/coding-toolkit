package worktree

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

func EnsureRepo(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		return fmt.Errorf("repo not found: %s", dir)
	}
	cmd := exec.Command("git", "-C", dir, "rev-parse", "--is-inside-work-tree")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("not a git repo: %s", dir)
	}
	return nil
}

func EnsureWorktreesIgnored(dir, wtDirname string) error {
	excl := filepath.Join(dir, ".git", "info", "exclude")
	if err := os.MkdirAll(filepath.Dir(excl), 0o755); err != nil {
		return err
	}
	if _, err := os.Stat(excl); err != nil {
		if err := os.WriteFile(excl, []byte(""), 0o644); err != nil {
			return err
		}
	}
	data, err := os.ReadFile(excl)
	if err != nil {
		return err
	}
	lines := strings.Split(string(data), "\n")
	entry := wtDirname + "/"
	for _, line := range lines {
		if strings.TrimSpace(line) == entry {
			return nil
		}
	}
	f, err := os.OpenFile(excl, os.O_APPEND|os.O_WRONLY, 0o644)
	if err != nil {
		return err
	}
	defer f.Close()
	if _, err := fmt.Fprintln(f, entry); err != nil {
		return err
	}
	return nil
}

func AddWorktree(repoDir, branch, wtDir, base string) error {
	args := []string{"-C", repoDir, "worktree", "add", "-b", branch, wtDir}
	if strings.TrimSpace(base) != "" {
		args = append(args, base)
	}
	cmd := exec.Command("git", args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func RemoveWorktree(repoDir, wtDir string) error {
	cmd := exec.Command("git", "-C", repoDir, "worktree", "remove", wtDir, "-f")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
