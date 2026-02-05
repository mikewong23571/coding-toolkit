package build

import (
	"os"
	"os/exec"
	"runtime/debug"
	"strconv"
	"strings"
	"time"
)

// Commit and Timestamp can be set via -ldflags.
var Commit = "dev"
var Timestamp = ""

// BuildFlag returns commit[:8] + "-" + timestamp (unix seconds).
func BuildFlag() string {
	commit := sanitizeCommit(Commit)
	ts := strings.TrimSpace(Timestamp)

	if commit == "" || commit == "dev" || ts == "" {
		bc, bt := buildInfo()
		if (commit == "" || commit == "dev") && bc != "" {
			commit = sanitizeCommit(bc)
		}
		if ts == "" && bt != "" {
			ts = bt
		}
	}

	if commit == "" || commit == "dev" {
		if c, err := gitCommit(); err == nil && c != "" {
			commit = sanitizeCommit(c)
		}
	}
	if commit == "" {
		commit = "dev"
	}
	if len(commit) > 8 {
		commit = commit[:8]
	}

	if ts == "" {
		ts = strconv.FormatInt(time.Now().Unix(), 10)
	}
	return commit + "-" + ts
}

func gitCommit() (string, error) {
	cmd := exec.Command("git", "rev-parse", "--short=8", "HEAD")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func sanitizeCommit(commit string) string {
	commit = strings.TrimSpace(commit)
	commit = strings.TrimPrefix(commit, "v")
	commit = strings.TrimPrefix(commit, "V")
	commit = strings.Trim(commit, "-")
	commit = strings.ReplaceAll(commit, "/", "")
	commit = strings.ReplaceAll(commit, " ", "")
	commit = strings.ReplaceAll(commit, "\n", "")
	if commit == "" {
		return ""
	}
	return commit
}

func buildInfo() (string, string) {
	info, ok := debug.ReadBuildInfo()
	if !ok {
		return "", ""
	}
	commit := ""
	ts := ""
	for _, setting := range info.Settings {
		switch setting.Key {
		case "vcs.revision":
			commit = setting.Value
		case "vcs.time":
			if t, err := time.Parse(time.RFC3339, setting.Value); err == nil {
				ts = strconv.FormatInt(t.Unix(), 10)
			}
		}
	}
	return commit, ts
}

func init() {
	if Commit == "" {
		Commit = os.Getenv("OXL_BUILD_COMMIT")
	}
	if Timestamp == "" {
		Timestamp = os.Getenv("OXL_BUILD_TS")
	}
}
