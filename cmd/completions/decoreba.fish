# fish completion for decoreba
# Install: decoreba completion fish > ~/.config/fish/completions/decoreba.fish

# Subcommands (no file completion)
complete -c decoreba -f

complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats version help completion' \
    -a 'add' -d 'Add a new command'

complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats version help completion' \
    -a 'list' -d 'List contexts and commands'

complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats version help completion' \
    -a 'ls' -d 'List contexts and commands (alias)'

complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats version help completion' \
    -a 'rm' -d 'Remove a command by id'

complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats version help completion' \
    -a 'remove' -d 'Remove a command by id (alias)'

complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats version help completion' \
    -a 'edit' -d 'Edit a command by id'

complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats version help completion' \
    -a 'stats' -d 'Show vault statistics'

complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats version help completion' \
    -a 'version' -d 'Show version'

complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats version help completion' \
    -a 'help' -d 'Show help'

complete -c decoreba -n 'not __fish_seen_subcommand_from add list ls rm remove edit stats version help completion' \
    -a 'completion' -d 'Generate shell completion'

# Context name completion for 'list <context>'
complete -c decoreba -n '__fish_seen_subcommand_from list ls; and test (count (commandline -opc)) -eq 2' \
    -a '(decoreba list 2>/dev/null | string match -r "  - \w+" | string replace -r "  - " "")'
