package cli

import (
	"fmt"
	"os"
	"strings"

	"owlx/internal/state"
	"owlx/internal/status"
	"owlx/internal/zellij"
)

func writeStatusFile(session state.Session) (string, error) {
	path := app.Store.StatusPath(session.ID)
	line := status.FormatLine(session.Layout, session.Category, session.Repo, session.Branch, session.Intent, app.Config.Status.RightLen)
	if err := status.WriteFile(path, line); err != nil {
		return "", err
	}
	return path, nil
}

func ensureLayout(session state.Session) (string, error) {
	if app.Config.ZellijLayout != "" {
		if _, err := os.Stat(app.Config.ZellijLayout); err != nil {
			return "", fmt.Errorf("zellij layout not found: %s", app.Config.ZellijLayout)
		}
		return app.Config.ZellijLayout, nil
	}
	layoutPath := app.Store.LayoutPath(session.ID)
	if _, err := os.Stat(layoutPath); err == nil {
		return layoutPath, nil
	}

	statusPath, err := writeStatusFile(session)
	if err != nil {
		return "", err
	}
	pluginPath := zellij.ResolveStatusPlugin(app.Config.Status.PluginPath)
	if pluginPath == "" {
		return "", fmt.Errorf("status plugin not found (set status.plugin_path or OXL_STATUS_PLUGIN)")
	}
	if _, err := os.Stat(pluginPath); err != nil {
		return "", fmt.Errorf("status plugin not found: %s", pluginPath)
	}

	opts := zellij.LayoutOptions{
		StatusPluginPath: pluginPath,
		StatusFilePath:   statusPath,
		IntervalMS:       app.Config.Status.IntervalMS,
	}
	if err := zellij.WriteLayout(layoutPath, opts); err != nil {
		return "", err
	}
	return layoutPath, nil
}

func zellijSessionName(session state.Session) string {
	name := strings.TrimSpace(session.Name)
	if name == "" {
		return ""
	}
	safe := sanitizeZellijName(name)
	id := session.ID
	if id == "" {
		id = state.GenID(name)
	}
	if safe == "" {
		return id
	}
	return fmt.Sprintf("%s-%s", safe, id)
}

func sanitizeZellijName(input string) string {
	var b strings.Builder
	b.Grow(len(input))
	for _, r := range input {
		switch {
		case r >= 'a' && r <= 'z':
			b.WriteRune(r)
		case r >= 'A' && r <= 'Z':
			b.WriteRune(r)
		case r >= '0' && r <= '9':
			b.WriteRune(r)
		case r == '-' || r == '_' || r == '.':
			b.WriteRune(r)
		default:
			b.WriteByte('-')
		}
	}
	return strings.Trim(b.String(), "-")
}
