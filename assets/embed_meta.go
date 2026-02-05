package assets

import (
	"embed"
	"io/fs"
)

//go:embed tmux/buildinfo.env
var metaFS embed.FS

const metaPath = "tmux/buildinfo.env"

// EmbeddedTmuxMeta returns the embedded tmux build metadata.
func EmbeddedTmuxMeta() ([]byte, bool) {
	data, err := fs.ReadFile(metaFS, metaPath)
	if err != nil || len(data) == 0 {
		return nil, false
	}
	return data, true
}
