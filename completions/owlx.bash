_owlx_layout_repo_list() {
  local root="${OXL_ROOT:-$HOME/projs}"
  local layout dir name
  for layout in main fork playground archive; do
    [ -d "$root/$layout" ] || continue
    for dir in "$root/$layout"/*; do
      [ -d "$dir" ] || continue
      name="${dir##*/}"
      printf '%s/%s\n' "$layout" "$name"
    done
  done
}

_owlx() {
  local cur subcmd
  cur="${COMP_WORDS[COMP_CWORD]}"
  subcmd="${COMP_WORDS[1]}"

  if [ "$COMP_CWORD" -eq 1 ]; then
    COMPREPLY=( $(compgen -W "new ls search s resume view del config panel status help" -- "$cur") )
    return 0
  fi

  case "$subcmd" in
    new)
      if [[ "${COMP_WORDS[COMP_CWORD-1]}" == "-C" ]]; then
        COMPREPLY=( $(compgen -W "$(_owlx_layout_repo_list)" -- "$cur") )
        return 0
      fi
      local i=2
      local cat_index=2
      while [ $i -lt ${#COMP_WORDS[@]} ]; do
        case "${COMP_WORDS[$i]}" in
          -d|--detached|--no-attach)
            i=$((i+1))
            cat_index=$((cat_index+1))
            ;;
          -C)
            i=$((i+2))
            cat_index=$((cat_index+2))
            ;;
          --)
            i=$((i+1))
            cat_index=$((cat_index+1))
            break
            ;;
          -*) break ;;
          *) break ;;
        esac
      done
      case "$COMP_CWORD" in
        2)
          if [[ "$cur" == -* ]]; then
            COMPREPLY=( $(compgen -W "--no-attach --detached -d -C" -- "$cur") )
          else
            COMPREPLY=( $(compgen -W "fix feat refactor research chore" -- "$cur") )
          fi
          ;;
        "$cat_index") COMPREPLY=( $(compgen -W "fix feat refactor research chore" -- "$cur") ) ;;
        *) COMPREPLY=() ;;
      esac
      ;;
    search|s)
      if [ "$COMP_CWORD" -eq 2 ]; then
        COMPREPLY=( $(compgen -W "--cd" -- "$cur") )
      else
        COMPREPLY=()
      fi
      ;;
    config)
      if [ "$COMP_CWORD" -eq 2 ]; then
        COMPREPLY=( $(compgen -W "init edit tmux gitignore completion" -- "$cur") )
      elif [ "$COMP_CWORD" -eq 3 ]; then
        case "${COMP_WORDS[2]}" in
          tmux|gitignore)
            COMPREPLY=( $(compgen -W "show install" -- "$cur") )
            ;;
          completion)
            COMPREPLY=( $(compgen -W "install uninstall" -- "$cur") )
            ;;
          *)
            COMPREPLY=()
            ;;
        esac
      else
        COMPREPLY=()
      fi
      ;;
    panel)
      COMPREPLY=( $(compgen -W "on off toggle --session --width" -- "$cur") )
      ;;
    status)
      COMPREPLY=( $(compgen -W "on off toggle --session" -- "$cur") )
      ;;
    resume|view|del)
      if command -v tmux >/dev/null 2>&1; then
        COMPREPLY=( $(compgen -W "$(tmux ls -F '#S' 2>/dev/null)" -- "$cur") )
      else
        COMPREPLY=()
      fi
      ;;
    *)
      COMPREPLY=()
      ;;
  esac
}

complete -F _owlx owlx
