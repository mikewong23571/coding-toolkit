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

	body, err := printNotifyMeta(cfg, tm, sess)
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

func printNotifyMeta(cfg config.Config, tm tmux.Manager, sess string) (string, error) {
	branch := tm.ShowOption(sess, domain.BranchKey)
	cat := tm.ShowOption(sess, domain.CatKey)
	intent := tm.ShowOption(sess, domain.IntentKey)
	id := ensureSessionID(tm, sess)
	repo := tm.ShowOption(sess, domain.RepoKey)
	layout := tm.ShowOption(sess, domain.LayoutKey)
	wt := tm.ShowOption(sess, domain.WorktreeKey)
	repoDir := tm.ShowOption(sess, domain.RepoDirKey)
	wtDir := tm.ShowOption(sess, domain.WorktreeDirKey)

	sessS := util.TSVEscape(sess)
	idS := util.TSVEscape(id)
	repoS := util.TSVEscape(repo)
	branchS := util.TSVEscape(branch)
	catS := util.TSVEscape(cat)
	intentS := util.TSVEscape(intent)
	layoutS := util.TSVEscape(layout)
	wtS := util.TSVEscape(wt)
	repoDirS := util.TSVEscape(repoDir)
	wtDirS := util.TSVEscape(wtDir)

	var buf strings.Builder
	fmt.Fprintf(&buf, "**session_id**: `%s`\n", idS)
	fmt.Fprintf(&buf, "**repo**: `%s`\n", repoS)
	fmt.Fprintf(&buf, "**branch**: `%s`\n", branchS)
	fmt.Fprintf(&buf, "**category**: `%s`\n", catS)
	fmt.Fprintf(&buf, "**intent**: %s\n", intentS)

	template, err := loadTemplate(cfg)
	if err != nil {
		return "", err
	}
	if template != "" {
		rendered := util.RenderTemplate(template, map[string]string{
			"session":      sessS,
			"session_id":   idS,
			"layout":       layoutS,
			"repo":         repoS,
			"branch":       branchS,
			"category":     catS,
			"intent":       intentS,
			"worktree":     wtS,
			"repo_dir":     repoDirS,
			"worktree_dir": wtDirS,
		})
		if rendered != "" {
			fmt.Fprintf(&buf, "\n%s\n", rendered)
		}
	}
	return buf.String(), nil
}

func loadTemplate(cfg config.Config) (string, error) {
	template := cfg.NotifyTemplate
	if cfg.NotifyTemplateFile == "" {
		return template, nil
	}
	data, err := os.ReadFile(cfg.NotifyTemplateFile)
	if err != nil {
		return "", fmt.Errorf("notify template file not found: %s", cfg.NotifyTemplateFile)
	}
	return string(data), nil
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
