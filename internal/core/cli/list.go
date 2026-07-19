package cli

import (
	"fmt"
	"sort"
	"strings"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
	"github.com/matheuzgomes/decoreba/internal/core/term"
	"github.com/matheuzgomes/decoreba/internal/core/tui"
)

func cmdList(s *core.Store, args []string, shellOutput bool) {
	onPin := func(cmd *core.Command) {
		_ = store.Save(s)
	}

	if len(args) == 0 {
		if term.IsTerminal() {
			chosen, action, err := tui.RunListBrowser(s, onPin)
			check(err)
			if chosen == nil {
				return
			}
			handleActionResult(s, chosen, action, shellOutput)
		} else {
			printContexts(s)
		}
		return
	}

	context := args[0]
	if term.IsTerminal() {
		chosen, action, err := tui.RunPalette(s, context, "", onPin)
		check(err)
		if chosen == nil {
			return
		}
		handleActionResult(s, chosen, action, shellOutput)
	} else {
		printContextCommands(s, context)
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

func printContextCommands(s *core.Store, context string) {
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


