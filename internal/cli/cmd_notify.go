package cli

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"owlx/internal/state"
)

type notifyPayload struct {
	Type string `json:"type"`
	Cwd  string `json:"cwd"`
}

func newCmdNotify() *cobra.Command {
	var sessionToken string
	var useStdin bool
	var topic string
	var url string

	cmd := &cobra.Command{
		Use:   "notify [--session <session|id>] [--stdin|-] [--topic <name>|--url <url>] <json>",
		Short: "send a notification",
		RunE: func(cmd *cobra.Command, args []string) error {
			payload, err := readNotifyPayload(useStdin, args)
			if err != nil {
				return err
			}
			if payload.Type != "agent-turn-complete" {
				return nil
			}

			session, err := resolveNotifySession(sessionToken, payload.Cwd)
			if err != nil {
				return err
			}
			if session == nil {
				return nil
			}

			message, err := renderNotifyMessage(*session)
			if err != nil {
				return err
			}

			if url == "" {
				if topic == "" {
					topic = app.Config.Notify.Topic
				}
				url = buildNotifyURL(app.Config.Notify.Host, topic)
			}
			if url == "" {
				fmt.Fprintln(cmd.OutOrStdout(), message)
				return nil
			}

			ctx := context.Background()
			return sendNotify(ctx, url, message)
		},
	}

	cmd.Flags().StringVar(&sessionToken, "session", "", "session name or id")
	cmd.Flags().BoolVar(&useStdin, "stdin", false, "read JSON payload from stdin")
	cmd.Flags().StringVar(&topic, "topic", "", "override notify topic")
	cmd.Flags().StringVar(&url, "url", "", "override notify url")

	return cmd
}

func readNotifyPayload(useStdin bool, args []string) (notifyPayload, error) {
	var input string
	if !useStdin && len(args) > 0 && args[0] == "-" {
		useStdin = true
	}
	if useStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return notifyPayload{}, err
		}
		input = string(data)
	} else if len(args) > 0 {
		input = args[0]
	}
	input = strings.TrimSpace(input)
	if input == "" {
		return notifyPayload{}, errors.New("notify requires JSON payload")
	}
	var payload notifyPayload
	if err := json.Unmarshal([]byte(input), &payload); err != nil {
		return notifyPayload{}, err
	}
	return payload, nil
}

func resolveNotifySession(token, cwd string) (*state.Session, error) {
	if token != "" {
		session, err := app.Store.FindByToken(token)
		if err != nil {
			return nil, err
		}
		return &session, nil
	}
	if cwd == "" {
		return nil, nil
	}
	cwdPath, err := filepath.Abs(cwd)
	if err != nil {
		return nil, nil
	}
	cwdPath = filepath.Clean(cwdPath)

	sessions, err := app.Store.List()
	if err != nil {
		return nil, err
	}
	for _, session := range sessions {
		if isWithin(cwdPath, session.WorktreeDir) {
			return &session, nil
		}
	}
	for _, session := range sessions {
		if isWithin(cwdPath, session.RepoDir) {
			return &session, nil
		}
	}
	return nil, nil
}

func renderNotifyMessage(session state.Session) (string, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "**session_id**: `%s`\n", session.ID)
	fmt.Fprintf(&b, "**repo**: `%s`\n", session.Repo)
	fmt.Fprintf(&b, "**branch**: `%s`\n", session.Branch)
	fmt.Fprintf(&b, "**category**: `%s`\n", session.Category)
	fmt.Fprintf(&b, "**intent**: %s\n", session.Intent)

	template, err := loadNotifyTemplate()
	if err != nil {
		return "", err
	}
	if template != "" {
		rendered := renderTemplate(template, map[string]string{
			"session":      session.Name,
			"session_id":   session.ID,
			"layout":       session.Layout,
			"repo":         session.Repo,
			"branch":       session.Branch,
			"category":     session.Category,
			"intent":       session.Intent,
			"worktree":     session.Worktree,
			"repo_dir":     session.RepoDir,
			"worktree_dir": session.WorktreeDir,
		})
		if strings.TrimSpace(rendered) != "" {
			b.WriteString("\n")
			b.WriteString(rendered)
			if !strings.HasSuffix(rendered, "\n") {
				b.WriteString("\n")
			}
		}
	}

	return b.String(), nil
}

func loadNotifyTemplate() (string, error) {
	cfg := app.Config.Notify
	if cfg.TemplateFile != "" {
		data, err := os.ReadFile(cfg.TemplateFile)
		if err != nil {
			return "", err
		}
		return string(data), nil
	}
	return cfg.Template, nil
}

func renderTemplate(template string, values map[string]string) string {
	out := template
	for key, val := range values {
		placeholder := "{{" + key + "}}"
		out = strings.ReplaceAll(out, placeholder, val)
	}
	return out
}

func buildNotifyURL(host, topic string) string {
	host = strings.TrimSpace(host)
	topic = strings.TrimSpace(topic)
	if host == "" || topic == "" {
		return ""
	}
	host = strings.TrimRight(host, "/")
	topic = strings.TrimLeft(topic, "/")
	if !strings.HasPrefix(host, "http://") && !strings.HasPrefix(host, "https://") {
		host = "http://" + host
	}
	return host + "/" + topic
}

func sendNotify(ctx context.Context, url, message string) error {
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, strings.NewReader(message))
	if err != nil {
		return err
	}
	req.Header.Set("Markdown", "yes")
	req.Header.Set("Content-Type", "text/markdown")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("notify failed: %s", strings.TrimSpace(string(body)))
	}
	return nil
}

func isWithin(child, parent string) bool {
	if parent == "" {
		return false
	}
	parentAbs, err := filepath.Abs(parent)
	if err != nil {
		return false
	}
	parentAbs = filepath.Clean(parentAbs)
	childAbs, err := filepath.Abs(child)
	if err != nil {
		return false
	}
	childAbs = filepath.Clean(childAbs)
	if childAbs == parentAbs {
		return true
	}
	rel, err := filepath.Rel(parentAbs, childAbs)
	if err != nil {
		return false
	}
	return rel != ".." && !strings.HasPrefix(rel, ".."+string(os.PathSeparator))
}
