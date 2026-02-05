package notify

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"owlx/internal/config"
	"owlx/internal/domain"
	"owlx/internal/tmux"
	"owlx/internal/util"
)

type Options struct {
	Session  string
	UseStdin bool
	Topic    string
	URL      string
	JSON     string
}

type templateData struct {
	Session     string
	SessionID   string
	Layout      string
	Repo        string
	Branch      string
	Category    string
	Intent      string
	Worktree    string
	RepoDir     string
	WorktreeDir string
	Payload     map[string]interface{}
}

func ParseArgs(args []string) (Options, error) {
	opts := Options{}
	for len(args) > 0 {
		switch args[0] {
		case "--session":
			if len(args) < 2 {
				return opts, fmt.Errorf("--session requires <session|id>")
			}
			opts.Session = args[1]
			args = args[2:]
		case "--stdin", "-":
			opts.UseStdin = true
			args = args[1:]
		case "--topic":
			if len(args) < 2 {
				return opts, fmt.Errorf("--topic requires <name>")
			}
			opts.Topic = args[1]
			args = args[2:]
		case "--url":
			if len(args) < 2 {
				return opts, fmt.Errorf("--url requires <url>")
			}
			opts.URL = args[1]
			args = args[2:]
		default:
			if opts.JSON == "" {
				opts.JSON = args[0]
				args = args[1:]
			} else {
				return opts, fmt.Errorf("unexpected arg: %s", args[0])
			}
		}
	}
	return opts, nil
}

func Handle(cfg config.Config, tm tmux.Manager, args []string) error {
	opts, err := ParseArgs(args)
	if err != nil {
		return err
	}
	return HandleOptions(cfg, tm, opts)
}

func HandleOptions(cfg config.Config, tm tmux.Manager, opts Options) error {
	if opts.UseStdin {
		data, err := io.ReadAll(os.Stdin)
		if err != nil {
			return err
		}
		opts.JSON = strings.TrimSpace(string(data))
	}
	if opts.JSON == "" {
		return fmt.Errorf("notify requires JSON payload")
	}

	var payload map[string]interface{}
	if err := json.Unmarshal([]byte(opts.JSON), &payload); err != nil {
		return err
	}
	typ, ok := payload["type"].(string)
	if !ok || strings.TrimSpace(typ) == "" {
		return fmt.Errorf("notify payload missing type")
	}
	if typ != "agent-turn-complete" {
		return fmt.Errorf("unsupported notify payload type: %s", typ)
	}

	if !cfg.NotifyAllowOutside {
		pane := os.Getenv("TMUX_PANE")
		if pane == "" {
			return nil
		}
		current := tm.DisplayMessage(pane, "#S")
		if current == "" || !isOwlxSession(tm, current) {
			return nil
		}
	}

	sess := ""
	if opts.Session != "" {
		resolved, err := resolveSession(tm, opts.Session)
		if err != nil {
			return err
		}
		sess = resolved
	} else {
		sess = resolveNotifySession(tm, payload)
	}
	if sess == "" {
		return nil
	}

	attached := tm.DisplayMessage(sess, "#{session_attached}")
	if attached != "" {
		if n, err := strconv.Atoi(attached); err == nil && n > 0 {
			return nil
		}
	}

	url := opts.URL
	topic := opts.Topic
	if url == "" && topic == "" {
		topic = cfg.NotifyTopic
	}
	if url == "" && topic != "" {
		host := strings.TrimRight(cfg.NotifyHost, "/")
		topic = strings.TrimLeft(topic, "/")
		if host != "" {
			if !strings.Contains(host, "://") {
				host = "http://" + host
			}
			url = host + "/" + topic
		}
	}

	data := buildTemplateData(tm, sess, payload)
	body, err := renderNotifyBody(cfg, data)
	if err != nil {
		return err
	}
	actions, err := renderNotifyActions(cfg, data)
	if err != nil {
		return err
	}
	if url == "" {
		fmt.Print(body)
		return nil
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBufferString(body))
	if err != nil {
		return err
	}
	req.Header.Set("Markdown", "yes")
	req.Header.Set("Content-Type", "text/markdown")
	if actions != "" {
		req.Header.Set("Actions", actions)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		return fmt.Errorf("notify failed: %s", resp.Status)
	}
	return nil
}

func resolveSession(tm tmux.Manager, token string) (string, error) {
	if tm.HasSession(token) {
		return token, nil
	}

	sessions, err := tm.ListSessions()
	if err != nil {
		return "", err
	}
	if len(sessions) == 0 {
		return "", fmt.Errorf("no such session or id: %s", token)
	}

	match := ""
	count := 0
	for _, sess := range sessions {
		if !isOwlxSession(tm, sess) {
			continue
		}
		if ensureSessionID(tm, sess) == token {
			match = sess
			count++
		}
	}
	if count == 1 {
		return match, nil
	}
	if count > 1 {
		return "", fmt.Errorf("ambiguous id: %s", token)
	}
	return "", fmt.Errorf("no such session or id: %s", token)
}

