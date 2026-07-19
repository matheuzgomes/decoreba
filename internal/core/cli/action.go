package cli

import (
	"fmt"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
	"github.com/matheuzgomes/decoreba/internal/core/tui"
)

// handleActionResult processes the command returned by the TUI palette or
// list browser: resolves variables, runs workflows, and dispatches to
// copy / edit / execute.
func handleActionResult(s *core.Store, chosen *core.Command, action tui.PaletteAction, shellOutput bool) {
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
