package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

func cmdImport(s *core.Store, args []string) {
	input := ""
	if len(args) > 0 {
		input = args[0]
	}

	var data []byte
	var err error
	if input == "" {
		data, err = os.ReadFile("/dev/stdin")
	} else {
		data, err = os.ReadFile(input)
	}
	check(err)

	// Try full format first ([]core.Command).
	var full []core.Command
	if err := json.Unmarshal(data, &full); err == nil {
		imported, skipped := s.Merge(full)
		check(store.Save(s))
		fmt.Printf("Imported %d commands, skipped %d (already exist)\n", imported, skipped)
		return
	}

	// Try simplified format.
	var clean []exportCmd
	if err := json.Unmarshal(data, &clean); err != nil {
		fmt.Fprintln(os.Stderr, "Invalid format: expected a JSON array of commands.")
		os.Exit(1)
	}

	now := time.Now()
	cmds := make([]core.Command, len(clean))
	for i, c := range clean {
		cmds[i] = core.Command{
			ID:        core.GenID(),
			Context:   c.Context,
			Title:     c.Title,
			Command:   c.Command,
			Tags:      c.Tags,
			Notes:     c.Notes,
			Pinned:    c.Pinned,
			Steps:     c.Steps,
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	imported, skipped := s.Merge(cmds)
	check(store.Save(s))
	fmt.Printf("Imported %d commands, skipped %d (already exist)\n", imported, skipped)
}
