package handlers

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"owlx/internal/config"
	"owlx/internal/domain"
	"owlx/internal/tmux"
	"owlx/internal/worktree"
)

func DelCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del <session|id>",
		Short: "Delete session and worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := getApp(cmd)
			if err != nil {
				return err
			}
			return cmdDel(app.Cfg, app.Tmux, args[0])
		},
	}
	return cmd
}

func cmdDel(_ config.Config, tm tmux.Manager, token string) error {
	if !hasCommand("git") {
		return Die("missing dependency: git")
	}
	sess, err := resolveSession(tm, token)
	if err != nil {
		return err
	}
	if !isOwlxSession(tm, sess) {
		return Die(fmt.Sprintf("refuse: not an owlx session: %s", sess))
	}

	repoDir := tm.ShowOption(sess, domain.RepoDirKey)
	wtDir := tm.ShowOption(sess, domain.WorktreeDirKey)
	if repoDir == "" {
		return Die(fmt.Sprintf("missing repo_dir metadata for %s", sess))
	}
	if wtDir == "" {
		return Die(fmt.Sprintf("missing worktree_dir metadata for %s", sess))
	}

	tm.KillSession(sess)
	if _, err := os.Stat(wtDir); err == nil {
		if err := worktree.RemoveWorktree(repoDir, wtDir); err != nil {
			return Die(fmt.Sprintf("failed to remove worktree: %s", wtDir))
		}
	}
	fmt.Printf("deleted: %s\n", sess)
	return nil
}
