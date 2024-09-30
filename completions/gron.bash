#!/usr/bin/env bash

# Bash shell commandline completions for gron.
#
# Copy the contents of this file into your ~/.bashrc file or whatever
# file you use for Bash completions.
#
# Example: cat ./completions/gron.bash >> ~/.bashrc

function _gron_completion {
  local AVAILABLE_COMMANDS="--colorize --insecure --json --monochrome --no-sort --stream --ungron --values --version"
  COMPREPLY=()

  local CURRENT_WORD=${COMP_WORDS[COMP_CWORD]}
  COMPREPLY=($(compgen -W "$AVAILABLE_COMMANDS" -- "$CURRENT_WORD"))
}

complete -F _gron_completion gron
