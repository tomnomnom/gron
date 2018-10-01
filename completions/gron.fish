#!/usr/bin/env fish

# Fish shell commandline completions for gron.
#
# Stick this file in your ~/.config/fish/completions/ directory.

complete -c gron -s u -l ungron     --description "Reverse the operation (turn assignments back into JSON)"
complete -c gron -s c -l colorize   --description "Colorize output (default on tty)"
complete -c gron -s m -l monochrome --description "Monochrome (don't colorize output)"
complete -c gron -s s -l stream     --description "Treat each line of input as a separate JSON object"
complete -c gron -s k -l insecure   --description "Disable certificate validation"
complete -c gron -s j -l json       --description "Represent gron data as JSON stream"
complete -c gron      -l no-sort    --description "Don't sort output (faster)"
complete -c gron      -l version    --description "Print version information"

# eof
