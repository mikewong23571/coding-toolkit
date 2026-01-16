_owlx() {
  local cur subcmd
  cur="${COMP_WORDS[COMP_CWORD]}"
  subcmd="${COMP_WORDS[1]}"

  if [ "$COMP_CWORD" -eq 1 ]; then
    COMPREPLY=( $(compgen -W "new ls resume view del config panel help" -- "$cur") )
    return 0
  fi

  case "$subcmd" in
    new)
      case "$COMP_CWORD" in
        2) COMPREPLY=( $(compgen -W "main fork playground archive" -- "$cur") ) ;;
        5) COMPREPLY=( $(compgen -W "fix feat refactor research chore" -- "$cur") ) ;;
        *) COMPREPLY=() ;;
      esac
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
