# decoreba shell integration for bash
#
# Add to ~/.bashrc:
#   source <(decoreba completion bash)
#   source <(decoreba shell bash)
#
# Press Ctrl+O to open decoreba. Whatever you typed becomes the search query.
#   Enter       → insert command at cursor
#   Ctrl+X      → execute command immediately (added to history)

__decoreba_widget() {
    local result cmd ret
    result=$(decoreba --shell-output "$READLINE_LINE")
    if [[ "$result" == "EXEC:"* ]]; then
        cmd="${result#EXEC:}"
        history -s "$cmd"
        eval "$cmd"
        ret=$?
        if [[ $ret -ne 0 ]]; then
            printf "\r\033[K\033[33m⚠ Command failed (exit %d)\033[0m\n" $ret
        fi
    else
        cmd="${result#✓ }"
        READLINE_LINE="$cmd"
        READLINE_POINT=${#cmd}
    fi
}
bind -x '"\C-o": __decoreba_widget'
