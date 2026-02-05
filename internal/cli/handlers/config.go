package handlers

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"owlx/internal/config"
)

func ConfigCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "config [init|edit|gitignore]",
		Short: "Manage config",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := getApp(cmd)
			if err != nil {
				return err
			}
			return cmdConfig(app.Cfg, args)
		},
	}
	return cmd
}

func cmdConfig(cfg config.Config, args []string) error {
	sub := ""
	if len(args) > 0 {
		sub = args[0]
	}
	switch sub {
	case "":
		fmt.Printf("config:    %s\n", cfg.RcPath)
		fmt.Printf("OXL_ROOT:  %s\n", cfg.Root)
		fmt.Printf("OXL_WT_DIRNAME: %s\n", cfg.WorktreeDirname)
		fmt.Println("precedence: env > config > defaults")
		return nil
	case "init":
		if _, err := os.Stat(cfg.RcPath); err == nil {
			return Die(fmt.Sprintf("config already exists: %s", cfg.RcPath))
		}
		if err := os.WriteFile(cfg.RcPath, []byte(`# owlx config (shell-style assignments)
OXL_ROOT="$HOME/projs"
OXL_WT_DIRNAME=".worktrees"
`), 0o644); err != nil {
			return err
		}
		fmt.Printf("created: %s\n", cfg.RcPath)
		return nil
	case "edit":
		ed := os.Getenv("EDITOR")
		if ed == "" {
			ed = "vi"
		}
		cmd := exec.Command(ed, cfg.RcPath)
		cmd.Stdin = os.Stdin
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr
		return cmd.Run()
	case "gitignore":
		action := "show"
		if len(args) > 1 {
			action = args[1]
		}
		return configGitignore(action)
	default:
		return Die(fmt.Sprintf("unknown config subcommand: %s (use: owlx config [init|edit|gitignore])", sub))
	}
}

func configGitignore(action string) error {
	conf := gitGlobalExcludes()
	switch action {
	case "", "show":
		fmt.Printf("global excludes: %s\n", conf)
		fmt.Println("line to add:")
		fmt.Println(".worktrees/")
		return nil
	case "install":
		if err := os.MkdirAll(filepath.Dir(conf), 0o755); err != nil {
			return err
		}
		if _, err := os.Stat(conf); err != nil {
			if err := os.WriteFile(conf, []byte(""), 0o644); err != nil {
				return err
			}
		}
		data, _ := os.ReadFile(conf)
		if !strings.Contains(string(data), ".worktrees/") {
			f, err := os.OpenFile(conf, os.O_APPEND|os.O_WRONLY, 0o644)
			if err != nil {
				return err
			}
			defer f.Close()
			fmt.Fprintln(f, ".worktrees/")
		}
		if gitConfig("core.excludesfile") == "" {
			_ = exec.Command("git", "config", "--global", "core.excludesfile", conf).Run()
		}
		fmt.Printf("installed gitignore entry in: %s\n", conf)
		return nil
	default:
		return Die(fmt.Sprintf("unknown config gitignore subcommand: %s (use: owlx config gitignore [show|install])", action))
	}
}

func gitGlobalExcludes() string {
	conf := gitConfig("core.excludesfile")
	if conf == "" {
		conf = filepath.Join(mustHomeDir(), ".gitignore_global")
	}
	return conf
}

func gitConfig(key string) string {
	cmd := exec.Command("git", "config", "--global", key)
	out, err := cmd.Output()
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(out))
}
