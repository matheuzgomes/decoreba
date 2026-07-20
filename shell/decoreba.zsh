# decoreba shell integration for zsh
#
# Add to ~/.zshrc:
#   source <(decoreba shell zsh)
#
# Press Ctrl+O to open decoreba. Whatever you typed becomes the search query.
#   Enter       → insert command at cursor
#   Ctrl+X      → execute command immediately (added to history)

__decoreba_widget() {
    local result
    result=$(decoreba --shell-output "$LBUFFER")
    if [[ "$result" == "EXEC:"* ]]; then
        LBUFFER="${result#EXEC:}"
        zle accept-line
    else
        LBUFFER="${result#✓ }"
        zle reset-prompt
    fi
}
zle -N __decoreba_widget
bindkey '^O' __decoreba_widget
