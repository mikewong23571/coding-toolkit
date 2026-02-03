package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

func newCmdList() *cobra.Command {
	var jsonOut bool
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "list sessions",
		RunE: func(cmd *cobra.Command, _ []string) error {
			sessions, err := app.Store.List()
			if err != nil {
				return err
			}
			if jsonOut {
				payload, err := json.MarshalIndent(sessions, "", "  ")
				if err != nil {
					return err
				}
				fmt.Fprintln(cmd.OutOrStdout(), string(payload))
				return nil
			}
			if len(sessions) == 0 {
				return nil
			}
			fmt.Fprintln(cmd.OutOrStdout(), "ID      LAYOUT     REPO               BRANCH           INTENT")
			fmt.Fprintln(cmd.OutOrStdout(), "--      ------     ----               ------           ------")
			for _, session := range sessions {
				intent := session.Intent
				intent = strings.ReplaceAll(intent, "\n", " ")
				fmt.Fprintf(cmd.OutOrStdout(), "%-8s %-10s %-18s %-16s [%s] %s\n",
					session.ID, session.Layout, session.Repo, session.Branch, session.Category, intent)
			}
			return nil
		},
	}
	cmd.Flags().BoolVar(&jsonOut, "json", false, "output JSON")
	return cmd
}
