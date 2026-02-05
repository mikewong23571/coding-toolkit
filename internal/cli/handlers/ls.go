package handlers

import (
	"fmt"
	"io"
	"strconv"
	"strings"
	"text/tabwriter"
	"time"

	"github.com/spf13/cobra"

	"owlx/internal/config"
	"owlx/internal/domain"
	"owlx/internal/tmux"
	"owlx/internal/util"
)

func LsCmd() *cobra.Command {
	var output string
	var noHeader bool
	cmd := &cobra.Command{
		Use:   "ls",
		Short: "List owlx sessions",
		RunE: func(cmd *cobra.Command, args []string) error {
			app, err := getApp(cmd)
			if err != nil {
				return err
			}
			return cmdLs(app.Cfg, app.Tmux, output, noHeader, cmd.OutOrStdout())
		},
	}
	cmd.Flags().StringVarP(&output, "output", "o", "table", "output format: table|wide|json")
	cmd.Flags().BoolVar(&noHeader, "no-header", false, "omit header row")
	return cmd
}

type lsRow struct {
	Session    string
	ID         string
	Layout     string
	Repo       string
	Branch     string
	Category   string
	Intent     string
	BaseBranch string
	CreatedTs  string
}

func cmdLs(_ config.Config, tm tmux.Manager, output string, noHeader bool, out io.Writer) error {
	sessions, _ := tm.ListSessions()
	rows := buildRows(tm, sessions)
	if len(rows) == 0 {
		if output == "json" {
			fmt.Fprintln(out, "[]")
		}
		return nil
	}

	switch output {
	case "table", "wide", "json":
	default:
		return Die(fmt.Sprintf("unknown output format: %s (use: table|wide|json)", output))
	}

	if output == "json" {
		first := true
		fmt.Fprint(out, "[")
		for _, row := range rows {
			if !first {
				fmt.Fprint(out, ",")
			}
			first = false
			fmt.Fprintf(out, "{\"session\":\"%s\",\"id\":\"%s\",\"layout\":\"%s\",\"repo\":\"%s\",\"branch\":\"%s\",\"category\":\"%s\",\"intent\":\"%s\"}",
				util.JSONEscape(row.Session),
				util.JSONEscape(row.ID),
				util.JSONEscape(row.Layout),
				util.JSONEscape(row.Repo),
				util.JSONEscape(row.Branch),
				util.JSONEscape(row.Category),
				util.JSONEscape(row.Intent),
			)
		}
		fmt.Fprintln(out, "]")
		return nil
	}

	w := tabwriter.NewWriter(out, 0, 0, 2, ' ', 0)
	if !noHeader {
		if output == "wide" {
			fmt.Fprintln(w, "Id\tLayout\tRepo\tBranch\tBase\tCreated\tIntent")
		} else {
			fmt.Fprintln(w, "Id\tLayout\tRepo\tBranch\tIntent")
		}
	}
	for _, row := range rows {
		if output == "wide" {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t[%s] %s\n",
				util.TSVEscape(row.ID),
				util.TSVEscape(row.Layout),
				util.TSVEscape(row.Repo),
				util.TSVEscape(row.Branch),
				util.TSVEscape(row.BaseBranch),
				util.TSVEscape(formatCreated(row.CreatedTs)),
				util.TSVEscape(row.Category),
				util.TSVEscape(row.Intent),
			)
			continue
		}
		fmt.Fprintf(w, "%s\t%s\t%s\t%s\t[%s] %s\n",
			util.TSVEscape(row.ID),
			util.TSVEscape(row.Layout),
			util.TSVEscape(row.Repo),
			util.TSVEscape(row.Branch),
			util.TSVEscape(row.Category),
			util.TSVEscape(row.Intent),
		)
	}
	return w.Flush()
}

func buildRows(tm tmux.Manager, sessions []string) []lsRow {
	rows := make([]lsRow, 0, len(sessions))
	for _, sess := range sessions {
		if !isOwlxSession(tm, sess) {
			continue
		}
		rows = append(rows, lsRow{
			Session:    sess,
			ID:         sessionID(tm, sess),
			Layout:     tm.ShowOption(sess, domain.LayoutKey),
			Repo:       tm.ShowOption(sess, domain.RepoKey),
			Branch:     tm.ShowOption(sess, domain.BranchKey),
			Category:   tm.ShowOption(sess, domain.CatKey),
			Intent:     tm.ShowOption(sess, domain.IntentKey),
			BaseBranch: tm.ShowOption(sess, domain.BaseBranchKey),
			CreatedTs:  tm.ShowOption(sess, domain.CreatedTsKey),
		})
	}
	return rows
}

func formatCreated(raw string) string {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return ""
	}
	ts, err := strconv.ParseInt(raw, 10, 64)
	if err != nil || ts <= 0 {
		return ""
	}
	return time.Unix(ts, 0).Local().Format("2006-01-02 15:04")
}
