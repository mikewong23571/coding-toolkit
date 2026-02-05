//go:build linux && amd64

package assets

import (
	"embed"
	"io/fs"
)

//go:embed tmux/linux_amd64/tmux
var embeddedFS embed.FS

const tmuxPath = "tmux/linux_amd64/tmux"

// EmbeddedTmux returns the embedded tmux binary for linux/amd64.
func EmbeddedTmux() ([]byte, bool) {
	data, err := fs.ReadFile(embeddedFS, tmuxPath)
	if err != nil || len(data) == 0 {
		return nil, false
	}
	return data, true
}
