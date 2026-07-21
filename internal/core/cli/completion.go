package cli

import (
	"fmt"
	"os"
)

const completionBash = `# bash completion for decoreba
# Install: source <(decoreba completion bash)
#   or copy to a directory in $fpath

_decoreba_completion() {
    local cur prev words cword
    _init_completion || return

    if ((cword == 1)); then
        COMPREPLY=($(compgen -W "add list ls rm remove edit stats init shell sync mcp export import version help completion" -- "$cur"))
        return
    fi

    case "${words[1]}" in
        list|ls)
            COMPREPLY=($(compgen -W "$(decoreba list 2>/dev/null | sed -n 's/^  - \([^ ]*\).*/\1/p')" -- "$cur"))
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "$cur"))
            ;;
        shell)
            COMPREPLY=($(compgen -W "bash zsh" -- "$cur"))
            ;;
        sync)
            COMPREPLY=($(compgen -W "init status push pull" -- "$cur"))
            ;;
    esac
}

complete -F _decoreba_completion decoreba
`

const completionZsh = `#compdef decoreba
# zsh completion for decoreba
# Install: source <(decoreba completion zsh)

_decoreba() {
    local -a subcmds
    subcmds=(
        'add:Add a new command'
        'list:List contexts and commands'
        'ls:List contexts and commands (alias)'
        'rm:Remove a command by id'
        'remove:Remove a command by id (alias)'
        'edit:Edit a command by id'
        'stats:Show vault statistics'
        'init:Bootstrap decoreba interactively'
        'shell:Print shell integration'
        'sync:Sync commands via Gist'
        'mcp:MCP server for AI agents'
        'export:Export commands to stdout'
        'import:Import commands from file'
        'version:Show version'
        'help:Show help'
        'completion:Generate shell completion'
    )
    _arguments -C '1: :{_describe command subcmds}' '*:: :->rest'

    case "$state" in
        rest)
            case "${words[2]}" in
                list|ls)
                    local -a contexts
                    contexts=(${(f)"$(decoreba list 2>/dev/null | sed -n 's/^  - \([^ ]*\).*/\1/p')"})
                    _describe 'context' contexts
                    ;;
                completion)
                    _values 'shell' bash zsh fish
                    ;;
                shell)
                    _values 'shell' bash zsh
                    ;;
                sync)
                    _values 'command' init status push pull
                    ;;
            esac
            ;;
    esac
}
`

const completionFish = `# fish completion for decoreba
# Install: decoreba completion fish > ~/.config/fish/completions/decoreba.fish

complete -c decoreba -f
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'add' -d 'Add a new command'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'list' -d 'List contexts and commands'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'ls' -d 'List contexts and commands (alias)'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'rm' -d 'Remove a command by id'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'remove' -d 'Remove a command by id (alias)'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'edit' -d 'Edit a command by id'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'stats' -d 'Show vault statistics'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'init' -d 'Bootstrap decoreba interactively'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'shell' -d 'Print shell integration'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'sync' -d 'Sync commands via Gist'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'mcp' -d 'MCP server for AI agents'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'export' -d 'Export commands to stdout'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'import' -d 'Import commands from file'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'version' -d 'Show version'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'help' -d 'Show help'
complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats init shell sync mcp export import version help completion' -a 'completion' -d 'Generate shell completion'
complete -c decoreba -n '__fish_seen_subcommand_from list ls; and test (count (commandline -opc)) -eq 2' -a '(decoreba list 2>/dev/null | string match -r "  - \w+" | string replace -r "  - " "")'
`

func cmdCompletion(args []string) {
	shell := "bash"
	if len(args) > 0 {
		shell = args[0]
	}

	switch shell {
	case "bash":
		fmt.Print(completionBash)
	case "zsh":
		fmt.Print(completionZsh)
	case "fish":
		fmt.Print(completionFish)
	default:
		fmt.Fprintf(os.Stderr, "unknown shell %q. Use: bash, zsh, or fish.\n", shell)
		os.Exit(1)
	}
}
