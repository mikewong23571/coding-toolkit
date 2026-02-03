package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"owlx/internal/config"
	"owlx/internal/zellij"
)

func newCmdConfigZellij() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "zellij",
		Short: "zellij integration helpers",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return showZellijConfig(cmd)
		},
	}

	cmd.AddCommand(&cobra.Command{
		Use:   "show",
		Short: "show zellij integration details",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return showZellijConfig(cmd)
		},
	})

	cmd.AddCommand(&cobra.Command{
		Use:   "install",
		Short: "install the owlx status plugin to the data dir",
		RunE: func(cmd *cobra.Command, _ []string) error {
			return installZellijPlugin(cmd)
		},
	})

	return cmd
}

func showZellijConfig(cmd *cobra.Command) error {
	cfg := app.Config
	bin := zellij.ResolveBin(cfg.ZellijBin)
	plugin := zellij.ResolveStatusPlugin(cfg.Status.PluginPath)
	dataHome := config.DataHome()
	pluginTarget := ""
	if dataHome != "" {
		pluginTarget = filepath.Join(dataHome, "owlx", "plugins", "owlx-status.wasm")
	}
	fmt.Fprintf(cmd.OutOrStdout(), "zellij_bin:     %s\n", bin)
	fmt.Fprintf(cmd.OutOrStdout(), "status_plugin:  %s\n", plugin)
	fmt.Fprintf(cmd.OutOrStdout(), "plugin_install: %s\n", pluginTarget)
	return nil
}

func installZellijPlugin(cmd *cobra.Command) error {
	plugin := zellij.ResolveStatusPlugin(app.Config.Status.PluginPath)
	if plugin == "" {
		return fmt.Errorf("status plugin not found (set status.plugin_path or OXL_STATUS_PLUGIN)")
	}
	dataHome := config.DataHome()
	if dataHome == "" {
		return fmt.Errorf("XDG data home not found")
	}
	pluginTarget := filepath.Join(dataHome, "owlx", "plugins", "owlx-status.wasm")
	if err := os.MkdirAll(filepath.Dir(pluginTarget), 0o755); err != nil {
		return err
	}
	data, err := os.ReadFile(plugin)
	if err != nil {
		return err
	}
	if err := os.WriteFile(pluginTarget, data, 0o755); err != nil {
		return err
	}
	fmt.Fprintf(cmd.OutOrStdout(), "installed: %s\n", pluginTarget)
	fmt.Fprintf(cmd.OutOrStdout(), "set status.plugin_path to this path if needed.\n")
	return nil
}
