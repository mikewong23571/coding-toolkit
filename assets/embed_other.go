//go:build !((linux && amd64) || (darwin && amd64) || (darwin && arm64))

package assets

// EmbeddedTmux returns no embedded binary on unsupported platforms.
func EmbeddedTmux() ([]byte, bool) {
	return nil, false
}
