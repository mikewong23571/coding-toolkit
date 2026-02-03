package cli

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"owlx/internal/git"
)

func newCmdDelete() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "del <session|id>",
		Short: "delete session and worktree",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			session, err := app.Store.FindByToken(args[0])
			if err != nil {
				return err
			}
			ctx := context.Background()
			zellijName := zellijSessionName(session)
			if err := app.Zellij.KillSession(ctx, zellijName); err != nil {
				fmt.Fprintf(cmd.ErrOrStderr(), "warning: %v\n", err)
			}
			if session.RepoDir != "" && session.WorktreeDir != "" {
				if err := git.WorktreeRemove(ctx, session.RepoDir, session.WorktreeDir); err != nil {
					return err
				}
			}
			if err := app.Store.Delete(session.ID); err != nil {
				return err
			}
			fmt.Fprintf(cmd.OutOrStdout(), "deleted: %s\n", session.Name)
			return nil
		},
	}
	return cmd
}
