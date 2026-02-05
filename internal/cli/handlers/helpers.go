package handlers

import (
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"owlx/internal/app"
	"owlx/internal/domain"
)

func getApp(cmd *cobra.Command) (*app.State, error) {
	return app.FromContext(cmd.Context())
}

func Die(msg string) error {
	return errors.New("owlx: " + msg)
}

func RepoPath(root, layout, repo string) string {
	return filepath.Join(root, layout, repo)
}

func WorktreePath(repoDir, wtDirname, wt string) string {
	return filepath.Join(repoDir, wtDirname, wt)
}

func SessionName(layout, repo, wt string) string {
	return fmt.Sprintf("%s/%s/%s", layout, repo, wt)
}

func DetectRepoContext(root string) (layout, repo, repoDir string, err error) {
	repoDir, err = gitTopLevel()
	if err != nil {
		return "", "", "", Die("not inside a git repo")
	}
	repoDir, err = filepath.Abs(repoDir)
	if err != nil {
		return "", "", "", err
	}
	rootAbs, err := filepath.Abs(root)
	if err != nil {
		return "", "", "", err
	}
	if !strings.HasPrefix(repoDir, rootAbs+string(os.PathSeparator)) {
		return "", "", "", Die(fmt.Sprintf("repo must live under %s/<layout>/<repo>", rootAbs))
	}
	rel := strings.TrimPrefix(repoDir, rootAbs+string(os.PathSeparator))
	parts := strings.Split(rel, string(os.PathSeparator))
	if len(parts) < 2 {
		return "", "", "", Die(fmt.Sprintf("repo must live under %s/<layout>/<repo>", rootAbs))
	}
	layout = parts[0]
	repo = parts[1]
	if layout == "" || repo == "" || len(parts) != 2 {
		return "", "", "", Die(fmt.Sprintf("repo must live under %s/<layout>/<repo>", rootAbs))
	}
	if !domain.IsValidLayout(layout) {
		return "", "", "", Die(fmt.Sprintf("invalid layout: %s", layout))
	}
	return layout, repo, repoDir, nil
}

func ParseLayoutRepo(root, input string) (layout, repo, repoDir string, err error) {
	if !strings.Contains(input, "/") {
		return "", "", "", Die("-C expects <layout>/<repo>")
	}
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return "", "", "", Die("-C expects <layout>/<repo>")
	}
	layout = parts[0]
	repo = parts[1]
	if layout == "" || repo == "" {
		return "", "", "", Die("-C expects <layout>/<repo>")
	}
	if !domain.IsValidLayout(layout) {
		return "", "", "", Die(fmt.Sprintf("invalid layout: %s", layout))
	}
	repoDir = RepoPath(root, layout, repo)
	if _, err := os.Stat(repoDir); err != nil {
		return "", "", "", Die(fmt.Sprintf("repo not found: %s", repoDir))
	}
	repoDir, err = filepath.Abs(repoDir)
	if err != nil {
		return "", "", "", err
	}
	return layout, repo, repoDir, nil
}

func gitTopLevel() (string, error) {
	cmd := execCommand("git", "rev-parse", "--show-toplevel")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func currentBranch(repoDir string) (string, error) {
	cmd := execCommand("git", "-C", repoDir, "rev-parse", "--abbrev-ref", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	branch := strings.TrimSpace(string(out))
	if branch == "" {
		return "", fmt.Errorf("failed to resolve current branch")
	}
	return branch, nil
}

func execCommand(name string, args ...string) *exec.Cmd {
	return exec.Command(name, args...)
}

func hasCommand(name string) bool {
	_, err := exec.LookPath(name)
	return err == nil
}

func mustHomeDir() string {
	home, _ := os.UserHomeDir()
	return home
}
