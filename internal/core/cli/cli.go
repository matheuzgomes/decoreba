package cli

import (
	"fmt"
	"os"
	"strings"
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
