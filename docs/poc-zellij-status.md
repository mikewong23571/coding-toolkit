# Zellij status POC (manual steps)

This POC validates that a custom status bar plugin can render a tmux-like
single-line status and read owlx state from disk.

## Prereqs
- Install the fixed zellij release (pin during product build).
- Ensure the zellij CLI is on PATH.
  - Verified with zellij 0.43.1.
- Use `scripts/fetch-zellij.sh` to download the pinned binary to `tools/zellij/zellij`.
- Ensure Rust toolchain can build WASM. On newer toolchains, use `wasm32-wasip1`.
  - Use `scripts/build-status-plugin.sh` to build `tools/zellij/owlx-status.wasm`.

## 1) Verify layout loading (KDL)
```sh
zellij setup --dump-layout default > /tmp/owlx-default.kdl
zellij --layout /tmp/owlx-default.kdl
```
Expected: zellij starts with the default layout from the dumped KDL.
Note: the dumped layout should include a `status-bar` plugin entry.

## 2) Verify plugin aliasing
In ~/.config/zellij/config.kdl add or update plugins block:
```kdl
plugins {
  status-bar location="zellij:status-bar"
}
```
Expected: zellij starts normally with the default status bar.

## 3) Replace status bar with a custom plugin
Build the local POC plugin:
```sh
cd poc/owlx-status-plugin
rustup target add wasm32-wasip1
cargo build --target wasm32-wasip1 --release
```

The plugin artifact is:
`poc/owlx-status-plugin/target/wasm32-wasip1/release/owlx_status_plugin.wasm`

Then update the plugins block to point to the custom plugin:
```kdl
plugins {
  status-bar location="file:/abs/path/to/owlx-status.wasm"
}
```
Expected: zellij loads the custom status bar.

Alternative: use the provided layout file `poc/owlx-status.kdl`, which references
`tools/zellij/owlx-status.wasm` and `poc/status-line.txt`.

## 4) Verify file access from plugin
- Create a test file:
```sh
echo "layout=main repo=demo intent=hello" > /tmp/owlx-status.txt
```
- Update the plugin config (if supported) to read this file path.
Expected: status bar renders contents based on the file.

## 5) Validate refresh
- Update the file and confirm the status bar refreshes on interval.
Expected: changes appear without restarting zellij.

## Notes
- If the plugin cannot read local files, fallback is to pass status data via
  plugin config and refresh on a timer (owlx updates config + restarts session).
- If status bar cannot be replaced, reconsider scope or fall back to a minimal
  status display (tab titles + CLI output).
