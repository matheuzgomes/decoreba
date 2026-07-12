package core

import (
	"crypto/rand"
	"encoding/hex"
	"time"
)

type Command struct {
	ID         string    `json:"id"`
	Context    string    `json:"context"`
	Title      string    `json:"title"`
	Command    string    `json:"command"`
	Tags       []string  `json:"tags,omitempty"`
	Notes      string    `json:"notes,omitempty"`
	UsageCount int       `json:"usage_count"`
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
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
