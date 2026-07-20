package cli

import (
	"fmt"
	"os"
	"strings"
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
`

func cmdShell(args []string) {
	install := false
	shell := "bash"
	for _, a := range args {
		if a == "--install" {
			install = true
		} else {
			shell = a
		}
	}

	if install {
		installShellWidget(shell)
		return
	}

	switch shell {
	case "bash":
		os.Stdout.WriteString(shellBash)
	case "zsh":
		os.Stdout.WriteString(shellZsh)
	default:
		fmt.Fprintf(os.Stderr, "unknown shell %q. Use: bash or zsh.\n", shell)
		os.Exit(1)
	}
}

func installShellWidget(shell string) {
	if shell == "fish" {
		fmt.Fprintln(os.Stderr, "error: widget not supported for fish")
		os.Exit(1)
	}

	rcFile := rcFilePath(shell)
	if rcFile == "" {
		fmt.Fprintf(os.Stderr, "error: no config file found for %s\n", shell)
		os.Exit(1)
	}

	shellArg := shell
	if shell == "bash" || shell == "zsh" {
		shellArg = shell
	} else {
		shellArg = "bash"
	}

	widgetCmd := fmt.Sprintf("source <(decoreba shell %s)", shellArg)
	block := fmt.Sprintf("%s\n%s\n%s\n", blockBegin, widgetCmd, blockEnd)

	existing, err := os.ReadFile(rcFile)
	if err != nil && !os.IsNotExist(err) {
		fmt.Fprintf(os.Stderr, "error reading %s: %v\n", rcFile, err)
		os.Exit(1)
	}

	content := string(existing)
	hasBlock := strings.Contains(content, blockBegin)

	if hasBlock {
		start := strings.Index(content, blockBegin)
		end := strings.Index(content, blockEnd)
		if end >= 0 {
			end += len(blockEnd)
		}
		before := content[:start]
		after := ""
		if end < len(content) {
			after = content[end:]
		}
		content = before + block + strings.TrimPrefix(after, "\n")
	} else {
		if !strings.HasSuffix(content, "\n") && content != "" {
			content += "\n"
		}
		content += "\n" + block
	}

	if err := os.WriteFile(rcFile, []byte(content), 0644); err != nil {
		fmt.Fprintf(os.Stderr, "error writing %s: %v\n", rcFile, err)
		os.Exit(1)
	}

	fmt.Printf("✔ Widget installed in %s (Ctrl+O)\n", rcFile)
}
