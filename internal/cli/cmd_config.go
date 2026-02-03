package cli

import (
	"fmt"
	"os"

	"github.com/spf13/cobra"

	"owlx/internal/config"
)

func newCmdConfig() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config",
		Short: "show or initialize configuration",
		RunE: func(cmd *cobra.Command, _ []string) error {
			cfg := app.Config
			fmt.Printf("config:    %s\n", cfg.ConfigFile)
			fmt.Printf("root:      %s\n", cfg.Root)
			fmt.Printf("worktrees: %s\n", cfg.WorktreeDirname)
			fmt.Printf("state:     %s\n", cfg.StateDir)
			fmt.Printf("notify:    host=%s topic=%s\n", cfg.Notify.Host, cfg.Notify.Topic)
			fmt.Printf("status:    right_len=%d on_new=%v interval_ms=%d plugin=%s\n",
				cfg.Status.RightLen, cfg.Status.OnNew, cfg.Status.IntervalMS, cfg.Status.PluginPath)
			fmt.Printf("zellij:    bin=%s layout=%s\n", cfg.ZellijBin, cfg.ZellijLayout)
			return nil
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "init",
		Short: "write a default config file",
		RunE: func(cmd *cobra.Command, _ []string) error {
			path := config.ConfigPath()
			if configPath != "" {
				path = configPath
			}
			if path == "" {
				return fmt.Errorf("config path is empty")
			}
			if _, err := os.Stat(path); err == nil {
				return fmt.Errorf("config already exists: %s", path)
			}
			if err := config.WriteDefault(path); err != nil {
				return err
			}
			fmt.Printf("created: %s\n", path)
			return nil
		},
	})

	cmd.AddCommand(newCmdConfigZellij())

	return cmd
}
