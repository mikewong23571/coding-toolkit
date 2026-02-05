package handlers

import (
	"fmt"

	"github.com/spf13/cobra"

	"owlx/internal/config"
	"owlx/internal/domain"
	"owlx/internal/tmux"
)

func ViewCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <session|id>",
		Short: "View session metadata",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := getApp(cmd)
			if err != nil {
				return err
			}
			return cmdView(app.Cfg, app.Tmux, args[0])
		},
	}
	return cmd
}

func cmdView(_ config.Config, tm tmux.Manager, token string) error {
	sess, err := resolveSession(tm, token)
	if err != nil {
		return err
	}

	fmt.Printf("session:   %s\n", sess)
	fmt.Printf("id:        %s\n", sessionID(tm, sess))
	fmt.Printf("owlx:      %s\n", tm.ShowOption(sess, domain.FlagKey))
	fmt.Printf("layout:    %s\n", tm.ShowOption(sess, domain.LayoutKey))
	fmt.Printf("repo:      %s\n", tm.ShowOption(sess, domain.RepoKey))
	fmt.Printf("worktree:  %s\n", tm.ShowOption(sess, domain.WorktreeKey))
	fmt.Printf("branch:    %s\n", tm.ShowOption(sess, domain.BranchKey))
	fmt.Printf("cat:       %s\n", tm.ShowOption(sess, domain.CatKey))
	fmt.Printf("intent:    %s\n", tm.ShowOption(sess, domain.IntentKey))
	fmt.Printf("repo_dir:  %s\n", tm.ShowOption(sess, domain.RepoDirKey))
	fmt.Printf("wt_dir:    %s\n", tm.ShowOption(sess, domain.WorktreeDirKey))
	return nil
}
