# owlx
resume your intent — faster.

Minimal tmux + git-worktree session runtime built around a single idea:

Intent (in your head)
  -> owlx new (tmux session + git worktree + metadata)
  -> You work
  -> owlx resume (instantly back to that world)

## Usage
  owlx new <layout> <repo> <worktree> <category> <intent...>
  owlx ls
  owlx resume <session|id>
  owlx view <session|id>
  owlx del <session|id>
  owlx config [init|edit|tmux|gitignore|completion]
  owlx config completion [install|uninstall]
  owlx panel [on|off|toggle] [--session <session|id>] [--width <pct>]

Layouts:
  main | fork | playground | archive

Categories (keep these high-level and orthogonal):
  feat      - new or expanded capability
  fix       - bug or regression correction
  refactor  - structure change without behavior change
  research  - exploration / uncertainty reduction
  chore     - support work (deps, docs, infra, cleanup)

Environment:
  OXL_ROOT=~/projs          (default)
  OXL_WT_DIRNAME=.worktrees (default)
  OXL_PANEL_ON_NEW=1        (default; open panel on new)
  OXL_PANEL_WIDTH=22        (default; percent)
  OXL_LEFT_PANE_BOTTOM_PCT=25 (default; left bottom pane percent)
  ~/.owlxrc                 (optional; shell-style overrides)

Example ~/.owlxrc:
  OXL_ROOT="$HOME/projs"
  OXL_WT_DIRNAME=".worktrees"

Config helpers:
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

Panel helpers:
  owlx panel           # toggle right-side info panel (current tmux session)
  owlx panel on        # open panel
  owlx panel off       # close panel
  owlx panel --width 18

Examples:
  owlx new main coding-toolkit feat-a research "read docs and decide minimal MVP"
  owlx ls
  owlx resume main/coding-toolkit/feat-a
  owlx resume a1b2c3
  owlx view a1b2c3
  owlx del a1b2c3
