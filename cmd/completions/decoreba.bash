# bash completion for decoreba
# Install: source <(decoreba completion bash)
#   or copy to /usr/share/bash-completion/completions/decoreba

_decoreba_completion() {
    local cur prev words cword
    _init_completion || return

    # Top-level after "decoreba": offer subcommands and contexts.
    if ((cword == 1)); then
        COMPREPLY=($(compgen -W "add list ls rm remove edit stats version help completion" -- "$cur"))
        return
    fi

    case "${words[1]}" in
        list|ls)
            COMPREPLY=($(compgen -W "$(decoreba list 2>/dev/null | sed -n 's/^  - \([^ ]*\).*/\1/p')" -- "$cur"))
            ;;
        rm|remove|edit)
            # Complete with command IDs from that context if a second arg exists.
            COMPREPLY=()
            ;;
        completion)
            COMPREPLY=($(compgen -W "bash zsh fish" -- "$cur"))
            ;;
        add|stats|version|help)
            COMPREPLY=()
            ;;
        *)
            # Default: subcommand already chosen, nothing more to complete.
            COMPREPLY=()
            ;;
    esac
}

complete -F _decoreba_completion decoreba
