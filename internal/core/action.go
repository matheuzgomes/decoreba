package core

import (
	"fmt"
	"os"
	"sort"
	"strings"
	"time"
)

const version string = "0.1.0"

func Run() {
	args := os.Args[1:]
	if len(args) == 0 {
		runSearch("", "")
		return
	}

	switch args[0] {
	case "add":
		cmdAdd()
	case "list", "ls":
		cmdList(args[1:])
	case "rm", "remove":
		cmdRemove(args[1:])
	case "version", "-v", "--version":
		fmt.Println("decoreba " + version)
	case "help", "-h", "--help":
		printHelp()
	default:
		query := strings.Join(args[1:], " ")
		runSearch(args[0], query)
	}
}

func cmdAdd() {
	store, err := Load()
	check(err)

	context := promptLine("Context (ex: tmux, git, docker): ")
	title := promptLine("Short title: ")
	command := promptLine("Command: ")
	tagsRaw := promptLine("Tags (comma separated, optional): ")
	notes := promptLine("Notes (optional): ")

	if context == "" || command == "" {
		fmt.Println("Context and command are required. Nothing was saved.")
		return
	}

	var tags []string
	for _, t := range strings.Split(tagsRaw, ",") {
		t = strings.TrimSpace(t)
		if t != "" {
			tags = append(tags, t)
		}
	}

	now := time.Now()
	cmd := Command{
		ID:        genID(),
		Context:   strings.ToLower(context),
		Title:     title,
		Command:   command,
		Tags:      tags,
		Notes:     notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
	store.Commands = append(store.Commands, cmd)
	check(Save(store))
	fmt.Printf("✓ Command saved in %q (id: %s)\n", cmd.Context, cmd.ID)
}

func cmdList(args []string) {
	store, err := Load()
	check(err)
	if len(args) == 0 {
		printContexts(store)
		return
	}
	context := args[0]
	var found []Command
	for _, c := range store.Commands {
		if strings.EqualFold(c.Context, context) {
			found = append(found, c)
		}
	}
	if len(found) == 0 {
		fmt.Printf("No commands in context %q.\n", context)
		return
	}
	sort.Slice(found, func(i, j int) bool { return found[i].UsageCount > found[j].UsageCount })
	for _, c := range found {
		fmt.Printf("[%s] %s\n     %s\n", c.ID, c.Title, c.Command)
	}
}

func cmdRemove(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: decoreba rm <id>")
		return
	}
	idPrefix := args[0]
	store, err := Load()
	check(err)

	matchIdx := -1
	matchCount := 0
	for i, c := range store.Commands {
		if strings.HasPrefix(c.ID, idPrefix) {
			matchIdx = i
			matchCount++
		}
	}
	if matchCount == 0 {
		fmt.Println("No command found with that id.")
		return
	}
	if matchCount > 1 {
		fmt.Println("Ambiguous id, use more characters.")
		return
	}
	removed := store.Commands[matchIdx]
	store.Commands = append(store.Commands[:matchIdx], store.Commands[matchIdx+1:]...)
	check(Save(store))
	fmt.Printf("✓ Removed: %s (%s)\n", removed.Title, removed.Context)
}

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
  decoreba version              Show version
  decoreba help                 Show this help

Commands are saved to: $XDG_CONFIG_HOME/decoreba/commands.json (or OS equivalent)
`)
}
