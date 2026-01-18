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
owlx ls
owlx search|s [--cd]
owlx resume <session|id>
owlx view <session|id>
owlx del <session|id>
owlx config [init|edit|tmux|gitignore|completion]
owlx config completion [install|uninstall]
owlx panel [on|off|toggle] [--session <session|id>] [--width <pct>]
owlx status [on|off|toggle] [--session <session|id>]
```

## Layouts

- `main`
- `fork`
- `playground`
- `archive`

## Notes

- `owlx new` must be run from a repo under `OXL_ROOT/<layout>/<repo>` unless `-C` is provided.

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
OXL_PANEL_ON_NEW=1              # default; open panel on new
OXL_PANEL_WIDTH=22              # default; percent
OXL_LEFT_PANE_BOTTOM_PCT=25     # default; left bottom pane percent
OXL_STATUS_ON_NEW=0             # default; show info in status line on new
OXL_STATUS_LINES=2              # default; status rows when enabled
OXL_STATUS_RIGHT_LEN=120        # default; intent max length
~/.owlxrc                        # optional; shell-style overrides
```

Example `~/.owlxrc`:

```sh
OXL_ROOT="$HOME/projs"
OXL_WT_DIRNAME=".worktrees"
```

## Helpers

### Config helpers

```sh
owlx config          # show effective values + config path
owlx config init     # write a starter ~/.owlxrc
owlx config edit     # open ~/.owlxrc in $EDITOR
owlx config tmux     # print tmux status snippet
owlx config tmux install
owlx config gitignore
owlx config gitignore install
owlx config completion
owlx config completion install
owlx config completion uninstall
```

### Panel helpers

```sh
owlx panel           # toggle right-side info panel (current tmux session)
owlx panel on        # open panel
owlx panel off       # close panel
owlx panel --width 18
```

### Status helpers

```sh
owlx status          # toggle status-line info (current tmux session)
owlx status on       # enable status-line info
owlx status off      # disable status-line info
```

Notes:

- `status on` closes the panel; when `OXL_STATUS_ON_NEW=1`, the panel won't auto-open.
- Adds a second status line, leaving the default tmux line untouched.
- Default left widths: layout=10, cat=8, repo=20, branch=20 (override via `OXL_STATUS_LEFT_FMT`).

## Examples

```sh
cd ~/projs/main/coding-toolkit
owlx new research feat-a "read docs and decide minimal MVP"
owlx new --no-attach research feat-a "prep worktree only"
owlx new -C main/coding-toolkit research feat-b "prep worktree only"
cd "$(owlx search)"
owlx ls
owlx resume main/coding-toolkit/feat-a
owlx resume a1b2c3
owlx view a1b2c3
owlx del a1b2c3
```

## Markdown lint

Run the lint script:

```sh
bin/lint-md
```

This uses `npx` to run `markdownlint-cli2` with the repo config.
