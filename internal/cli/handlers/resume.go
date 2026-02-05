package handlers

import (
	"os"

	"github.com/spf13/cobra"
	"golang.org/x/term"

	"owlx/internal/config"
	"owlx/internal/tmux"
)

func ResumeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "resume <session|id>",
		Short: "Attach to a session",
		Args:  cobra.ExactArgs(1),
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := getApp(cmd)
			if err != nil {
				return err
			}
			return cmdResume(app.Cfg, app.Tmux, args[0])
		},
	}
	return cmd
}

func cmdResume(_ config.Config, tm tmux.Manager, token string) error {
	if !term.IsTerminal(int(os.Stdin.Fd())) {
		return Die("TTY required for attach")
	}
	sess, err := resolveSession(tm, token)
	if err != nil {
		return err
	}
	return tm.Attach(sess)
}
