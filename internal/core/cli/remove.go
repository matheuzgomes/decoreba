package cli

import (
	"fmt"
	"strings"

	"decoreba/internal/core/store"
)

func cmdRemove(args []string) {
	if len(args) == 0 {
		fmt.Println("Usage: decoreba rm <id>")
		return
	}
	idPrefix := args[0]
	s, err := store.Load()
	check(err)

	matchIdx := -1
	matchCount := 0
	for i, c := range s.Commands {
		if strings.HasPrefix(c.ID, idPrefix) {
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
