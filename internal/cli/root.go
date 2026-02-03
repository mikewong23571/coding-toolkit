package cli

import (
	"github.com/spf13/cobra"
)

var (
	rootCmd = &cobra.Command{
		Use:           "owlx",
		Short:         "resume your intent — faster.",
		SilenceErrors: true,
		SilenceUsage:  true,
	}
	configPath string
)

func Execute() error {
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config file path")
	rootCmd.PersistentPreRunE = loadApp
	rootCmd.CompletionOptions.DisableDefaultCmd = true

	rootCmd.AddCommand(newCmdConfig())
	rootCmd.AddCommand(newCmdCompletion())
	rootCmd.AddCommand(newCmdNew())
	rootCmd.AddCommand(newCmdResume())
	rootCmd.AddCommand(newCmdList())
	rootCmd.AddCommand(newCmdView())
	rootCmd.AddCommand(newCmdDelete())
	rootCmd.AddCommand(newCmdStatus())
	rootCmd.AddCommand(newCmdSearch())
	rootCmd.AddCommand(newCmdNotify())
}
