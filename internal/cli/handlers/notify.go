package handlers

import (
	"fmt"

	"github.com/spf13/cobra"

	"owlx/internal/notify"
)

func NotifyCmd() *cobra.Command {
	var session string
	var useStdin bool
	var topic string
	var url string
	cmd := &cobra.Command{
		Use:   "notify [--session <session|id>] [--stdin|-] [--topic <name>|--url <url>] <json>",
		Short: "Send notification",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := getApp(cmd)
			if err != nil {
				return err
			}
			jsonPayload := ""
			if len(args) == 1 && args[0] == "-" {
				useStdin = true
			}
			if !useStdin {
				if len(args) != 1 {
					return fmt.Errorf("notify requires JSON payload")
				}
				jsonPayload = args[0]
			}
			opts := notify.Options{
				Session:  session,
				UseStdin: useStdin,
				Topic:    topic,
				URL:      url,
				JSON:     jsonPayload,
			}
			return notify.HandleOptions(app.Cfg, app.Tmux, opts)
		},
	}
	cmd.Flags().StringVar(&session, "session", "", "session or id")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "read JSON from stdin")
	cmd.Flags().StringVar(&topic, "topic", "", "notify topic")
	cmd.Flags().StringVar(&url, "url", "", "notify URL")
	return cmd
}
