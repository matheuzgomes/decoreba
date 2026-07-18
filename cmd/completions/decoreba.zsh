#compdef decoreba
# zsh completion for decoreba
# Install: source <(decoreba completion zsh)
#   or copy to a directory in $fpath as _decoreba

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
        'version:Show version'
        'help:Show help'
        'completion:Generate shell completion'
    )

    local context
    _arguments -C \
        '1: :->first' \
        '*:: :->rest'

    case "$state" in
        first)
            _describe -t commands 'subcommand or context' subcmds
            ;;
        rest)
            case "${words[2]}" in
                list|ls)
                    # Complete with context names.
                    local -a contexts
                    contexts=(${(f)"$(decoreba list 2>/dev/null | sed -n 's/^  - \([^ ]*\).*/\1/p')"})
                    _describe -t contexts 'context' contexts
                    ;;
                rm|remove|edit)
                    _message 'command id'
                    ;;
                completion)
                    _values 'shell' bash zsh fish
                    ;;
            esac
            ;;
    esac
}

_decoreba "$@"
