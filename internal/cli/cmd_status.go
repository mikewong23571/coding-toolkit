package cli

import (
	"fmt"

	"github.com/spf13/cobra"

	"owlx/internal/status"
)

func newCmdStatus() *cobra.Command {
	var outputPath string
	cmd := &cobra.Command{
		Use:   "status <session|id>",
		Short: "render a status line",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			session, err := app.Store.FindByToken(args[0])
			if err != nil {
				return err
			}
			line := status.FormatLine(session.Layout, session.Category, session.Repo, session.Branch, session.Intent, app.Config.Status.RightLen)
			if outputPath == "" {
				fmt.Fprintln(cmd.OutOrStdout(), line)
				return nil
			}
			return status.WriteFile(outputPath, line)
		},
	}
	cmd.Flags().StringVar(&outputPath, "output", "", "write status line to file")
	return cmd
}
