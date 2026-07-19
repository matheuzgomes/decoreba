package cli

import (
	"fmt"
	"os"
)

const shellBash = `# decoreba shell integration for bash
#
# Add to ~/.bashrc:
#   source <(decoreba completion bash)
#   source <(decoreba shell bash)
#
# Press Ctrl+O to open decoreba. Whatever you typed becomes the search query.
#   Enter       → insert command at cursor
#   Ctrl+X      → execute command immediately

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
`

const shellZsh = `# decoreba shell integration for zsh
#
# Add to ~/.zshrc:
#   source <(decoreba shell zsh)
#
# Press Ctrl+O to open decoreba. Whatever you typed becomes the search query.
#   Enter       → insert command at cursor
#   Ctrl+X      → execute command immediately

__decoreba_widget() {
    local result
    result=$(decoreba --shell-output "$LBUFFER")
    if [[ "$result" == EXEC:* ]]; then
        LBUFFER="${result#EXEC:}"
        zle accept-line
    else
        LBUFFER="${result}"
        zle reset-prompt
    fi
}
zle -N __decoreba_widget
bindkey '^O' __decoreba_widget
`

func cmdShell(args []string) {
	shell := "bash"
	if len(args) > 0 {
		shell = args[0]
	}

	switch shell {
	case "bash":
		fmt.Print(shellBash)
	case "zsh":
		fmt.Print(shellZsh)
	default:
		fmt.Fprintf(os.Stderr, "unknown shell %q. Use: bash or zsh.\n", shell)
		os.Exit(1)
	}
}
