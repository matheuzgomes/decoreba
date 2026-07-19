package core

import (
	"crypto/rand"
	"encoding/hex"
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

func GenID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
