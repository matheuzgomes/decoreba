package cli

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/importer"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

func cmdImport(s *core.Store, args []string) {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "Usage: decoreba import [pet|navi|<file>]")
		os.Exit(1)
	}

	switch args[0] {
	case "pet":
		importPet(s, args[1:])
	case "navi":
		importNavi(s, args[1:])
	default:
		importJSON(s, args)
	}
}

func importPet(s *core.Store, args []string) {
	path, dryRun := parseImportFlags(args)
	if path == "" {
		fmt.Fprintln(os.Stderr, "Usage: decoreba import pet --path <file> [--dry-run]")
		os.Exit(1)
	}

	report, err := importer.ImportPet(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	finishImport(s, report, dryRun)
}

func importNavi(s *core.Store, args []string) {
	path, dryRun := parseImportFlags(args)
	if path == "" {
		fmt.Fprintln(os.Stderr, "Usage: decoreba import navi --path <path> [--dry-run]")
		os.Exit(1)
	}

	report, err := importer.ImportNavi(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %v\n", err)
		os.Exit(1)
	}

	finishImport(s, report, dryRun)
}

func parseImportFlags(args []string) (path string, dryRun bool) {
	for i := 0; i < len(args); i++ {
		switch args[i] {
		case "--path":
			if i+1 < len(args) {
				path = args[i+1]
				i++
			}
		case "--dry-run":
			dryRun = true
		}
	}
	return
}

func finishImport(s *core.Store, report *importer.Report, dryRun bool) {
	if len(report.Commands) == 0 && !report.HasWarnings() {
		fmt.Println("No commands to import.")
		return
	}

	resolveContexts(s, report.Commands)

	report.DryRun = dryRun
	imported, skipped := s.Merge(report.Commands)
	report.Imported = imported
	report.Skipped = skipped

	if !dryRun && imported > 0 {
		check(store.Save(s))
	}

	fmt.Print(report.String())
	if len(report.Commands) == 0 && report.HasWarnings() {
		os.Exit(1)
	}
}

func resolveContexts(s *core.Store, cmds []core.Command) {
	existing := make(map[string]string)
	for _, c := range s.Commands {
		lower := strings.ToLower(c.Context)
		if _, ok := existing[lower]; !ok {
			existing[lower] = c.Context
		}
	}
	for i := range cmds {
		lower := strings.ToLower(cmds[i].Context)
		if match, ok := existing[lower]; ok {
			cmds[i].Context = match
		}
	}
}

func importJSON(s *core.Store, args []string) {
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

	var full []core.Command
	if err := json.Unmarshal(data, &full); err == nil {
		imported, skipped := s.Merge(full)
		check(store.Save(s))
		fmt.Printf("Imported %d commands, skipped %d (already exist)\n", imported, skipped)
		return
	}

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
