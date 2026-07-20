package cli

import (
	"fmt"
	"os"
	"strconv"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/search"
	"github.com/matheuzgomes/decoreba/internal/core/store"
	"github.com/matheuzgomes/decoreba/internal/core/term"
	"github.com/matheuzgomes/decoreba/internal/core/tui"
)

func runSearch(s *core.Store, context, query string, shellOutput bool) {
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

	chosen, action, err := tui.RunPalette(s, context, query, func(cmd *core.Command) {
		_ = store.Save(s)
	})
	check(err)
	if chosen == nil {
		return
	}

	handleActionResult(s, chosen, action, shellOutput)
}

func runSearchFallback(s *core.Store, context, query string) {
	pool := s.FilterByContext(context)
	if context != "" && len(pool) == 0 {
		fmt.Printf("No commands in context %q yet.\n\n", context)
		printContexts(s)
		return
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
