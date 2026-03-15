#!/usr/bin/env bash
# Show SupMiner welcome screen on login (only for interactive shells on tty)
if [[ $- == *i* ]] && [[ -z "$DISPLAY" ]] && [[ -z "$WAYLAND_DISPLAY" ]]; then
    /usr/local/bin/supminer-welcome
fi