func resolveNotifySession(tm tmux.Manager, payload map[string]interface{}) string {
	if pane := os.Getenv("TMUX_PANE"); pane != "" {
		sess := tm.DisplayMessage(pane, "#S")
		if sess != "" && isOwlxSession(tm, sess) {
			return sess
		}
	}

	cwdVal, _ := payload["cwd"].(string)
	cwd := resolvePath(cwdVal)
	if cwd == "" {
		return ""
	}

	sessions, _ := tm.ListSessions()
	var owlxSessions []string
	for _, sess := range sessions {
		if isOwlxSession(tm, sess) {
			owlxSessions = append(owlxSessions, sess)
		}
	}

	for _, sess := range owlxSessions {
		wtDir := tm.ShowOption(sess, domain.WorktreeDirKey)
		if matchPath(cwd, wtDir) {
			return sess
		}
	}
	for _, sess := range owlxSessions {
		repoDir := tm.ShowOption(sess, domain.RepoDirKey)
		if matchPath(cwd, repoDir) {
			return sess
		}
	}
	return ""
}

func resolvePath(path string) string {
	if path == "" {
		return ""
	}
	abs, err := filepath.Abs(path)
	if err != nil {
		return ""
	}
	resolved, err := filepath.EvalSymlinks(abs)
	if err != nil {
		return abs
	}
	return resolved
}

func matchPath(cwd, target string) bool {
	if target == "" {
		return false
	}
	resolved := resolvePath(target)
	if resolved == "" {
		return false
	}
	if cwd == resolved {
		return true
	}
	if strings.HasPrefix(cwd, resolved+string(os.PathSeparator)) {
		return true
	}
	return false
}

func buildTemplateData(tm tmux.Manager, sess string, payload map[string]interface{}) templateData {
	branch := tm.ShowOption(sess, domain.BranchKey)
	cat := tm.ShowOption(sess, domain.CatKey)
	intent := tm.ShowOption(sess, domain.IntentKey)
	id := ensureSessionID(tm, sess)
	repo := tm.ShowOption(sess, domain.RepoKey)
	layout := tm.ShowOption(sess, domain.LayoutKey)
	wt := tm.ShowOption(sess, domain.WorktreeKey)
	repoDir := tm.ShowOption(sess, domain.RepoDirKey)
	wtDir := tm.ShowOption(sess, domain.WorktreeDirKey)

	return templateData{
		Session:     util.TSVEscape(sess),
		SessionID:   util.TSVEscape(id),
		Layout:      util.TSVEscape(layout),
		Repo:        util.TSVEscape(repo),
		Branch:      util.TSVEscape(branch),
		Category:    util.TSVEscape(cat),
		Intent:      util.TSVEscape(intent),
		Worktree:    util.TSVEscape(wt),
		RepoDir:     util.TSVEscape(repoDir),
		WorktreeDir: util.TSVEscape(wtDir),
		Payload:     payload,
	}
}

func renderNotifyBody(cfg config.Config, data templateData) (string, error) {
	var buf strings.Builder
	fmt.Fprintf(&buf, "**session_id**: `%s`\n", data.SessionID)
	fmt.Fprintf(&buf, "**repo**: `%s`\n", data.Repo)
	fmt.Fprintf(&buf, "**branch**: `%s`\n", data.Branch)
	fmt.Fprintf(&buf, "**category**: `%s`\n", data.Category)
	fmt.Fprintf(&buf, "**intent**: %s\n", data.Intent)

	templateText, err := loadTemplateInlineOrFile(cfg.NotifyTemplate, cfg.NotifyTemplateFile, "notify template")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(templateText) != "" {
		rendered, err := renderTemplate(templateText, data)
		if err != nil {
			return "", err
		}
		if strings.TrimSpace(rendered) != "" {
			fmt.Fprintf(&buf, "\n%s\n", rendered)
		}
	}
	return buf.String(), nil
}

func renderNotifyActions(cfg config.Config, data templateData) (string, error) {
	templateText, err := loadTemplateInlineOrFile(cfg.NotifyActions, cfg.NotifyActionsFile, "notify actions template")
	if err != nil {
		return "", err
	}
	if strings.TrimSpace(templateText) == "" {
		return "", nil
	}
	rendered, err := renderTemplate(templateText, data)
	if err != nil {
		return "", err
	}
	rendered = strings.TrimSpace(rendered)
	if rendered == "" {
		return "", nil
	}
	rendered = strings.ReplaceAll(rendered, "\n", " ")
	rendered = strings.ReplaceAll(rendered, "\r", " ")
	return rendered, nil
}

func loadTemplateInlineOrFile(inline, path, label string) (string, error) {
	if strings.TrimSpace(path) == "" {
		return inline, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("%s file not found: %s", label, path)
	}
	return string(data), nil
}

func renderTemplate(templateText string, data templateData) (string, error) {
	tmpl, err := template.New("notify").
		Funcs(template.FuncMap{
			"json":  util.JSONEscape,
			"tsv":   util.TSVEscape,
			"shell": util.ShellQuote,
		}).
		Option("missingkey=zero").
		Parse(templateText)
	if err != nil {
		return "", err
	}
	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, data); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func isOwlxSession(tm tmux.Manager, sess string) bool {
	return tm.ShowOption(sess, domain.FlagKey) == "1"
}

func ensureSessionID(tm tmux.Manager, sess string) string {
	id := tm.ShowOption(sess, domain.IDKey)
	if id != "" {
		return id
	}
	id = util.GenID(sess)
	_ = tm.SetOption(sess, domain.IDKey, id)
	return id
}
