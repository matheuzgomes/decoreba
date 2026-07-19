package cli

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func cmdExport(s *core.Store, args []string) {
	full := false
	output := ""
	for _, a := range args {
		if a == "--full" {
			full = true
		} else {
			output = a
		}
	}

	var data []byte
	var err error
	if full {
		data, err = json.MarshalIndent(s.Commands, "", "  ")
	} else {
		clean := make([]exportCmd, len(s.Commands))
		for i, c := range s.Commands {
			clean[i] = exportCmd{
				Context: c.Context,
				Title:   c.Title,
				Command: c.Command,
				Tags:    c.Tags,
				Notes:   c.Notes,
				Pinned:  c.Pinned,
				Steps:   c.Steps,
			}
		}
		data, err = json.MarshalIndent(clean, "", "  ")
	}
	check(err)

	if output == "" {
		fmt.Println(string(data))
		return
	}
	check(os.WriteFile(output, data, 0o600))
	fmt.Printf("Exported %d commands to %s\n", len(s.Commands), output)
}

type exportCmd struct {
	Context string             `json:"context"`
	Title   string             `json:"title"`
	Command string             `json:"command"`
	Tags    []string           `json:"tags,omitempty"`
	Notes   string             `json:"notes,omitempty"`
	Pinned  bool               `json:"pinned,omitempty"`
	Steps   []core.WorkflowStep `json:"steps,omitempty"`
}
