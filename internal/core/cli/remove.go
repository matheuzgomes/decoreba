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

	removed, count := s.RemoveByPrefix(idPrefix)
	switch {
	case count == 0:
		fmt.Println("No command found with that id.")
		return
	case count > 1:
		fmt.Println("Ambiguous id, use more characters.")
		return
	}
	check(store.Save(s))
	fmt.Printf("✓ Removed: %s (%s)\n", removed.Title, removed.Context)
}
