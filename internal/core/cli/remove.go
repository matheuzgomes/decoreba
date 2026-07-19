package cli

import (
	"fmt"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

func cmdRemove(s *core.Store, args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: decoreba rm <id>")
		return
	}
	idPrefix := args[0]

	matchIdx := -1
	matchCount := 0
	// linear scan for the index (FindByPrefix returns pointer, but we need index)
	for i, c := range s.Commands {
		if len(c.ID) >= len(idPrefix) && c.ID[:len(idPrefix)] == idPrefix {
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
	removed := s.Commands[matchIdx]
	s.Commands = append(s.Commands[:matchIdx], s.Commands[matchIdx+1:]...)
	check(store.Save(s))
	fmt.Printf("✓ Removed: %s (%s)\n", removed.Title, removed.Context)
}
