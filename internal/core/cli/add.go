package cli

import (
	"fmt"
	"strings"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
	"github.com/matheuzgomes/decoreba/internal/core/term"
	"github.com/matheuzgomes/decoreba/internal/core/tui"
)

func cmdAdd(s *core.Store) {
	if !term.IsTerminal() {
		cmdAddFallback(s)
		return
	}

	cmd, err := tui.RunAddForm(s)
	check(err)
	if cmd == nil {
		return
	}
	s.Commands = append(s.Commands, *cmd)
	check(store.Save(s))
	fmt.Printf("✓ Command saved in %q (id: %s)\n", cmd.Context, cmd.ID)
}

func cmdAddFallback(s *core.Store) {
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
