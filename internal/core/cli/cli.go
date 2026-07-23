package cli

import (
	"os"
	"strings"

	"github.com/matheuzgomes/decoreba/internal/core/store"
	"github.com/matheuzgomes/decoreba/internal/core/tui"
)

const version string = "0.3.0"

func Run() {
	args := os.Args[1:]

	noColor := os.Getenv("NO_COLOR") != ""
	shellOutput := false
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

	if noColor {
		tui.SetNoColor(true)
	}

	if isFirstRun() {
		showFirstRunWelcome()
	}

	s, err := store.Load()
	check(err)

	if len(args) == 0 {
		runSearch(s, "", "", shellOutput)
		return
	}

	switch args[0] {
	case "add":
		cmdAdd(s, args[1:]...)
	case "list", "ls":
		cmdList(s, args[1:], shellOutput)
	case "rm", "remove":
		cmdRemove(s, args[1:])
	case "edit":
		cmdEdit(s, args[1:])
	case "stats":
		cmdStats(s)
	case "init":
		cmdInit(args[1:])
	case "completion":
		cmdCompletion(args[1:])
	case "shell":
		cmdShell(args[1:])
	case "mcp":
		cmdMCP()
	case "sync":
		cmdSync(args[1:])
	case "export":
		cmdExport(s, args[1:])
	case "import":
		cmdImport(s, args[1:])
	case "version", "-v", "--version":
		showVersionBox()
	case "help", "-h", "--help":
		printHelp()
	default:
		if shellOutput {
			runSearch(s, "", strings.Join(args, " "), shellOutput)
		} else {
			query := strings.Join(args[1:], " ")
			runSearch(s, args[0], query, shellOutput)
		}
	}
}
