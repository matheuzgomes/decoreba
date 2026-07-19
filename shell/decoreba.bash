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
    local result
    result=$(decoreba --shell-output "$READLINE_LINE")
    if [[ "$result" == EXEC:* ]]; then
        local cmd="${result#EXEC:}"
        history -s "$cmd"
        eval "$cmd"
    else
        READLINE_LINE="${result}"
        READLINE_POINT=${#result}
    fi
}
bind -x '"\C-o": __decoreba_widget'
