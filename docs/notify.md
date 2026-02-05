# Notify configuration example

This document shows an example `~/.owlxrc` and a matching notify template file
using Go `text/template` syntax.

## ~/.owlxrc

```sh
# owlx config (shell-style assignments)
OXL_ROOT="$HOME/projs"
OXL_WT_DIRNAME=".worktrees"
OXL_NOTIFY_HOST="https://ntfy.local"
OXL_NOTIFY_TOPIC="owlx-alert"
OXL_NOTIFY_TEMPLATE_FILE="$HOME/.owlx-notify-template.md"
OXL_NOTIFY_ACTIONS="view, Open session, http://10.134.94.200:4000/s/{{.SessionID}}"
```

## ~/.owlx-notify-template.md

```md
## owlx notify

**Session**: `{{.Session}}`  
**Session ID**: `{{.SessionID}}`

**Repo**: `{{.Repo}}`  
**Branch**: `{{.Branch}}`  
**Category**: `{{.Category}}`  
**Intent**: {{.Intent}}
```

## Template fields

Common fields:

- `.Session`
- `.SessionID`
- `.Repo`
- `.Branch`
- `.Category`
- `.Intent`
- `.Layout`
- `.Worktree`
- `.RepoDir`
- `.WorktreeDir`
- `.Payload` (the original notify JSON payload)

Template functions:

- `json` (escape for JSON)
- `tsv` (escape newlines/tabs)
- `shell` (shell-quote a string)
