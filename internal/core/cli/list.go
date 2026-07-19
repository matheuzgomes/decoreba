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

func cmdList(args []string) {
	s, err := store.Load()
	check(err)

	if len(args) == 0 {
		if term.IsTerminal() {
			chosen, action, err := tui.RunListBrowser(s)
			check(err)
			if chosen == nil {
				return
			}
			handleListResult(s, chosen, action)
		} else {
			printContexts(s)
		}
		return
	}

	context := args[0]
	if term.IsTerminal() {
		chosen, action, err := tui.RunPalette(s, context, "")
		check(err)
		if chosen == nil {
			return
		}
		handleListResult(s, chosen, action)
	} else {
		printContextCommands(s, context)
	}
}

func handleListResult(s *core.Store, chosen *core.Command, action tui.PaletteAction) {
	if chosen.IsWorkflow() {
		bumpUsage(s, chosen)
		_ = tui.RunWorkflow(chosen)
		return
	}

	if tui.HasVariables(chosen.Command) && action == tui.ActionCopy {
		resolved, cancelled, err := tui.ResolveCommandInteractive(chosen.Command)
		check(err)
		if cancelled {
			return
		}
		chosen.Command = resolved
	}

	switch action {
	case tui.ActionEdit:
		edited, err := tui.RunEditForm(s, chosen)
		check(err)
		if edited == nil {
			return
		}
		replaceCommand(s, edited)
		check(store.Save(s))
		fmt.Printf("✓ Command updated in %q (id: %s)\n", edited.Context, edited.ID)
	case tui.ActionExecute:
		bumpUsage(s, chosen)
		if shellOutput {
			fmt.Print("EXEC:" + chosen.Command)
		} else {
			_ = tui.RunCommand(chosen)
		}
	default:
		bumpUsage(s, chosen)
		if shellOutput {
			fmt.Print(chosen.Command)
		} else {
			confirmCopy(s, chosen)
		}
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


