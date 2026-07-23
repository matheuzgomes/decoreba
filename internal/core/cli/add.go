package cli

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/history"
	"github.com/matheuzgomes/decoreba/internal/core/store"
	"github.com/matheuzgomes/decoreba/internal/core/term"
	"github.com/matheuzgomes/decoreba/internal/core/tui"
)

func cmdAdd(s *core.Store, args ...string) {
	fromHistory := false
	for _, arg := range args {
		if arg == "--last" {
			fromHistory = true
			continue
		}
		fmt.Fprintf(os.Stderr, "Unknown add option %q.\n", arg)
		return
	}

	initialCommand := ""
	if fromHistory {
		var err error
		initialCommand, err = history.LastExcludingSelf()
		if err != nil {
			fmt.Fprintf(os.Stderr, "Unable to load the last shell command: %v\n", err)
			return
		}
	}

	if !term.IsTerminal() {
		cmdAddFallback(s, initialCommand)
		return
	}

	var cmd *core.Command
	var err error
	if initialCommand == "" {
		cmd, err = tui.RunAddForm(s)
	} else {
		cmd, err = tui.RunAddFormWithCommand(s, initialCommand)
	}
	check(err)
	if cmd == nil {
		return
	}
	s.Commands = append(s.Commands, *cmd)
	check(store.Save(s))
	fmt.Printf("✓ Command saved in %q (id: %s)\n", cmd.Context, cmd.ID)
}

func cmdAddFallback(s *core.Store, initialCommand ...string) {
	context := promptLine("Context (ex: tmux, git, docker): ")
	title := promptLine("Short title: ")
	command := ""
	if len(initialCommand) > 0 {
		command = initialCommand[0]
		fmt.Printf("Command (from history): %s\n", command)
	} else {
		command = promptLine("Command: ")
	}
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
	cmd := core.Command{
		ID:        core.GenID(),
		Context:   strings.ToLower(context),
		Title:     title,
		Command:   command,
		Tags:      tags,
		Notes:     notes,
		CreatedAt: now,
		UpdatedAt: now,
	}
	s.Commands = append(s.Commands, cmd)
	check(store.Save(s))
	fmt.Printf("✓ Command saved in %q (id: %s)\n", cmd.Context, cmd.ID)
}
