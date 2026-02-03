package cli

import (
	"fmt"

	"github.com/spf13/cobra"
)

func newCmdCompletion() *cobra.Command {
	cmd := &cobra.Command{
		Use:       "completion bash",
		Short:     "generate bash completion",
		Args:      cobra.ExactArgs(1),
		ValidArgs: []string{"bash"},
		RunE: func(cmd *cobra.Command, args []string) error {
			switch args[0] {
			case "bash":
				return rootCmd.GenBashCompletionV2(cmd.OutOrStdout(), true)
			default:
				return fmt.Errorf("unsupported shell: %s", args[0])
			}
		},
	}
	return cmd
}
