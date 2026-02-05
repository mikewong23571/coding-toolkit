package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"owlx/internal/config"
	"owlx/internal/domain"
	"owlx/internal/tmux"
	"owlx/internal/util"
	"owlx/internal/worktree"
)

func NewCmd() *cobra.Command {
	var noAttach bool
	var detached bool
	var layoutRepo string
	var baseBranch string
	cmd := &cobra.Command{
		Use:   "new [flags] <category> <worktree> <intent...>",
		Short: "Create a new tmux session + worktree",
		Args:  cobra.MinimumNArgs(3),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := getApp(cmd)
			if err != nil {
				return err
			}
			cat := args[0]
			wt := args[1]
			intent := strings.Join(args[2:], " ")
			attach := !(noAttach || detached)
			return cmdNew(app.Cfg, app.Tmux, cat, wt, intent, attach, layoutRepo, baseBranch)
		},
	}
	cmd.Flags().BoolVarP(&noAttach, "no-attach", "d", false, "create without attaching")
	cmd.Flags().BoolVar(&detached, "detached", false, "create without attaching")
	cmd.Flags().StringVarP(&layoutRepo, "C", "C", "", "layout/repo")
	cmd.Flags().StringVar(&baseBranch, "base-branch", "", "base branch for new worktree")
	return cmd
}

func cmdNew(cfg config.Config, tm tmux.Manager, cat, wt, intent string, attach bool, layoutRepo, baseBranch string) error {
	if !hasCommand("git") {
		return Die("missing dependency: git")
	}

	if !domain.IsValidCategory(cat) {
		return Die(fmt.Sprintf("invalid category: %s (use: fix|feat|refactor|research|chore)", cat))
	}
	if strings.TrimSpace(intent) == "" {
		return Die("intent must be non-empty")
	}

	if attach && !term.IsTerminal(int(os.Stdin.Fd())) {
		return Die("TTY required for attach (use: --no-attach)")
	}

	var layout, repo, repoDir string
	var err error
	if layoutRepo != "" {
		layout, repo, repoDir, err = ParseLayoutRepo(cfg.Root, layoutRepo)
	} else {
		layout, repo, repoDir, err = DetectRepoContext(cfg.Root)
	}
	if err != nil {
		return err
	}

	if err := worktree.EnsureRepo(repoDir); err != nil {
		return err
	}
	if err := worktree.EnsureWorktreesIgnored(repoDir, cfg.WorktreeDirname); err != nil {
		return err
	}

	wtDir := WorktreePath(repoDir, cfg.WorktreeDirname, wt)
	sess := SessionName(layout, repo, wt)
	branch := wt
	if strings.TrimSpace(baseBranch) == "" {
		baseBranch, err = currentBranch(repoDir)
		if err != nil {
			return err
		}
	}

	if tm.HasSession(sess) {
		return Die(fmt.Sprintf("session already exists: %s (use: owlx resume \"%s\")", sess, sess))
	}
	if _, err := os.Stat(wtDir); err == nil {
		return Die(fmt.Sprintf("worktree path already exists: %s", wtDir))
	}
	if err := os.MkdirAll(filepath.Dir(wtDir), 0o755); err != nil {
		return err
	}

	worktreeCreated := false
	sessionCreated := false
	cleanup := func(orig error) error {
		if sessionCreated {
			tm.KillSession(sess)
		}
		if err := worktree.RemoveWorktree(repoDir, wtDir); err != nil {
			return fmt.Errorf("%w (cleanup failed: %v)", orig, err)
		}
		return orig
	}

	if err := worktree.AddWorktree(repoDir, branch, wtDir, baseBranch); err != nil {
		return err
	}
	worktreeCreated = true

	if err := tm.NewSession(sess, wtDir); err != nil {
		if worktreeCreated {
			return cleanup(err)
		}
		return err
	}
	sessionCreated = true

	if err := tm.SetOption(sess, domain.FlagKey, "1"); err != nil {
		if worktreeCreated {
			return cleanup(err)
		}
		return err
	}
	_ = tm.SetOption(sess, domain.LayoutKey, layout)
	_ = tm.SetOption(sess, domain.RepoKey, repo)
	_ = tm.SetOption(sess, domain.RepoDirKey, repoDir)
	_ = tm.SetOption(sess, domain.WorktreeKey, wt)
	_ = tm.SetOption(sess, domain.WorktreeDirKey, wtDir)
	_ = tm.SetOption(sess, domain.BranchKey, branch)
	_ = tm.SetOption(sess, domain.BaseBranchKey, baseBranch)
	_ = tm.SetOption(sess, domain.CatKey, cat)
	_ = tm.SetOption(sess, domain.IntentKey, intent)
	_ = tm.SetOption(sess, domain.IDKey, util.GenID(sess))
	_ = tm.SetOption(sess, domain.CreatedTsKey, fmt.Sprintf("%d", time.Now().Unix()))

	if err := setupLayout(cfg, tm, sess); err != nil {
		if worktreeCreated {
			return cleanup(err)
		}
		return err
	}

	if attach {
		if err := tm.Attach(sess); err != nil {
			if worktreeCreated {
				return cleanup(err)
			}
			return err
		}
		return nil
	}
	fmt.Printf("created: %s\n", sess)
	return nil
}
