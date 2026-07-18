package cli

import (
	"fmt"
	"strconv"
	"strings"

	"decoreba/internal/core"
	"decoreba/internal/core/clipboard"
	"decoreba/internal/core/search"
	"decoreba/internal/core/store"
	"decoreba/internal/core/term"
	"decoreba/internal/core/tui"
)

func runSearch(context, query string) {
	s, err := store.Load()
	check(err)

	if !term.IsTerminal() {
		runSearchFallback(s, context, query)
		return
	}

	chosen, err := tui.RunPalette(s, context, query)
	check(err)
	if chosen == nil {
		return
	}
	confirmCopy(s, chosen)
}

func confirmCopy(s *core.Store, chosen *core.Command) {
	if err := clipboard.Copy(chosen.Command); err != nil {
		fmt.Printf("Could not copy automatically (%v).\nCommand: %s\n", err, chosen.Command)
	} else {
		fmt.Printf("✓ Copied: %s\n", chosen.Command)
	}

	for i := range s.Commands {
		if s.Commands[i].ID == chosen.ID {
			s.Commands[i].UsageCount++
			break
		}
	}
	_ = store.Save(s)
}

func runSearchFallback(s *core.Store, context, query string) {
	var pool []core.Command
	if context != "" {
		for _, c := range s.Commands {
			if strings.EqualFold(c.Context, context) {
				pool = append(pool, c)
			}
		}
		if len(pool) == 0 {
			fmt.Printf("No commands in context %q yet.\n\n", context)
			printContexts(s)
			return
		}
	} else {
		pool = s.Commands
	}

	if query == "" {
		label := "Search> "
		if context != "" {
			label = fmt.Sprintf("Search in %s> ", context)
		}
		query = promptLine(label)
	}

	results := search.Sort(pool, query)
	if len(results) == 0 {
		fmt.Println("No commands found.")
		return
	}

	cmdList := make([]core.Command, len(results))
	for i, r := range results {
		cmdList[i] = r.Cmd
	}

	chosen, err := fallbackSelector(cmdList)
	if err != nil || chosen == nil {
		return
	}
	confirmCopy(s, chosen)
}

func fallbackSelector(cmds []core.Command) (*core.Command, error) {
	fmt.Println()
	for i, c := range cmds {
		fmt.Printf("[%d] (%s) %s\n     %s\n", i+1, c.Context, c.Title, c.Command)
	}
	fmt.Println()

	choice := promptLine("Copy which? (number, ENTER cancels): ")
	if choice == "" {
		return nil, nil
	}
	idx, err := strconv.Atoi(choice)
	if err != nil || idx < 1 || idx > len(cmds) {
		fmt.Println("Invalid choice.")
		return nil, nil
	}
	return &cmds[idx-1], nil
}
