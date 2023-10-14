#!/usr/bin/env bash

# Bash Shell commandline completions for gron
#
# Stick this file in your ~/.bash_completion file
#
# You can run the command: $ mv ./gron.bash ~/.bash_completion

function _gron_completion {
  local AVAILABLE_COMMANDS="-c --colorize -j --json -k --insecure -m --monochrome -s --stream -u --ungron --no-sort --version"
  COMPREPLY=()

  if [ "$COMP_CWORD" -eq 1 ]; then
    local CURRENT_WORD=${COMP_WORDS[COMP_CWORD]}
    COMPREPLY=($(compgen -W "$AVAILABLE_COMMANDS" -- $CURRENT_WORD))
  fi
}

complete -F _gron_completion gron
