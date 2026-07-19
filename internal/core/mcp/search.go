package mcp

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/search"
)

func handleSearch(s *core.Store, id json.RawMessage, args json.RawMessage) {
	var params struct {
		Query   string `json:"query"`
		Context string `json:"context,omitempty"`
		Limit   int    `json:"limit,omitempty"`
	}
	if args != nil {
		json.Unmarshal(args, &params)
	}
	if params.Query == "" {
		writeError(id, -32602, "query is required")
		return
	}
	if params.Limit <= 0 || params.Limit > 50 {
		params.Limit = 10
	}

	pool := s.Commands
	if params.Context != "" {
		var filtered []core.Command
		for _, c := range pool {
			if strings.EqualFold(c.Context, params.Context) {
				filtered = append(filtered, c)
			}
		}
		pool = filtered
	}

	results := search.Sort(pool, params.Query)
	if len(results) > params.Limit {
		results = results[:params.Limit]
	}

	var lines []string
	for i, r := range results {
		cmd := r.Cmd
		ns := ""
		if r.Cmd.IsWorkflow() {
			ns = fmt.Sprintf(" (%d steps)", len(cmd.Steps))
		}
		lines = append(lines, fmt.Sprintf("%d. [%s] (%s) %s%s",
			i+1, cmd.ID, cmd.Context, cmd.Title, ns))
		lines = append(lines, fmt.Sprintf("   %s", cmd.Command))
		if len(cmd.Tags) > 0 {
			lines = append(lines, fmt.Sprintf("   tags: %s", strings.Join(cmd.Tags, ", ")))
		}
		if r.Score > 0 {
			lines = append(lines, fmt.Sprintf("   score: %d", r.Score))
		}
		lines = append(lines, "")
	}

	if len(lines) == 0 {
		writeResult(id, map[string]interface{}{
			"content": textContent("No commands found."),
		})
		return
	}

	writeResult(id, map[string]interface{}{
		"content": textContent(strings.Join(lines, "\n")),
	})
}

func handleGet(s *core.Store, id json.RawMessage, args json.RawMessage) {
	var params struct {
		ID string `json:"id"`
	}
	if args != nil {
		json.Unmarshal(args, &params)
	}
	if params.ID == "" {
		writeError(id, -32602, "id is required")
		return
	}

	cmd, count := s.FindByPrefix(params.ID)
	if count == 0 {
		writeResult(id, map[string]interface{}{
			"content": textContent("No command found with that id."),
		})
		return
	}
	if count > 1 {
		writeResult(id, map[string]interface{}{
			"content": textContent(fmt.Sprintf("Ambiguous id %q (%d matches). Use more characters.", params.ID, count)),
		})
		return
	}

	b, _ := json.MarshalIndent(cmd, "", "  ")
	writeResult(id, map[string]interface{}{
		"content": textContent(string(b)),
	})
}

func handleListContexts(s *core.Store, id json.RawMessage, args json.RawMessage) {
	counts := map[string]int{}
	for _, c := range s.Commands {
		counts[c.Context]++
	}

	if len(counts) == 0 {
		writeResult(id, map[string]interface{}{
			"content": textContent("No commands saved yet."),
		})
		return
	}

	var lines []string
	lines = append(lines, "Contexts:")
	for name, n := range counts {
		label := "commands"
		if n == 1 {
			label = "command"
		}
		lines = append(lines, fmt.Sprintf("  - %s (%d %s)", name, n, label))
	}
	writeResult(id, map[string]interface{}{
		"content": textContent(strings.Join(lines, "\n")),
	})
}
