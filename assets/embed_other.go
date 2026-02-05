//go:build !(linux && amd64)

package assets

// EmbeddedTmux returns no embedded binary on unsupported platforms.
func EmbeddedTmux() ([]byte, bool) {
	return nil, false
}
