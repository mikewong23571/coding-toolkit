# owlx

_resume your intent — faster._

Minimal tmux + git-worktree session runtime built around a single idea:

1. Intent (in your head)
2. `owlx new` (tmux session + git worktree + metadata)
3. You work
4. `owlx resume` (instantly back to that world)

## Usage

```sh
owlx new [--no-attach] [-C <layout/repo>] <category> <worktree> <intent...>
owlx ls [-o table|wide|json] [--no-header]
owlx search|s [--cd]
owlx resume <session|id>
owlx view <session|id>
owlx del <session|id>
owlx notify [--session <session|id>] [--stdin|-] [--topic <name>|--url <url>] <json>
owlx config [init|edit|gitignore]
owlx completion [bash|zsh|fish|powershell]
owlx status [on|off|toggle] [--session <session|id>]
owlx version
```

## Layouts

- `main`
- `fork`
- `playground`
- `archive`

## Notes

- `owlx new` must be run from a repo under `OXL_ROOT/<layout>/<repo>` unless `-C` is provided.
- owlx uses its own tmux socket and config under `~/.owlx/tmux/default/` by default.
  It always uses the embedded tmux binary on supported platforms.
- Embedded tmux is extracted to `~/.local/share/owlx/<buildflag>/libexec/tmux`.
  Buildflag format: commit[:8] + "-" + build timestamp.

## tmux requirements

owlx requires tmux with the following capabilities:

- Platform: embedded tmux is supported on `linux/amd64`, `darwin/amd64`,
  and `darwin/arm64` (system tmux is not used).
- Binary format: embedded tmux must match platform executable format
  (ELF on Linux, Mach-O on macOS).
- Flags: tmux must support `-S <socket>` and `-f <config>` (owlx always supplies both).
- Commands used:
  - `new-session -d -s <name> -c <dir>`
  - `attach -t <name>`
  - `has-session`, `ls -F '#S'`
  - `list-panes -F '#{pane_id}'`
  - `split-window -h/-v -p <pct> -t <target> -c <dir> -P -F '#{pane_id}'`
  - `select-pane -t <pane>`
  - `set-option`, `show-option -qv`, `display-message -p`
  - `kill-session`, `kill-pane`
- Linkage: Linux build is static (musl + libevent + ncurses). macOS builds are
  fixed-version pinned-source builds.

See `docs/embedded-tmux.md` for details on the embedded tmux binaries and how to update them.

## LS output

`owlx ls -o json` prints a JSON array of objects with keys:
`session`, `id`, `layout`, `repo`, `branch`, `category`, `intent`.

## Search

```sh
owlx search        # print repo path (use: cd "$(owlx search)")
owlx search --cd   # print shell cd command (use: eval "$(owlx search --cd)")
```

Requires `fzf`.

## Categories

Keep these high-level and orthogonal:

- `feat`     - new or expanded capability
- `fix`      - bug or regression correction
- `refactor` - structure change without behavior change
- `research` - exploration / uncertainty reduction
- `chore`    - support work (deps, docs, infra, cleanup)

## Environment

```sh
OXL_ROOT=~/projs                # default
OXL_WT_DIRNAME=.worktrees       # default
~/.owlxrc                        # optional; shell-style overrides
```

Example `~/.owlxrc`:

```sh
OXL_ROOT="$HOME/projs"
OXL_WT_DIRNAME=".worktrees"
```

Notify env (optional):

```sh
OXL_NOTIFY_HOST="https://ntfy.local"
OXL_NOTIFY_TOPIC="owlx-alert"
OXL_NOTIFY_TEMPLATE_FILE="$HOME/.owlx-notify-template.md"
OXL_NOTIFY_ACTIONS="view, Open session, http://10.0.0.1/s/{{.SessionID}}"
OXL_NOTIFY_ACTIONS_FILE="$HOME/.owlx-notify-actions.txt"
OXL_NOTIFY_ALLOW_OUTSIDE=1
```

## Helpers

### Config helpers

```sh
owlx config          # show effective values + config path
owlx config init     # write a starter ~/.owlxrc
owlx config edit     # open ~/.owlxrc in $EDITOR
owlx config gitignore
owlx config gitignore install
owlx completion bash
```

### Version

```sh
owlx version
```

### Status helpers

```sh
owlx status          # toggle status-line info (current tmux session)
owlx status on       # enable status-line info
owlx status off      # disable status-line info
```

### Notify helpers

`owlx notify` posts Markdown to ntfy. Use `--topic` or `--url` to override the
default destination.

```sh
owlx notify --topic my-topic '{"type":"agent-turn-complete","cwd":"/path"}'
```

Notes:

- Adds a second status line, leaving the default tmux line untouched.
- Default left widths: layout=10, cat=8, repo=20, branch=20.
- Notify templates use Go `text/template` with fields like `.SessionID`, `.Repo`, `.Branch`, `.WorktreeDir`, and `.Payload`.

## Examples

```sh
cd ~/projs/main/coding-toolkit
owlx new research feat-a "read docs and decide minimal MVP"
owlx new --no-attach research feat-a "prep worktree only"
owlx new -C main/coding-toolkit research feat-b "prep worktree only"
cd "$(owlx search)"
owlx ls
owlx ls -o json
owlx resume main/coding-toolkit/feat-a
owlx resume a1b2c3
owlx view a1b2c3
owlx del a1b2c3
```

## Markdown lint

Run the lint script:

```sh
scripts/lint-md.sh
```

This uses `npx` to run `markdownlint-cli2` with the repo config.
Edit `.markdownlint.jsonc` to change rules, or set `MARKDOWNLINT_CONFIG` /
`MARKDOWNLINT_GLOB` to override the defaults.

## Development

Lint:

- `make lint` (sh + md + go)
- `make lint-sh` (ShellCheck)
- `make lint-md` (Markdown)
- `make lint-go` (gofmt check + `go vet`)

Go:

- Go 1.21+ is required when running from the repo (build `./owlx` with `make build`).

Makefile workflow:

- `make build` (builds `./owlx` with commit/timestamp ldflags)
- `make test`
- `make lint`
- `make tmux-update` (builds embedded tmux via `scripts/build-embedded-tmux.sh`)
- `make tmux-verify`
- `make clean` (removes build outputs and embedded tmux artifacts)

`make build` always builds/updates the embedded tmux first. You can provide an
existing artifact with `TMUX_STRIPPED=/path/to/tmux.*.stripped.gz`. You can also
set `CLEAN_BUILD=1` to remove the build cache after a successful build.

Embedded tmux versions and defaults live in `assets/tmux/buildinfo.env`. Update it
to bump tmux/libevent/ncurses/musl or defaults, then run `make tmux-update`.
`owlx version` prints the embedded tmux metadata.
