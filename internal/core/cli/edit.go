package cli

import (
	"fmt"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
	"github.com/matheuzgomes/decoreba/internal/core/term"
	"github.com/matheuzgomes/decoreba/internal/core/tui"
)

func cmdEdit(s *core.Store, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: decoreba edit <id>")
		return
	}
	idPrefix := args[0]

	cmd, matchCount := s.FindByPrefix(idPrefix)
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
