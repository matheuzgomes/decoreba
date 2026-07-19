package mcp

import (
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
)

func handleStats(s *core.Store, id json.RawMessage, args json.RawMessage) {
	total := len(s.Commands)
	if total == 0 {
		writeResult(id, map[string]interface{}{
			"content": textContent("Vault is empty. Use decoreba_add to save commands."),
		})
		return
	}

	ctxCounts := map[string]int{}
	tagCounts := map[string]int{}
	var mostUsed []core.Command
	var recent []core.Command

	for _, c := range s.Commands {
		ctxCounts[c.Context]++
		for _, t := range c.Tags {
			tagCounts[t]++
		}
	}

	sorted := append([]core.Command(nil), s.Commands...)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].UsageCount > sorted[j].UsageCount
	})
	if len(sorted) > 5 {
		mostUsed = sorted[:5]
	} else {
		mostUsed = sorted
	}

	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].LastUsedAt.After(sorted[j].LastUsedAt)
	})
	for _, c := range sorted {
		if !c.LastUsedAt.IsZero() && len(recent) < 5 {
			recent = append(recent, c)
		}
	}

	var lines []string
	lines = append(lines, fmt.Sprintf("Total commands: %d", total))

	wfCount := 0
	for _, c := range s.Commands {
		if c.IsWorkflow() {
			wfCount++
		}
	}
	if wfCount > 0 {
		lines = append(lines, fmt.Sprintf("Workflows:     %d", wfCount))
	}

	pinnedCount := 0
	for _, c := range s.Commands {
		if c.Pinned {
			pinnedCount++
		}
	}
	if pinnedCount > 0 {
		lines = append(lines, fmt.Sprintf("Pinned:        %d", pinnedCount))
	}

	lines = append(lines, "")
	lines = append(lines, "Commands per context:")
	for name, n := range ctxCounts {
		label := "commands"
		if n == 1 {
			label = "command"
		}
		lines = append(lines, fmt.Sprintf("  %s: %d %s", name, n, label))
	}

	if len(tagCounts) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Tags:")
		for tag, n := range tagCounts {
			lines = append(lines, fmt.Sprintf("  %s: %d", tag, n))
		}
	}

	if len(mostUsed) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Most used:")
		for _, c := range mostUsed {
			lines = append(lines, fmt.Sprintf("  [%s] %s (%s) — %d uses", c.ID, c.Title, c.Context, c.UsageCount))
		}
	}

	if len(recent) > 0 {
		lines = append(lines, "")
		lines = append(lines, "Recently used:")
		for _, c := range recent {
			ago := time.Since(c.LastUsedAt).Truncate(time.Minute)
			lines = append(lines, fmt.Sprintf("  [%s] %s (%s) — %s ago", c.ID, c.Title, c.Context, ago))
		}
	}

	writeResult(id, map[string]interface{}{
		"content": textContent(strings.Join(lines, "\n")),
	})
}
