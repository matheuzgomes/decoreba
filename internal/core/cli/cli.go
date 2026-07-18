package cli

import (
	"os"
	"strings"
)

const version string = "0.1.0"

func Run() {
	args := os.Args[1:]

	// Respect the NO_COLOR convention and --no-color flag.
	if os.Getenv("NO_COLOR") != "" {
		noColor = true
	}
	filtered := args[:0]
	for _, a := range args {
		if a == "--no-color" {
			noColor = true
			continue
		}
		if a == "--logo" {
			showLogo()
			return
		}
		filtered = append(filtered, a)
	}
	args = filtered

	// First-run: show the animated logo + welcome message.
	if isFirstRun() {
		showFirstRunWelcome()
	}

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
	case "edit":
		cmdEdit(args[1:])
	case "stats":
		cmdStats()
	case "completion":
		cmdCompletion(args[1:])
	case "version", "-v", "--version":
		showVersionBox()
	case "help", "-h", "--help":
		printHelp()
	default:
		query := strings.Join(args[1:], " ")
		runSearch(args[0], query)
	}
}
