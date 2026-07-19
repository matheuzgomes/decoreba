package cli

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/clipboard"
	"github.com/matheuzgomes/decoreba/internal/core/search"
	"github.com/matheuzgomes/decoreba/internal/core/store"
	"github.com/matheuzgomes/decoreba/internal/core/term"
	"github.com/matheuzgomes/decoreba/internal/core/tui"
)

func runSearch(context, query string) {
	s, err := store.Load()
	check(err)

	// Auto-detect context when none was specified.
	if context == "" {
		context = detectContext(s)
	}

	if !term.IsTerminal() {
		if shellOutput {
			// stdout is captured by $(), but we can still use /dev/tty
			// for the interactive palette.
			tui.UseTTY = true
			// Re-open stdin as /dev/tty so raw mode works.
			if tty, err := os.OpenFile("/dev/tty", os.O_RDWR, 0); err == nil {
				oldStdin := os.Stdin
				os.Stdin = tty
				defer func() { os.Stdin = oldStdin; tty.Close() }()
			}
		} else {
			runSearchFallback(s, context, query)
			return
		}
	}

	chosen, action, err := tui.RunPalette(s, context, query)
	check(err)
	if chosen == nil {
		return
	}

	// Workflows run their steps interactively.
	if chosen.IsWorkflow() {
		bumpUsage(s, chosen)
		_ = tui.RunWorkflow(chosen)
		return
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
			confirmCopy(s, chosen)
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

func bumpUsage(s *core.Store, chosen *core.Command) {
	for i := range s.Commands {
		if s.Commands[i].ID == chosen.ID {
			s.Commands[i].UsageCount++
			s.Commands[i].LastUsedAt = time.Now()
			break
		}
	}
	_ = store.Save(s)
}

func confirmCopy(s *core.Store, chosen *core.Command) {
	if err := clipboard.Copy(chosen.Command); err != nil {
		fmt.Printf("Could not copy automatically (%v).\nCommand: %s\n", err, chosen.Command)
	} else {
		fmt.Printf("✓ Copied: %s\n", chosen.Command)
	}
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
		label := "› "
		if context != "" {
			label = context + " › "
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
