package cli

import (
	"os"

	"github.com/spf13/cobra"

	"owlx/internal/app"
	"owlx/internal/cli/handlers"
)

func Execute() error {
	root := newRootCmd()
	return root.Execute()
}

func newRootCmd() *cobra.Command {
	root := &cobra.Command{
		Use:   "owlx",
		Short: "resume your intent — faster.",
		Long:  "resume your intent — faster.",
	}
	root.SetOut(os.Stdout)
	root.SetErr(os.Stderr)
	root.SilenceUsage = true

	root.PersistentPreRunE = func(cmd *cobra.Command, args []string) error {
		state, err := app.Initialize()
		if err != nil {
			return err
		}
		cmd.SetContext(app.WithContext(cmd.Context(), state))
		return nil
	}

	root.AddCommand(
		handlers.NewCmd(),
		handlers.LsCmd(),
		handlers.SearchCmd(),
		handlers.ResumeCmd(),
		handlers.ViewCmd(),
		handlers.DelCmd(),
		handlers.NotifyCmd(),
		handlers.ConfigCmd(),
		handlers.StatusCmd(),
		handlers.CompletionCmd(root),
		handlers.VersionCmd(),
	)
	return root
}
