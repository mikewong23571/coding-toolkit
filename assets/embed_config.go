package assets

import (
	"embed"
	"io/fs"
)

//go:embed tmux/default.conf
var configFS embed.FS

const configPath = "tmux/default.conf"

// EmbeddedTmuxConfig returns the embedded default tmux config.
func EmbeddedTmuxConfig() ([]byte, bool) {
	data, err := fs.ReadFile(configFS, configPath)
	if err != nil || len(data) == 0 {
		return nil, false
	}
	return data, true
}
