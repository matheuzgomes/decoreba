package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

func cmdImport(args []string) {
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
		s, err := store.Load()
		check(err)
		imported, skipped := mergeCommands(s, full)
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

	s, err := store.Load()
	check(err)

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
			CreatedAt: now,
			UpdatedAt: now,
		}
	}
	imported, skipped := mergeCommands(s, cmds)
	check(store.Save(s))
	fmt.Printf("Imported %d commands, skipped %d (already exist)\n", imported, skipped)
}

// mergeCommands appends commands to the store, skipping duplicates (same
// context+command, case-insensitive). Returns counts of imported and skipped.
func mergeCommands(s *core.Store, incoming []core.Command) (imported, skipped int) {
	seen := make(map[string]bool, len(s.Commands))
	for _, c := range s.Commands {
		seen[key(c)] = true
	}
	for _, c := range incoming {
		if seen[key(c)] {
			skipped++
			continue
		}
		s.Commands = append(s.Commands, c)
		seen[key(c)] = true
		imported++
	}
	return
}

func key(c core.Command) string {
	return strings.ToLower(c.Context) + "\x00" + strings.ToLower(c.Command)
}
