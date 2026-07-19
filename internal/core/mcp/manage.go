package mcp

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/matheuzgomes/decoreba/internal/core"
	"github.com/matheuzgomes/decoreba/internal/core/store"
)

type addParams struct {
	Context string   `json:"context"`
	Title   string   `json:"title"`
	Command string   `json:"command"`
	Tags    []string `json:"tags,omitempty"`
	Notes   string   `json:"notes,omitempty"`
	Confirm bool     `json:"confirm"`
}

func handleAdd(s *core.Store, id json.RawMessage, args json.RawMessage) {
	var p addParams
	if err := json.Unmarshal(args, &p); err != nil {
		writeError(id, -32602, "Invalid params: "+err.Error())
		return
	}
	if p.Context == "" || p.Title == "" || p.Command == "" {
		writeError(id, -32602, "context, title, and command are required")
		return
	}

	preview := fmt.Sprintf("Will add:\n  context: %s\n  title:   %s\n  command: %s",
		p.Context, p.Title, p.Command)
	if len(p.Tags) > 0 {
		preview += fmt.Sprintf("\n  tags:    %s", strings.Join(p.Tags, ", "))
	}
	if p.Notes != "" {
		preview += fmt.Sprintf("\n  notes:   %s", p.Notes)
	}

	if !p.Confirm {
		writeResult(id, map[string]interface{}{
			"content": textContent(preview + "\n\nPass \"confirm\": true to save."),
		})
		return
	}

	cmd := core.Command{
		ID:        core.GenID(),
		Context:   p.Context,
		Title:     p.Title,
		Command:   p.Command,
		Tags:      p.Tags,
		Notes:     p.Notes,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := BackupStore(); err != nil {
		writeError(id, -32603, "Backup failed: "+err.Error())
		return
	}

	s.Commands = append(s.Commands, cmd)
	if err := store.Save(s); err != nil {
		writeError(id, -32603, "Save failed: "+err.Error())
		return
	}

	writeResult(id, map[string]interface{}{
		"content": textContent(fmt.Sprintf("✓ Saved [%s] %s (%s)", cmd.ID, cmd.Title, cmd.Context)),
	})
}

type editParams struct {
	ID      string   `json:"id"`
	Context string   `json:"context,omitempty"`
	Title   string   `json:"title,omitempty"`
	Command string   `json:"command,omitempty"`
	Tags    []string `json:"tags,omitempty"`
	Notes   string   `json:"notes,omitempty"`
	Confirm bool     `json:"confirm"`
}

func handleEdit(s *core.Store, id json.RawMessage, args json.RawMessage) {
	var p editParams
	if err := json.Unmarshal(args, &p); err != nil {
		writeError(id, -32602, "Invalid params: "+err.Error())
		return
	}
	if p.ID == "" {
		writeError(id, -32602, "id is required")
		return
	}

	cmd, count := s.FindByPrefix(p.ID)
	if count == 0 {
		writeResult(id, map[string]interface{}{
			"content": textContent("No command found with that id."),
		})
		return
	}
	if count > 1 {
		writeResult(id, map[string]interface{}{
			"content": textContent(fmt.Sprintf("Ambiguous id %q (%d matches). Use more characters.", p.ID, count)),
		})
		return
	}

	var diffs []string
	if p.Context != "" && p.Context != cmd.Context {
		diffs = append(diffs, fmt.Sprintf("  context: %s → %s", cmd.Context, p.Context))
	}
	if p.Title != "" && p.Title != cmd.Title {
		diffs = append(diffs, fmt.Sprintf("  title:   %s → %s", cmd.Title, p.Title))
	}
	if p.Command != "" && p.Command != cmd.Command {
		diffs = append(diffs, fmt.Sprintf("  command: %s → %s", cmd.Command, p.Command))
	}
	if p.Notes != "" && p.Notes != cmd.Notes {
		diffs = append(diffs, fmt.Sprintf("  notes:   %s → %s", cmd.Notes, p.Notes))
	}
	if p.Tags != nil {
		old := strings.Join(cmd.Tags, ", ")
		new := strings.Join(p.Tags, ", ")
		if old != new {
			diffs = append(diffs, fmt.Sprintf("  tags:    [%s] → [%s]", old, new))
		}
	}

	if len(diffs) == 0 {
		writeResult(id, map[string]interface{}{
			"content": textContent("No changes detected."),
		})
		return
	}

	preview := "Changes:\n" + strings.Join(diffs, "\n")

	if !p.Confirm {
		writeResult(id, map[string]interface{}{
			"content": textContent(preview + "\n\nPass \"confirm\": true to save changes."),
		})
		return
	}

	if err := BackupStore(); err != nil {
		writeError(id, -32603, "Backup failed: "+err.Error())
		return
	}

	if p.Context != "" {
		cmd.Context = p.Context
	}
	if p.Title != "" {
		cmd.Title = p.Title
	}
	if p.Command != "" {
		cmd.Command = p.Command
	}
	if p.Notes != "" {
		cmd.Notes = p.Notes
	}
	if p.Tags != nil {
		cmd.Tags = p.Tags
	}
	cmd.UpdatedAt = time.Now()

	if err := store.Save(s); err != nil {
		writeError(id, -32603, "Save failed: "+err.Error())
		return
	}

	writeResult(id, map[string]interface{}{
		"content": textContent("✓ Changes saved.\n" + preview),
	})
}

type removeParams struct {
	ID      string `json:"id"`
	Confirm bool   `json:"confirm"`
}

func handleRemove(s *core.Store, id json.RawMessage, args json.RawMessage) {
	var p removeParams
	if err := json.Unmarshal(args, &p); err != nil {
		writeError(id, -32602, "Invalid params: "+err.Error())
		return
	}
	if p.ID == "" {
		writeError(id, -32602, "id is required")
		return
	}

	matchIdx := -1
	matchCount := 0
	for i, c := range s.Commands {
		if len(c.ID) >= len(p.ID) && c.ID[:len(p.ID)] == p.ID {
			matchIdx = i
			matchCount++
		}
	}

	if matchCount == 0 {
		writeResult(id, map[string]interface{}{
			"content": textContent("No command found with that id."),
		})
		return
	}
	if matchCount > 1 {
		writeResult(id, map[string]interface{}{
			"content": textContent(fmt.Sprintf("Ambiguous id %q (%d matches). Use more characters.", p.ID, matchCount)),
		})
		return
	}

	target := s.Commands[matchIdx]
	preview := fmt.Sprintf("Will delete:\n  [%s] %s (%s)\n  %s",
		target.ID, target.Title, target.Context, target.Command)
	if target.IsWorkflow() {
		preview += fmt.Sprintf("\n  (%d steps)", len(target.Steps))
	}

	if !p.Confirm {
		writeResult(id, map[string]interface{}{
			"content": textContent(preview + "\n\nPass \"confirm\": true to permanently delete."),
		})
		return
	}

	if err := BackupStore(); err != nil {
		writeError(id, -32603, "Backup failed: "+err.Error())
		return
	}

	s.Commands = append(s.Commands[:matchIdx], s.Commands[matchIdx+1:]...)
	if err := store.Save(s); err != nil {
		writeError(id, -32603, "Save failed: "+err.Error())
		return
	}

	writeResult(id, map[string]interface{}{
		"content": textContent(fmt.Sprintf("✓ Deleted [%s] %s (%s)", target.ID, target.Title, target.Context)),
	})
}
