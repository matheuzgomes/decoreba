package core

import (
	"crypto/rand"
	"encoding/hex"
	"strings"
	"time"
)

type Command struct {
	ID         string         `json:"id"`
	Context    string         `json:"context"`
	Title      string         `json:"title"`
	Command    string         `json:"command"`
	Tags       []string       `json:"tags,omitempty"`
	Notes      string         `json:"notes,omitempty"`
	Pinned     bool           `json:"pinned,omitempty"`
	Steps      []WorkflowStep `json:"steps,omitempty"`
	UsageCount int            `json:"usage_count"`
	LastUsedAt time.Time      `json:"last_used_at"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

// WorkflowStep is a single command within a workflow.
type WorkflowStep struct {
	Title   string `json:"title"`
	Command string `json:"command"`
}

// IsWorkflow returns true when the command has multiple steps.
func (c Command) IsWorkflow() bool {
	return len(c.Steps) > 0
}

type Store struct {
	Version  int       `json:"version"`
	Commands []Command `json:"commands"`
}

// Storer is the persistence seam for the command store.
type Storer interface {
	Load() (*Store, error)
	Save(*Store) error
}

// FindByPrefix returns the unique command whose ID starts with the given
// prefix, and the total number of matches.
func (s *Store) FindByPrefix(prefix string) (*Command, int) {
	var found *Command
	count := 0
	for i := range s.Commands {
		if strings.HasPrefix(s.Commands[i].ID, prefix) {
			found = &s.Commands[i]
			count++
		}
	}
	return found, count
}

// Replace replaces the command with a matching ID. Returns false if not found.
func (s *Store) Replace(cmd *Command) bool {
	for i := range s.Commands {
		if s.Commands[i].ID == cmd.ID {
			s.Commands[i] = *cmd
			return true
		}
	}
	return false
}

// RemoveByPrefix removes the uniquely-matching command by ID prefix.
// Returns the removed command and the match count (0=not found, 1=success, >1=ambiguous).
func (s *Store) RemoveByPrefix(prefix string) (*Command, int) {
	idx := -1
	count := 0
	for i, c := range s.Commands {
		if len(c.ID) >= len(prefix) && c.ID[:len(prefix)] == prefix {
			idx = i
			count++
		}
	}
	if count != 1 {
		return nil, count
	}
	removed := s.Commands[idx]
	s.Commands = append(s.Commands[:idx], s.Commands[idx+1:]...)
	return &removed, count
}

// BumpUsage increments the usage count and updates the last-used timestamp
// for the command with the given ID.
func (s *Store) BumpUsage(id string) {
	for i := range s.Commands {
		if s.Commands[i].ID == id {
			s.Commands[i].UsageCount++
			s.Commands[i].LastUsedAt = time.Now()
			return
		}
	}
}

// TogglePin flips the pinned state for the command with the given ID.
// Returns the new pinned state.
func (s *Store) TogglePin(id string) bool {
	for i := range s.Commands {
		if s.Commands[i].ID == id {
			s.Commands[i].Pinned = !s.Commands[i].Pinned
			return s.Commands[i].Pinned
		}
	}
	return false
}

// FilterByContext returns commands matching the given context
// (case-insensitive). Returns all commands when context is empty.
func (s *Store) FilterByContext(context string) []Command {
	if context == "" {
		return s.Commands
	}
	var filtered []Command
	for _, c := range s.Commands {
		if strings.EqualFold(c.Context, context) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// storeKey builds a dedup key from a command's context and command text.
func storeKey(c Command) string {
	return strings.ToLower(c.Context) + "\x00" + strings.ToLower(c.Command)
}

// Merge appends commands, skipping duplicates (same context+command,
// case-insensitive). Returns imported and skipped counts.
func (s *Store) Merge(incoming []Command) (imported, skipped int) {
	seen := make(map[string]bool, len(s.Commands))
	for _, c := range s.Commands {
		seen[storeKey(c)] = true
	}
	for _, c := range incoming {
		k := storeKey(c)
		if seen[k] {
			skipped++
			continue
		}
		s.Commands = append(s.Commands, c)
		seen[k] = true
		imported++
	}
	return
}

func GenID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
