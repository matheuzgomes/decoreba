package cli

import (
	"fmt"
	"sort"
	"strings"

	"decoreba/internal/core"
	"decoreba/internal/core/store"
)

func cmdList(args []string) {
	s, err := store.Load()
	check(err)
	if len(args) == 0 {
		printContexts(s)
		return
	}
	context := args[0]
	var found []core.Command
	for _, c := range s.Commands {
		if strings.EqualFold(c.Context, context) {
			found = append(found, c)
		}
	}
	if len(found) == 0 {
		fmt.Printf("No commands in context %q.\n", context)
		return
	}
	sort.Slice(found, func(i, j int) bool { return found[i].UsageCount > found[j].UsageCount })
	for _, c := range found {
		fmt.Printf("[%s] %s\n     %s\n", c.ID, c.Title, c.Command)
	}
}

func printContexts(s *core.Store) {
	counts := map[string]int{}
	for _, c := range s.Commands {
		counts[c.Context]++
	}
	if len(counts) == 0 {
		fmt.Println("No commands saved yet. Use 'decoreba add' to start.")
		return
	}
	var names []string
	for name := range counts {
		names = append(names, name)
	}
	sort.Strings(names)
	fmt.Println("Available contexts:")
	for _, name := range names {
		fmt.Printf("  - %s (%d)\n", name, counts[name])
	}
}
