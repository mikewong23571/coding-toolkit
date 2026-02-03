package cli

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"owlx/internal/git"
	"owlx/internal/state"
	"owlx/internal/zellij"
)

var validLayouts = map[string]bool{
	"main":       true,
	"fork":       true,
	"playground": true,
	"archive":    true,
}

var validCategories = map[string]bool{
	"fix":      true,
	"feat":     true,
	"refactor": true,
	"research": true,
	"chore":    true,
}

func newCmdNew() *cobra.Command {
	var noAttach bool
	var layoutRepo string

	cmd := &cobra.Command{
		Use:   "new [--no-attach] [-C <layout/repo>] <category> <worktree> <intent...>",
		Short: "create a worktree + session",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			category := args[0]
			worktree := args[1]
			intent := strings.TrimSpace(strings.Join(args[2:], " "))
			if !validCategories[category] {
				return fmt.Errorf("invalid category: %s", category)
			}
			if intent == "" {
				return errors.New("intent must be non-empty")
			}

			if err := requireGit(); err != nil {
				return err
			}
			if !noAttach {
				if err := requireZellij(); err != nil {
					return err
				}
			}

			ctx := context.Background()
			cfg := app.Config
			var layout, repo, repoDir string
			if layoutRepo != "" {
				var err error
				layout, repo, repoDir, err = parseLayoutRepo(cfg.Root, layoutRepo)
				if err != nil {
					return err
				}
			} else {
				var err error
				layout, repo, repoDir, err = detectRepoContext(ctx, cfg.Root)
				if err != nil {
					return err
				}
			}

			if !git.IsRepo(ctx, repoDir) {
				return fmt.Errorf("not a git repo: %s", repoDir)
			}

			worktreeDir := filepath.Join(repoDir, cfg.WorktreeDirname, worktree)
			if _, err := os.Stat(worktreeDir); err == nil {
				return fmt.Errorf("worktree path already exists: %s", worktreeDir)
			}
			if err := os.MkdirAll(filepath.Dir(worktreeDir), 0o755); err != nil {
				return fmt.Errorf("create worktree dir: %w", err)
			}

			sessionName := fmt.Sprintf("%s/%s/%s", layout, repo, worktree)
			sessionID := state.GenID(sessionName)
			zellijName := zellijSessionName(state.Session{
				Name: sessionName,
				ID:   sessionID,
			})
			if _, err := app.Store.FindByToken(sessionName); err == nil {
				return fmt.Errorf("session already exists: %s", sessionName)
			}
			if !noAttach {
				if exists, err := app.Zellij.HasSession(ctx, zellijName); err == nil && exists {
					return fmt.Errorf("zellij session already exists: %s", zellijName)
				}
			}
			if err := git.EnsureWorktreesIgnored(repoDir, cfg.WorktreeDirname); err != nil {
				return err
			}
			branch := worktree
			if err := git.WorktreeAdd(ctx, repoDir, branch, worktreeDir); err != nil {
				return err
			}

			session := state.Session{
				ID:          sessionID,
				Name:        sessionName,
				Layout:      layout,
				Repo:        repo,
				Branch:      branch,
				Category:    category,
				Intent:      intent,
				Worktree:    worktree,
				RepoDir:     repoDir,
				WorktreeDir: worktreeDir,
			}
			session, err := app.Store.Save(session)
			if err != nil {
				return err
			}

			if _, err := writeStatusFile(session); err != nil {
				return err
			}

			if noAttach {
				fmt.Fprintf(cmd.OutOrStdout(), "created: %s (%s)\n", session.Name, session.ID)
				return nil
			}

			layoutPath, err := ensureLayout(session)
			if err != nil {
				return err
			}

			opts := zellij.StartOptions{
				SessionName: zellijName,
				LayoutPath:  layoutPath,
				Workdir:     worktreeDir,
				Attach:      true,
			}
			if err := app.Zellij.StartSession(ctx, opts); err != nil {
				return err
			}
			return nil
		},
	}

	cmd.Flags().BoolVar(&noAttach, "no-attach", false, "create without attaching")
	cmd.Flags().StringVarP(&layoutRepo, "chdir", "C", "", "use <layout/repo> instead of cwd")

	return cmd
}

func parseLayoutRepo(root, input string) (string, string, string, error) {
	parts := strings.Split(input, "/")
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("-C expects <layout>/<repo>")
	}
	layout := parts[0]
	repo := parts[1]
	if layout == "" || repo == "" {
		return "", "", "", fmt.Errorf("-C expects <layout>/<repo>")
	}
	if !validLayouts[layout] {
		return "", "", "", fmt.Errorf("invalid layout: %s", layout)
	}
	repoDir := filepath.Join(root, layout, repo)
	if _, err := os.Stat(repoDir); err != nil {
		return "", "", "", fmt.Errorf("repo not found: %s", repoDir)
	}
	return layout, repo, repoDir, nil
}

func detectRepoContext(ctx context.Context, root string) (string, string, string, error) {
	cwd, err := os.Getwd()
	if err != nil {
		return "", "", "", err
	}
	repoDir, err := git.TopLevel(ctx, cwd)
	if err != nil {
		return "", "", "", fmt.Errorf("not inside a git repo")
	}
	repoDir = filepath.Clean(repoDir)
	root = filepath.Clean(root)
	rel, err := filepath.Rel(root, repoDir)
	if err != nil {
		return "", "", "", fmt.Errorf("repo must live under %s/<layout>/<repo>", root)
	}
	if strings.HasPrefix(rel, "..") {
		return "", "", "", fmt.Errorf("repo must live under %s/<layout>/<repo>", root)
	}
	parts := strings.Split(rel, string(os.PathSeparator))
	if len(parts) != 2 {
		return "", "", "", fmt.Errorf("repo must live under %s/<layout>/<repo>", root)
	}
	layout := parts[0]
	repo := parts[1]
	if !validLayouts[layout] {
		return "", "", "", fmt.Errorf("invalid layout: %s", layout)
	}
	return layout, repo, repoDir, nil
}
