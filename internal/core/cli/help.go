package cli

import "fmt"

func printHelp() {
	fmt.Print(`decoreba - your personal command vault, organized by context

Usage:
  decoreba                      Interactive search across all contexts
  decoreba <context>            Interactive search within a context (e.g. tmux)
  decoreba <context> <query>    Direct search (e.g. decoreba git undo)
  decoreba add                  Add a new command (interactive mode)
  decoreba list                 List contexts and command count
  decoreba list <context>       List commands saved in a context
  decoreba rm <id>              Remove a command by id (or id prefix)
  decoreba edit <id>            Edit a command by id (or id prefix)
  decoreba stats                Show vault statistics
  decoreba completion <shell>   Generate shell completion (bash|zsh|fish)
  decoreba version              Show version
  decoreba help                 Show this help

Commands are saved to: $XDG_CONFIG_HOME/decoreba/commands.json (or OS equivalent)
`)
}
