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

func GenID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
