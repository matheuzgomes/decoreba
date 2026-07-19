package cli

import (
	"os"
	"strings"
)

const version string = "0.1.0"

// shellOutput is set when --shell-output is passed. In this mode the final
// result is printed to stdout (with an optional EXEC: prefix for execute
// actions) instead of going to the clipboard.
var shellOutput bool

func Run() {
	args := os.Args[1:]

	if os.Getenv("NO_COLOR") != "" {
		noColor = true
	}
	filtered := args[:0]
	for _, a := range args {
		if a == "--no-color" {
			noColor = true
			continue
		}
		if a == "--shell-output" {
			shellOutput = true
			continue
		}
		if a == "--logo" {
			showLogo()
			return
		}
		filtered = append(filtered, a)
	}
	args = filtered

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
	case "shell":
		cmdShell(args[1:])
	case "export":
		cmdExport(args[1:])
	case "import":
		cmdImport(args[1:])
	case "version", "-v", "--version":
		showVersionBox()
	case "help", "-h", "--help":
		printHelp()
	default:
		if shellOutput {
			runSearch("", strings.Join(args, " "))
		} else {
			query := strings.Join(args[1:], " ")
			runSearch(args[0], query)
		}
	}
}
