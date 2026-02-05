package handlers

import (
	"os"

	"github.com/spf13/cobra"

	"owlx/internal/config"
	"owlx/internal/tmux"
)

func StatusCmd() *cobra.Command {
	var session string
	cmd := &cobra.Command{
		Use:   "status [on|off|toggle] [--session <session|id>]",
		Short: "Manage status line",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := getApp(cmd)
			if err != nil {
				return err
			}
			action := "toggle"
			if len(args) > 0 {
				switch args[0] {
				case "on", "off", "toggle":
					action = args[0]
				default:
					if session == "" {
						session = args[0]
					}
				}
			}
			return cmdStatus(app.Cfg, app.Tmux, action, session)
		},
	}
	cmd.Flags().StringVar(&session, "session", "", "session or id")
	return cmd
}

func cmdStatus(cfg config.Config, tm tmux.Manager, action, target string) error {
	if target == "" {
		if os.Getenv("TMUX") == "" {
			return Die("session required (use: owlx status --session <session|id>)")
		}
		target = tm.DisplayMessage(os.Getenv("TMUX_PANE"), "#S")
		if target == "" {
			target = tm.DisplayMessage("", "#S")
		}
		if target == "" {
			return Die("session required (use: owlx status --session <session|id>)")
		}
	}

	sess, err := resolveSession(tm, target)
	if err != nil {
		return err
	}
	if !isOwlxSession(tm, sess) {
		return Die("refuse: not an owlx session: " + sess)
	}

	switch action {
	case "on":
		return statusOn(cfg, tm, sess)
	case "off":
		return statusOff(tm, sess)
	case "toggle":
		if statusIsOn(tm, sess) {
			return statusOff(tm, sess)
		}
		return statusOn(cfg, tm, sess)
	}
	return nil
}
