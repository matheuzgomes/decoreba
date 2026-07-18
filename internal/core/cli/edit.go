package cli

import (
	"fmt"
	"strings"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
	"github.com/matheuzgomes/decoreba/internal/core/term"
	"github.com/matheuzgomes/decoreba/internal/core/tui"
)

func cmdEdit(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: decoreba edit <id>")
		return
	}
	idPrefix := args[0]
	s, err := store.Load()
	check(err)

	cmd, matchCount := findCommandByPrefix(s, idPrefix)
	if matchCount == 0 {
		fmt.Println("No command found with that id.")
		return
	}
	if matchCount > 1 {
		fmt.Println("Ambiguous id, use more characters.")
		return
	}

	if !term.IsTerminal() {
		path, _ := store.ConfigPath()
		fmt.Printf("Terminal required for editing. Edit the file directly:\n  %s\n", path)
		return
	}

	edited, err := tui.RunEditForm(s, cmd)
	check(err)
	if edited == nil {
		return
	}

	replaceCommand(s, edited)
	check(store.Save(s))
	fmt.Printf("✓ Command updated in %q (id: %s)\n", edited.Context, edited.ID)
}

// findCommandByPrefix returns the unique command matching the id prefix and
// the total number of matches.
func findCommandByPrefix(s *core.Store, prefix string) (*core.Command, int) {
	count := 0
	var found *core.Command
	for i := range s.Commands {
		if strings.HasPrefix(s.Commands[i].ID, prefix) {
			found = &s.Commands[i]
			count++
		}
	}
	return found, count
}

// replaceCommand replaces a command in the store (matched by ID) with the
// edited version.
func replaceCommand(s *core.Store, updated *core.Command) {
	for i := range s.Commands {
		if s.Commands[i].ID == updated.ID {
			s.Commands[i] = *updated
			return
		}
	}
}
