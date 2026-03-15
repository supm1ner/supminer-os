# SupMiner OS - root bashrc (live environment)
export PS1='\[\033[1;37m\][\[\033[0;37m\]\u@\h\[\033[1;37m\]] \[\033[0;90m\]\w\[\033[0m\] \[\033[1;37m\]▶\[\033[0m\] '

# Run fastfetch on first interactive shell
if [[ -z "$SUPMINER_FETCHED" ]]; then
    export SUPMINER_FETCHED=1
    fastfetch 2>/dev/null || true
fi

alias ll='ls -la --color=auto'
alias la='ls -A --color=auto'
alias l='ls -CF --color=auto'
alias grep='grep --color=auto'
alias install='sudo /usr/local/bin/install'
