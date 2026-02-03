package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCmdView() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "view <session|id>",
		Short: "show session details",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			session, err := app.Store.FindByToken(args[0])
			if err != nil {
				return err
			}
			out := cmd.OutOrStdout()
			fmt.Fprintf(out, "session:   %s\n", session.Name)
			fmt.Fprintf(out, "id:        %s\n", session.ID)
			fmt.Fprintf(out, "layout:    %s\n", session.Layout)
			fmt.Fprintf(out, "repo:      %s\n", session.Repo)
			fmt.Fprintf(out, "worktree:  %s\n", session.Worktree)
			fmt.Fprintf(out, "branch:    %s\n", session.Branch)
			fmt.Fprintf(out, "category:  %s\n", session.Category)
			fmt.Fprintf(out, "intent:    %s\n", session.Intent)
			fmt.Fprintf(out, "repo_dir:  %s\n", session.RepoDir)
			fmt.Fprintf(out, "wt_dir:    %s\n", session.WorktreeDir)
			return nil
		},
	}
	return cmd
}
