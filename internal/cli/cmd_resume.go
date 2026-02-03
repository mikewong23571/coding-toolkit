package cli

import (
	"context"

	"github.com/spf13/cobra"

	"owlx/internal/zellij"
)

func newCmdResume() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <session|id>",
		Short: "attach to a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			if err := requireZellij(); err != nil {
				return err
			}
			session, err := app.Store.FindByToken(args[0])
			if err != nil {
				return err
			}
			ctx := context.Background()
			if _, err := writeStatusFile(session); err != nil {
				return err
			}
			zellijName := zellijSessionName(session)
			exists, err := app.Zellij.HasSession(ctx, zellijName)
			if err != nil {
				return err
			}
			if !exists {
				layoutPath, err := ensureLayout(session)
				if err != nil {
					return err
				}
				opts := zellij.StartOptions{
					SessionName: zellijName,
					LayoutPath:  layoutPath,
					Workdir:     session.WorktreeDir,
					Attach:      true,
				}
				return app.Zellij.StartSession(ctx, opts)
			}
			return app.Zellij.Attach(ctx, zellijName, false)
		},
	}
	return cmd
}
